package ws

import (
	"context"
	"encoding/json"
	"errors"
	"lib/ws/coder"
)

type Request struct {
	RequestID string
	Action    string
	Args      coder.Raw
	Timeout   uint // in seconds

	// action - request ctx, with timeout
	//
	// subscription - session ctx, canceled when subscription is changed or disconnected
	Ctx    context.Context    `json:"-"`
	Cancel context.CancelFunc `json:"-"`
	UserID string             `json:"-"`
	UserIP string             `json:"-"`
}

var (
	ErrMissingRequestID = errors.New("invalid request - missing `RequestID`")
	ErrMissingAction    = errors.New("invalid request - missing `Action`")
)

func (r *Request) validate() error {
	if r.RequestID == "" {
		return ErrMissingRequestID
	}
	if r.Action == "" {
		return ErrMissingAction
	}
	return nil
}

func (r Request) String() string {
	buf, _ := json.MarshalIndent(r, "", "    ")
	return string(buf)
}
