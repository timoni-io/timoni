package ws

import (
	"fmt"
	"lib/terrors"
	"lib/tlog"
	"lib/ws/coder"
	"log"
	"reflect"
)

type HandlerI interface {
	Handle(*Request) (code terrors.Error, data any)
}

// Subscription handler.
//
// Request have context which sends signal Done when subscription is canceled.
type SubHandlerI interface {
	HandleSub(*Request, chan<- *ResponseS)
}

func RegisterAction[T any](server *Server, name string) {
	a := (*T)(nil)

	if _, ok := any(a).(HandlerI); !ok {
		log.Fatal("invalid action type: %s %T", name, a)
	}

	if _, ok := server.handlers[name]; ok {
		log.Fatal("handler `%s` already registered", name)
	}

	server.handlers[name] = &handler{
		t:      actionT,
		action: reflect.TypeOf(a).Elem(),
	}
}

func RegisterSubscription[T any](server *Server, name string) {
	a := (*T)(nil)

	if _, ok := any(a).(SubHandlerI); !ok {
		log.Fatal("invalid subscription type: %s %T", name, a)
	}

	if _, ok := server.handlers[name]; ok {
		log.Fatal("handler `%s` already registered", name)
	}

	server.handlers[name] = &handler{
		t:      subscriptionT,
		action: reflect.TypeOf(a).Elem(),
	}
}

type handlerType byte

const (
	actionT handlerType = 1 << iota
	subscriptionT
)

type handler struct {
	t      handlerType
	action reflect.Type
}

// decode returns action interface
func (h *handler) decode(coder coder.Coder, data coder.Raw) (act any, err error) {
	// Create action struct pointer
	action := reflect.New(h.action).Interface()
	// Decode action
	if data.Len() > 0 {
		err = coder.Decode(data, action)
		if err != nil {
			return nil, fmt.Errorf("invalid request: %s", err)
		}
	}
	// Extract action from pointer
	return action, nil
}

func panicHandler(r *Request, w chan<- *ResponseS) {
	if err := recover(); err != nil {
		tlog.Error(err)
		w <- Response(r, terrors.InternalServerError, err)
	}
}
