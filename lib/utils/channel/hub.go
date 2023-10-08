package channel

import (
	"context"
)

type Client[T any] chan T

type Hub[T any] struct {
	ctx        context.Context
	clients    map[Client[T]]struct{}
	broadcast  chan T
	register   chan Client[T]
	unregister chan Client[T]
}

func NewHub[T any](ctx context.Context, buffer int) *Hub[T] {
	h := &Hub[T]{
		ctx:        ctx,
		clients:    make(map[Client[T]]struct{}),
		broadcast:  make(chan T, buffer),
		register:   make(chan Client[T]),
		unregister: make(chan Client[T]),
	}
	go h.run()
	return h
}

func (h *Hub[T]) run() {
	for {
		select {
		case <-h.ctx.Done():
			return
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unregister:
			delete(h.clients, client)
			close(client)
		case message := <-h.broadcast:
			for client := range h.clients {
				client <- message
			}
		}
	}
}

func (h *Hub[T]) Register(ctx context.Context) Client[T] {
	client := make(Client[T])
	h.register <- client

	go func() {
		<-ctx.Done()
		h.unregister <- client
	}()
	return client
}

func (h *Hub[T]) Broadcast(data T) {
	if len(h.clients) > 0 {
		h.broadcast <- data
	}
}
