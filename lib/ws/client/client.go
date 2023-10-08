package wsc

import (
	"errors"
	"io"
	"lib/utils/random"
	"lib/ws/coder"
	ws "lib/ws/server"
	"os"

	"github.com/gorilla/websocket"
)

var (
	ErrConfigNil = errors.New("wsc: Config is nil")
	ErrUrlEmpty  = errors.New("wsc: URL is empty")
)

type ClientConfig struct {
	URL              string
	DisableReconnect bool
	Dialer           *websocket.Dialer
	Stdout           io.Writer
	Coder            coder.Coder
}

func (c *ClientConfig) validate() error {
	if c == nil {
		return ErrConfigNil
	}
	if c.URL == "" {
		return ErrUrlEmpty
	}
	if c.Dialer == nil {
		c.Dialer = websocket.DefaultDialer
	}
	if c.Stdout == nil {
		c.Stdout = os.Stdout
	}
	if c.Coder == nil {
		c.Coder = coder.JSON{}
	}
	return nil
}

type Client struct {
	*ClientConfig
	*ChatClient
}

func NewClient(c ClientConfig) (*Client, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	chat, err := NewChatClient(c)
	if err != nil {
		return nil, err
	}

	return &Client{
		ClientConfig: &c,
		ChatClient:   chat,
	}, nil
}

func (c *Client) NextResponse() (*ws.ResponseS, error) {
	_, r, err := c.ws.NextReader()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err) {
			c.connected = false
		}
		return nil, err
	}
	msg := new(ws.ResponseS)
	if err = c.Coder.Decode(r, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Client) Send(action string, v any) (*ws.ResponseS, error) {
	c.reconnect()

	var raw coder.Raw
	if err := c.Coder.Encode(&raw, v); err != nil {
		return nil, err
	}
	if err := c.sendAction(action, raw); err != nil {
		return nil, err
	}

	res, err := c.NextResponse()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c *Client) sendAction(action string, raw coder.Raw) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()

	w, err := c.ws.NextWriter(websocket.TextMessage)
	if err != nil {
		c.connected = false
		return err
	}
	defer w.Close()

	if err = c.Coder.Encode(w, ws.Request{
		RequestID: random.ShortID(),
		Action:    action,
		Args:      raw,
	}); err != nil {
		return err
	}
	return nil
}
