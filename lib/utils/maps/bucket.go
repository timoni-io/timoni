package maps

import (
	"context"
	"lib/utils/types"
	"strings"
)

type Bucket[V any] struct {
	m   Maper[string, V]
	pfx string
}

func NewBucket[V any](m Maper[string, V], pfx string) *Bucket[V] {
	return &Bucket[V]{
		m:   m,
		pfx: pfx,
	}
}

func (b Bucket[V]) Bucket(pfx string) *Bucket[V] {
	return &Bucket[V]{
		m:   b.m,
		pfx: b.pfx + pfx,
	}
}

func (b Bucket[V]) Exists(k string) bool {
	return b.m.Exists(b.pfx + k)
}

func (b Bucket[V]) Get(k string) V {
	return b.m.Get(b.pfx + k)
}

func (b Bucket[V]) GetFull(k string) (V, bool) {
	return b.m.GetFull(b.pfx + k)
}

func (b Bucket[V]) Set(k string, v V) {
	b.m.Set(b.pfx+k, v)
}

func (b Bucket[V]) Delete(k string) {
	b.m.Delete(b.pfx + k)
}

func (b Bucket[V]) Iter() types.Iterator[string, V] {
	out := make(chan types.Item[string, V], b.m.Len())
	go func() {
		defer close(out)
		for item := range b.m.Iter() {
			if strings.HasPrefix(item.Key, b.pfx) {
				out <- types.Item[string, V]{
					Value: item.Value,
					Key:   item.Key,
				}
			}
		}
	}()

	return out
}

func (b Bucket[V]) Watch(ctx context.Context) types.Watcher[string, V] {
	eMap, ok := b.m.(*EventfulMap[string, V])
	if !ok {
		return nil
	}

	out := make(chan types.WatchMsg[string, V])
	go func() {
		for event := range eMap.Register(ctx) {
			if strings.HasPrefix(event.Key, b.pfx) {
				out <- types.WatchMsg[string, V]{
					Event: event.Event,
					Item: types.Item[string, V]{
						Key:   event.Key,
						Value: event.Value,
					},
				}
			}
		}
	}()

	return out
}

func (b Bucket[V]) ForEach(fn func(k string, v V) error) error {
	var err error
	for item := range b.Iter() {
		err = fn(item.Key, item.Value)
		if err != nil {
			return err
		}
	}
	return err
}

func (b Bucket[V]) Keys() (keys []string) {
	keys = make([]string, 0, b.m.Len())
	for item := range b.Iter() {
		keys = append(keys, item.Key)
	}
	return keys
}

func (b Bucket[V]) Values() (values []V) {
	values = make([]V, 0, b.m.Len())
	for item := range b.Iter() {
		values = append(values, item.Value)
	}
	return values
}

func (b Bucket[V]) Len() int {
	return len(b.Keys())
}
