package ws

import (
	"context"
	"fmt"
	"lib/terrors"
	"lib/tlog"
	"lib/ws/coder"
	"net/http"

	"github.com/gorilla/websocket"
)

type ServerConfig struct {
	Upgrader          *websocket.Upgrader
	ConnectHandler    func(r *http.Request) (string, terrors.Error)
	DisconnectHandler func(r *http.Request)
	Coder             coder.Coder
	Context           context.Context
}

func baseHandler(
	w http.ResponseWriter,
	r *http.Request,
	connectHandler func(r *http.Request) (string, terrors.Error),
	upgrader *websocket.Upgrader,
) (*websocket.Conn, string) {

	userID, errorCode := connectHandler(r)
	if errorCode != 0 {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("ERROR " + fmt.Sprint(errorCode)))
		return nil, ""
	}

	// Upgrade raw HTTP connection to a websocket based one
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		tlog.Error("Error during connection upgrade:", err)
		return nil, ""
	}

	return conn, userID
}

func (c *ServerConfig) validate() error {

	if c.Upgrader == nil {
		c.Upgrader = &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
	}

	if c.ConnectHandler == nil {
		c.ConnectHandler = func(r *http.Request) (string, terrors.Error) { return "", 0 }
	}

	if c.DisconnectHandler == nil {
		c.DisconnectHandler = func(r *http.Request) {}
	}

	if c.Coder == nil {
		c.Coder = coder.JSON{}
	}

	if c.Context == nil {
		c.Context = context.Background()
	}

	return nil
}
