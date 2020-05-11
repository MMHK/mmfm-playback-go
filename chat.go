package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/graarh/golang-socketio"
	"strings"

	"github.com/graarh/golang-socketio/transport"
)

type MessageArgs struct {
	Command string        `json:"cmd"`
	Params  []interface{} `json:"args"`
}

func (this *MessageArgs) ToJSON() (string, error) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	err := encoder.Encode(this)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

const EVENT_PLAYING = "player.playing"
const EVENT_PAUSE = "player.pause"
const EVENT_CURRENT = "player.current"
const EVENT_CONTINUE = "player.continue"
const EVENT_PLAY = "player.play"
const EVENT_UPDATE = "update"
const CHAT_EVENT_MESSAGE = "msg"

type Song struct {
	Cover    string  `json:"cover"`
	URL      string  `json:"url"`
	Src      string  `json:"src"`
	Name     string  `json:"name"`
	Author   string  `json:"author"`
	Duration float64 `json:"duration"`
	Index    float64 `json:"index"`
}

func (this *Song) GetURL() string {
	url := this.URL
	if len(url) <= 0 {
		url = this.Src
	}

	return url
}

type PlayingEvent struct {
	Song    *Song
	Current float64
}

func ParseMessageArgs(source string) *MessageArgs {
	var params MessageArgs
	decoder := json.NewDecoder(strings.NewReader(source))
	err := decoder.Decode(&params)
	if err != nil {
		log.Error(err)
	}

	return &params
}

func (this *MessageArgs) GetPlayingEvent() (out *PlayingEvent, err error) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)

			out = &PlayingEvent{}
			err = nil
		}
	}()

	if this.Command != EVENT_PLAYING {
		return nil, errors.New("not match event")
	}

	item := this.Params[0]
	data, ok := item.(map[string]interface{})

	if ok {
		song := &Song{
			Cover:    data["cover"].(string),
			URL:      data["url"].(string),
			Name:     data["name"].(string),
			Author:   data["author"].(string),
			Index:    this.Params[1].(float64),
			Duration: this.Params[3].(float64),
		}

		return &PlayingEvent{
			Song:    song,
			Current: this.Params[2].(float64),
		}, nil
	}

	return &PlayingEvent{}, nil
}

type ChatClient struct {
	url               string
	client            *gosocketio.Client
	listener          chan *MessageArgs
	connectedCallback func()
}

func NewChatClient(url string) *ChatClient {
	return &ChatClient{
		url:      url,
		listener: make(chan *MessageArgs, 32),
	}
}

func (this *ChatClient) Connect() error {
	var err error
	this.client, err = gosocketio.Dial(
		this.url,
		transport.GetDefaultWebsocketTransport())
	if err != nil {
		log.Error(err)
		return err
	}

	return nil
}

func (this *ChatClient) Listen() (chan *MessageArgs, error) {
	err := this.Connect()
	if err != nil {
		return nil, err
	}

	err = this.client.On(gosocketio.OnConnection, func(h *gosocketio.Channel) {
		log.Info("connected")
		if this.connectedCallback != nil {
			this.connectedCallback()
		}
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = this.client.On(gosocketio.OnDisconnection, func(h *gosocketio.Channel) {
		log.Info("Disconnected")

		defer this.client.Close()
		defer this.Listen()
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = this.client.On(CHAT_EVENT_MESSAGE, func(h *gosocketio.Channel, sourceParams string) {
		log.Debug("--- Got chat message: ", sourceParams)

		this.listener <- ParseMessageArgs(sourceParams)
	})

	if err != nil {
		log.Error(err)
		return nil, err
	}

	return this.listener, nil
}

func (this *ChatClient) Close(callbackList ...func()) {
	defer this.client.Close()

	for _, callback := range callbackList {
		callback()
	}
}

func (this *ChatClient) OnConnected(callback func()) {
	this.connectedCallback = callback
}

func (this *ChatClient) SendEvent(eventName string, params *MessageArgs) error {
	if this.client == nil {
		return errors.New("client connection is not ready")
	}
	args, err := params.ToJSON()
	if err != nil {
		log.Error(err)
		return err
	}
	return this.client.Emit(eventName, args)
}
