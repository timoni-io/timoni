package ws

// var (
// 	ErrClientDisconnected = errors.New("client disconnected")
// )

// type ChatFunc func(hub *Hub)

// type Hub struct {
// 	Coder    coder.Coder
// 	ctx      context.Context
// 	receiver chan []byte
// 	sender   chan any
// }

// // Read waits for a next message, and unmarshals it into v. v has to be a pointer.
// // Read will read all messages sent from client, even if processing takes long
// // and client disconnects before every message was processed.
// // When there are no more messages, Read will return false.
// func (h *Hub) Read(v any) (more bool, err error) {
// 	msg, ok := <-h.receiver
// 	if !ok {
// 		return false, ErrClientDisconnected
// 	}
// 	return true, h.Coder.Decode(
// 		coder.NewRaw(msg),
// 		v,
// 	)
// }

// // Send sends a message or returns error when connection is closed
// func (h *Hub) Send(v any) error {
// 	select {
// 	case <-h.ctx.Done():
// 		log.Warning("Sending message on closed connection")
// 		return ErrClientDisconnected
// 	case h.sender <- v:
// 		return nil
// 	}
// }

// type ChatServer struct {
// 	*ServerConfig
// 	handler ChatFunc
// }

// func NewChatServer(c ServerConfig) (*ChatServer, error) {
// 	if err := c.validate(); err != nil {
// 		return nil, err
// 	}

// 	return &ChatServer{
// 		ServerConfig: &c,
// 	}, nil
// }

// func (server *ChatServer) Handler(chatHandler ChatFunc) http.HandlerFunc {
// 	if chatHandler == nil {
// 		panic("ws.ChatServer: chatHandler is nil")
// 	}
// 	server.handler = chatHandler

// 	return func(w http.ResponseWriter, r *http.Request) {
// 		conn, userID := baseHandler(w, r, server.ConnectHandler, server.Upgrader)
// 		if conn == nil {
// 			return
// 		}
// 		server.socketHandler(conn)
// 	}
// }

// func (server *ChatServer) socketHandler(conn *websocket.Conn) {
// 	// session ctx
// 	ctx, cancel := context.WithCancel(context.Background())

// 	wg := sync.WaitGroup{}
// 	r := make(chan []byte, 4) // buffer x messages
// 	w := make(chan any)       // there is no need to buffer this chan

// 	hub := &Hub{
// 		Coder:    server.Coder,
// 		ctx:      ctx,
// 		receiver: r,
// 		sender:   w,
// 	}

// 	defer func() {
// 		// close reader when all messages are sent to channel
// 		wg.Wait()
// 		close(r)
// 	}()

// 	defer cancel()

// 	go func() {
// 		// close writer when writing is done
// 		defer close(w)
// 		server.handler(hub)
// 	}()

// 	// Response writer
// 	go func() {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				log.Debug("Closing writer")
// 				return

// 			case res := <-hub.sender:
// 				// time.Sleep(5 * time.Millisecond) // front jest wolny
// 				writer, err := conn.NextWriter(websocket.TextMessage)
// 				if err != nil {
// 					log.Error("Error during message writing:", err)
// 					continue
// 				}
// 				err = server.Coder.Encode(writer, res)
// 				if err != nil {
// 					log.Error("Error during encoding message:", err)
// 				}
// 				writer.Close()
// 			}
// 		}
// 	}()

// 	// The event loop
// 	for {
// 		// read
// 		_, msg, err := conn.ReadMessage()
// 		if err != nil {
// 			if !websocket.IsCloseError(
// 				err,
// 				websocket.CloseNormalClosure,
// 				websocket.CloseGoingAway,
// 				websocket.CloseAbnormalClosure,
// 			) {
// 				log.Error("Error during message reading:", err)
// 			}
// 			defer conn.Close()
// 			log.Debug("Closing connection")
// 			return // will cancel ctx
// 		}

// 		// send to receiver
// 		wg.Add(1)
// 		go func(msg []byte) {
// 			defer wg.Done()
// 			hub.receiver <- msg
// 		}(msg)
// 	}
// }
