package ws

import (
	"context"
	"fmt"
	"lib/terrors"
	"lib/tlog"
	"lib/utils/math"
	"lib/utils/net"
	"lib/validator"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type ContextValue string

const (
	UpgradeVars ContextValue = "Upgrade-Vars" // url.Values
)

type Server struct {
	*ServerConfig
	handlers       map[string]*handler
	initialRequest http.Request
	userID         string
}

func NewServer(c ServerConfig) (*Server, error) {
	if err := c.validate(); err != nil {
		return nil, err
	}

	return &Server{
		ServerConfig: &c,
		handlers:     map[string]*handler{},
	}, nil
}

type websocketSession struct {
	// Session ctx
	Ctx context.Context

	// map of active subscriptions (key is request id)
	// Subscriptions maps.Maper[string, context.CancelFunc]

	CancelSubscription context.CancelFunc
}

// Handler returns a HTTP handler that will upgrade connection to websocket based one.
//
// Before upgrade, it runs AuthHandler. If it returns error, handler will return 401 Unauthorized to client.
// Otherwise handler will keep that connection and run socketHandler for that connection.
func (server *Server) Handler(w http.ResponseWriter, r *http.Request) {
	server.Context = context.WithValue(server.Context, UpgradeVars, r.URL.Query())
	conn, userID := baseHandler(w, r, server.ConnectHandler, server.Upgrader)
	if conn == nil {
		return
	}
	server.initialRequest = *r
	server.userID = userID
	server.socketHandler(conn)
}

func (server *Server) socketHandler(conn *websocket.Conn) {
	defer server.DisconnectHandler(&server.initialRequest)

	// session ctx
	ctx, cancel := context.WithCancel(server.Context)
	defer cancel()

	//-------------------
	// Print new connection info
	// go func() {
	userIP := net.RequestIP(&server.initialRequest)
	// service := net.DNSLookup(ip)
	// if service == "" {
	// service = ip
	// }
	tlog.Info("New connection " + userIP)
	// }()
	//-------------------

	webSess := &websocketSession{
		Ctx:                ctx,
		CancelSubscription: func() {},
	}

	// webSess := &websocketSession{
	// 	Ctx:           ctx,
	// 	Subscriptions: maps.NewSafe[string, context.CancelFunc](nil),
	// }

	// Response writer
	w := make(chan *ResponseS, 4)
	go func() {
		for {
			select {
			case <-ctx.Done():
				tlog.Debug("Closing writer")
				return

			case res := <-w:
				tlog.Debug("WS.Server: sending response:", res)
				writer, err := conn.NextWriter(websocket.TextMessage)
				if err != nil {
					tlog.Error("Error during message writing:", err)
					continue
				}
				err = server.Coder.Encode(writer, res)
				if err != nil {
					tlog.Error("Error during encoding message:", err)
				}
				writer.Close()
			}
		}
	}()

	// The event loop
	for {
		// read
		_, reader, err := conn.NextReader()
		if err != nil {
			if !websocket.IsCloseError(
				err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				tlog.Error("Error during message reading:", err)
			}
			defer conn.Close()
			tlog.Debug("Closing connection")
			return // will cancel ctx
		}

		// decode
		r := &Request{
			UserID: server.userID,
			UserIP: userIP,
		}
		if err = server.Coder.Decode(reader, r); err != nil {
			tlog.Error("Error during message decoding:", err)
			continue
		}
		tlog.Debug("WS.Server: received request:", r)
		if err = r.validate(); err != nil {
			tlog.Error("Validation error:", err)
			w <- Response(r, terrors.BadRequest, err)
			continue
		}

		// handle
		go func(r *Request) {
			requestCtx, cancelRequest := context.WithTimeout(
				ctx,
				time.Duration(math.Clamp(r.Timeout, 5, 60))*time.Second,
			)
			defer cancelRequest()

			r.Ctx = requestCtx
			server.requestHandler(r, w, webSess)
		}(r)
	}
}

// requestHandler handles every request sent from client.
func (server *Server) requestHandler(r *Request, w chan<- *ResponseS, sess *websocketSession) {
	// Get action handler
	handler, exists := server.handlers[r.Action]
	if !exists {
		prefix, _, _ := strings.Cut(r.Action, ".")
		handler, exists = server.handlers[prefix]
		if !exists {
			w <- Response(r, terrors.ActionNotFound, fmt.Errorf("unknown action %s", r.Action))
			return
		}
	}

	// Decode action data
	action, err := handler.decode(server.Coder, r.Args)
	if err != nil {
		w <- Response(r, terrors.BadRequest, err)
		return
	}

	if err = validator.Validate(&action); err != nil {
		w <- Response(r, terrors.BadRequest, err)
		return
	}

	tlog.Debug("WS.Server", r.Action, action)

	// Exec action handler
	result := make(chan *ResponseS)

	go func() {
		defer panicHandler(r, result)

		switch handler.t {
		case actionT:
			// action has this request context
			resCode, resData := action.(HandlerI).Handle(r)
			result <- Response(r, resCode, resData)

		case subscriptionT:
			sess.CancelSubscription()
			r.Ctx, sess.CancelSubscription = context.WithCancel(sess.Ctx)
			// cancel, ok := sess.Subscriptions.GetFull(r.RequestID)
			// if sufix == "cancel" {
			// 	if !ok {
			// 		result <- InvalidRequest(r, "Subscription not active")
			// 		return
			// 	}
			// 	cancel()
			// 	result <- Ok(r, "Subscription canceled")
			// 	return
			// }

			// save subscription
			// if ok {
			// 	result <- InvalidRequest(r, "Subscription already active")
			// 	return
			// }

			// r.Ctx, r.Cancel = context.WithCancel(sess.Ctx)
			// sess.Subscriptions.Set(r.RequestID, r.Cancel)

			// Start subscription writer
			go action.(SubHandlerI).HandleSub(r, w)
			// // Remove key from map when finished
			// go func() {
			// 	<-r.Ctx.Done()
			// 	sess.Subscriptions.Delete(r.RequestID)
			// }()
			result <- Response(r, terrors.Success, "Subscription updated")
		}
	}()

	// Wait for result with timeout
	select {
	case res := <-result:
		w <- res
	case <-r.Ctx.Done():
		w <- Response(r, terrors.Timeout, "Timeout")
	}
}
