package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
	"mmfm-playback-go/internal/logger"
	"mmfm-playback-go/pkg/types"
	"strings"
)

// MessageArgs represents arguments for a message
type MessageArgs struct {
	Command string        `json:"cmd"`
	Params  []interface{} `json:"args"`
}

// ToJSON converts MessageArgs to JSON string
func (ma *MessageArgs) ToJSON() (string, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(ma)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

const (
	EVENT_PLAYING      = "player.playing"
	EVENT_PAUSE        = "player.pause"
	EVENT_CURRENT      = "player.current"
	EVENT_CONTINUE     = "player.continue"
	EVENT_PLAY         = "player.play"
	EVENT_UPDATE       = "update"
	CHAT_EVENT_MESSAGE = "msg"
)

// PlayingEvent represents playing event data
type PlayingEvent struct {
	Song    *types.Song
	Current float64
}

// ParseMessageArgs parses a JSON string to MessageArgs
func ParseMessageArgs(source string) *MessageArgs {
	var params MessageArgs
	decoder := json.NewDecoder(strings.NewReader(source))
	err := decoder.Decode(&params)
	if err != nil {
		logger.Logger.Error(err)
	}

	return &params
}

// GetPlayingEvent extracts playing event from message args
func (ma *MessageArgs) GetPlayingEvent() (out *PlayingEvent, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Logger.Error(r)
			out = &PlayingEvent{}
			err = nil
		}
	}()

	if ma.Command != EVENT_PLAYING {
		return nil, errors.New("not match event")
	}

	item := ma.Params[0]
	data, ok := item.(map[string]interface{})

	if ok {
		song := &types.Song{
			Cover:    data["cover"].(string),
			URL:      data["url"].(string),
			Name:     data["name"].(string),
			Author:   data["author"].(string),
			Index:    ma.Params[1].(float64),
			Duration: ma.Params[3].(float64),
		}

		return &PlayingEvent{
			Song:    song,
			Current: ma.Params[2].(float64),
		}, nil
	}

	return &PlayingEvent{}, nil
}

// ChatClient handles chat communication
type ChatClient struct {
	url               string
	client            *gosocketio.Client
	listener          chan *MessageArgs
	connectedCallback func()
}

// NewChatClient creates a new ChatClient instance
func NewChatClient(url string) *ChatClient {
	return &ChatClient{
		url:      url,
		listener: make(chan *MessageArgs, 32),
	}
}

// Connect establishes connection to the chat server
func (cc *ChatClient) Connect() error {
	var err error
	cc.client, err = gosocketio.Dial(
		cc.url,
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		logger.Logger.Error(err)
		return err
	}

	return nil
}

// Listen starts listening for messages
func (cc *ChatClient) Listen() (chan *MessageArgs, error) {
	err := cc.Connect()
	if err != nil {
		return nil, err
	}

	err = cc.client.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		logger.Logger.Info("connected")
		if cc.connectedCallback != nil {
			cc.connectedCallback()
		}
	})

	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	err = cc.client.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		logger.Logger.Info("Disconnected")

		defer cc.client.Close()
		defer cc.Listen()
	})

	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	err = cc.client.On(CHAT_EVENT_MESSAGE, func(h *gosocketio.Channel, sourceParams string) {
		logger.Logger.Debug("--- Got chat message: ", sourceParams)

		cc.listener <- ParseMessageArgs(sourceParams)
	})

	if err != nil {
		logger.Logger.Error(err)
		return nil, err
	}

	return cc.listener, nil
}

// Close closes the chat connection
func (cc *ChatClient) Close(callbackList ...func()) {
	defer cc.client.Close()

	for _, callback := range callbackList {
		callback()
	}
}

// OnConnected sets a callback for when connected
func (cc *ChatClient) OnConnected(callback func()) {
	cc.connectedCallback = callback
}

// SendEvent sends an event to the chat server
func (cc *ChatClient) SendEvent(eventName string, params *MessageArgs) error {
	if cc.client == nil {
		return errors.New("client connection is not ready")
	}
	args, err := params.ToJSON()
	if err != nil {
		logger.Logger.Error(err)
		return err
	}
	return cc.client.Emit(eventName, args)
}
