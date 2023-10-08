package maps_test

import (
	"context"
	"lib/utils/maps"
	"testing"
	"time"
)

func TestEvents(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m := maps.New[string, string](nil).Eventful(ctx, 10)
	if m == nil {
		t.Fail()
		return
	}

	done := make(chan struct{})

	c := m.Register(ctx)

	go func() {
		<-c
		done <- struct{}{}
	}()

	m.Set("X", "x")

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fail()
	}
}
