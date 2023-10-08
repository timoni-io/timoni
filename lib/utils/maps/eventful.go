package maps

import (
	"context"
	"lib/utils/channel"
	"lib/utils/types"
)

var _ EventfulMaper[string, string] = (*EventfulMap[string, string])(nil)

type EventfulMap[K comparable, V any] struct {
	Maper[K, V]
	*channel.Hub[types.WatchMsg[K, V]]
}

func NewEventful[K comparable, V any](m Maper[K, V], ctx context.Context, buf int) *EventfulMap[K, V] {
	return &EventfulMap[K, V]{
		Maper: m.Safe(),
		Hub:   channel.NewHub[types.WatchMsg[K, V]](ctx, buf),
	}
}

func (m *EventfulMap[K, V]) Set(k K, v V) {
	m.Hub.Broadcast(types.WatchMsg[K, V]{
		Event: types.PutEvent,
		Item: types.Item[K, V]{
			Key:   k,
			Value: v,
		},
	})

	m.Maper.Set(k, v)
}

func (m *EventfulMap[K, V]) Delete(k K) {
	m.Hub.Broadcast(types.WatchMsg[K, V]{
		Event: types.DeleteEvent,
		Item: types.Item[K, V]{
			Key:   k,
			Value: m.Get(k),
		},
	})

	m.Maper.Delete(k)
}

func (m *EventfulMap[K, V]) UnmarshalJSON(data []byte) error {
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	return m.Maper.UnmarshalJSON(data)
}

func (m *EventfulMap[K, V]) UnmarshalCBOR(data []byte) error {
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	return m.Maper.UnmarshalCBOR(data)
}

func (m *EventfulMap[K, V]) UnmarshalText(data []byte) error {
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	return m.Maper.UnmarshalText(data)
}
