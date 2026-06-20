package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"mmfm-playback-go/pkg/types"
	"strings"
	"time"

	socketio "github.com/zishang520/socket.io/clients/socket/v3"
	"github.com/zishang520/socket.io/clients/engine/v3/transports"
	siotypes "github.com/zishang520/socket.io/v3/pkg/types"
)

type MessageArgs struct {
	Command string        `json:"cmd"`
	Params  []interface{} `json:"args"`
}

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

type PlayingEvent struct {
	Song    *types.Song
	Current float64
}

func ParseMessageArgs(source string) *MessageArgs {
	var params MessageArgs
	decoder := json.NewDecoder(strings.NewReader(source))
	err := decoder.Decode(&params)
	if err != nil {
		slog.Error("parse message failed", "error", err, "raw", source)
	}

	return &params
}

func (ma *MessageArgs) GetPlayingEvent() (out *PlayingEvent, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in GetPlayingEvent", "error", fmt.Sprintf("%v", r))
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

type ChatClient struct {
	url               string
	client            *socketio.Socket
	listener          chan *MessageArgs
	connectedCallback func()
}

func NewChatClient(url string) *ChatClient {
	return &ChatClient{
		url:      url,
		listener: make(chan *MessageArgs, 32),
	}
}

func (cc *ChatClient) Connect() error {
	slog.Debug("connecting", "url", cc.url, "path", "/io")

	opts := socketio.DefaultOptions()
	opts.SetPath("/io")
	opts.SetTransports(siotypes.NewSet(transports.WebSocket))
	opts.SetReconnection(true)
	opts.SetReconnectionAttempts(999999)

	var err error
	cc.client, err = socketio.Connect(cc.url, opts)
	if err != nil {
		slog.Error("connect failed", "url", cc.url, "error", err)
		return err
	}

	slog.Info("connected", "url", cc.url)
	return nil
}

func (cc *ChatClient) Listen(stopCh <-chan struct{}) (chan *MessageArgs, error) {
	retryCounter := 0
	const maxRetries = 30
	const retryInterval = 2 * time.Second

connect:
	for {
		select {
		case <-stopCh:
			return nil, errors.New("connection cancelled")
		default:
		}

		err := cc.Connect()
		if err == nil {
			break
		}

		retryCounter++
		if retryCounter >= maxRetries {
			return nil, fmt.Errorf("connect failed after %d retries: %w", maxRetries, err)
		}

		slog.Error("connect failed, retrying",
			"attempt", retryCounter,
			"max_retries", maxRetries,
			"error", err,
		)

		select {
		case <-stopCh:
			return nil, errors.New("connection cancelled during retry")
		case <-time.After(retryInterval):
			goto connect
		}
	}

	_ = cc.client.On("connect", func(...any) {
		slog.Info("connected", "url", cc.url)
		if cc.connectedCallback != nil {
			cc.connectedCallback()
		}
	})

	_ = cc.client.On("connect_error", func(args ...any) {
		slog.Error("connect error", "args", fmt.Sprintf("%v", args))
	})

	_ = cc.client.On("disconnect", func(args ...any) {
		reason := ""
		if len(args) > 0 {
			reason = fmt.Sprintf("%v", args[0])
		}
		slog.Info("disconnected", "reason", reason)
	})

	_ = cc.client.On("reconnect", func(args ...any) {
		attempt := 0
		if len(args) > 0 {
			if a, ok := args[0].(int); ok {
				attempt = a
			}
		}
		slog.Info("reconnected", "attempt", attempt)
	})

	_ = cc.client.On("reconnect_attempt", func(args ...any) {
		attempt := 0
		if len(args) > 0 {
			if a, ok := args[0].(int); ok {
				attempt = a
			}
		}
		slog.Debug("reconnect attempt", "attempt", attempt)
	})

	_ = cc.client.On("reconnect_failed", func(...any) {
		slog.Error("reconnect failed")
	})

	_ = cc.client.On(CHAT_EVENT_MESSAGE, func(data ...any) {
		if len(data) > 0 {
			var sourceParams string
			switch v := data[0].(type) {
			case string:
				sourceParams = v
			default:
				b, _ := json.Marshal(v)
				sourceParams = string(b)
			}
			slog.Debug("received message", "raw", sourceParams)

			parsed := ParseMessageArgs(sourceParams)
			slog.Debug("parsed message", "command", parsed.Command, "param_count", len(parsed.Params))

			select {
			case cc.listener <- parsed:
			default:
				slog.Warn("listener channel full, message dropped", "command", parsed.Command)
			}
		}
	})

	return cc.listener, nil
}

func (cc *ChatClient) Close(callbackList ...func()) {
	slog.Info("closing chat client")
	defer cc.client.Disconnect()

	for _, callback := range callbackList {
		callback()
	}
}

func (cc *ChatClient) OnConnected(callback func()) {
	cc.connectedCallback = callback
}

func (cc *ChatClient) SendEvent(eventName string, params *MessageArgs) error {
	if cc.client == nil {
		return errors.New("client connection is not ready")
	}
	slog.Debug("sending event", "event", eventName, "command", params.Command)
	args, err := params.ToJSON()
	if err != nil {
		slog.Error("send event serialize failed", "event", eventName, "error", err)
		return err
	}
	err = cc.client.Emit(eventName, args)
	if err != nil {
		slog.Error("send event failed", "event", eventName, "error", err)
		return err
	}
	slog.Debug("event sent", "event", eventName)
	return nil
}
