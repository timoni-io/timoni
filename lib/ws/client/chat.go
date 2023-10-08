package wsc

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type ChatClient struct {
	*ClientConfig
	ws        *websocket.Conn
	connected bool
	writeLock sync.Mutex
}

// NewChatClient creates new client, connects to the ws server and tries to reconnect when disconnected.
func NewChatClient(c ClientConfig) (*ChatClient, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	client := &ChatClient{
		ClientConfig: &c,
	}

	fmt.Fprintln(c.Stdout, "ws.Client: Connecting to", c.URL)
	err = client.initConn()
	if err != nil {
		fmt.Fprintf(c.Stdout, "ws.Client: Error connecting to %s: %s\n", c.URL, err.Error())
		if client.DisableReconnect {
			return nil, err
		}
		client.reconnect()
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM)
	// register interupt handler
	go func(ch chan os.Signal) {
		<-ch
		if client.ws != nil {
			client.ws.WriteMessage(
				websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, "papa"),
			)
			client.ws.Close()
		}
		os.Exit(1)
	}(interrupt)

	return client, nil
}

// initConn initializes the connection.
func (c *ChatClient) initConn() error {
	conn, _, err := c.Dialer.Dial(c.URL, nil)
	if err != nil {
		return err
	}

	c.connected = true
	c.ws = conn

	c.ws.SetCloseHandler(func(code int, text string) error {
		// mark connection as disconnected
		c.connected = false
		return nil
	})

	return nil
}

func (c *ChatClient) IsConnected() bool {
	return c.connected
}

// RegisterReceiver registers a handler for every incoming ws message.
func RegisterReceiver[T any](c *ChatClient, fn func(msg T, err error)) {
	go func() {
		for {
			// reconnect
			if !c.connected && c.DisableReconnect {
				return
			}
			c.reconnect()

			msg := new(T)
			_, r, err := c.ws.NextReader()
			if err != nil && websocket.IsUnexpectedCloseError(err) {
				c.connected = false
				go fn(*msg, err)
				continue
			}
			err = c.Coder.Decode(r, msg)
			go fn(*msg, err)
		}
	}()
}

func (c *ChatClient) Send(v any) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	c.reconnect()

	w, err := c.ws.NextWriter(websocket.TextMessage)
	if err != nil {
		c.connected = false
		return err
	}
	defer w.Close()

	err = c.Coder.Encode(w, v)
	if err != nil {
		return err
	}
	return nil
}

func (c *ChatClient) SendBytes(data []byte) error {
	c.writeLock.Lock()
	defer c.writeLock.Unlock()
	c.reconnect()

	err := c.ws.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		c.connected = false
		return err
	}
	return nil
}

func (c *ChatClient) reconnect() {
	for !c.connected {
		// retry failed connection
		// fmt.Fprintln(c.Stdout, "ws.Client: Reconnecting", c.URL)
		c.initConn()
		time.Sleep(500 * time.Millisecond)
	}
}
