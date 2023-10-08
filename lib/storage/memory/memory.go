package memory

import (
	"context"
	"fmt"
	"lib/storage"
	"lib/storage/encoding"
	"lib/storage/encoding/key"
	"lib/storage/encoding/value"
	"lib/storage/helpers"
	"lib/storage/options"
	"lib/utils/maps"
	"lib/utils/types"
	"sync"
	"time"

	log "lib/tlog"
)

var _ storage.Connection = (*InMemory)(nil)

type InMemory struct {
	data     maps.EventfulMaper[string, []byte] // database data
	encoding encoding.Coder                     // db key/value encoder

	// prefix mutex map
	pfxMutex maps.Maper[string, sync.Locker]

	// cancel for data map event watcher/hub
	cancel context.CancelFunc
}

func New() (storage.Connection, error) {
	ctx, cancel := context.WithCancel(context.Background())

	m := &InMemory{
		data:     maps.New[string, []byte](nil).Eventful(ctx, 10),
		encoding: encoding.NewCoder(key.Simple, value.CBOR),

		pfxMutex: maps.New[string, sync.Locker](nil).Safe(),
		cancel:   cancel,
	}
	return m, nil
}

func (m *InMemory) Close() {
	m.cancel()
}

func (m *InMemory) Encoding() encoding.Coder {
	return m.encoding
}

func (m *InMemory) PrintDebug(pfx string) error {
	var err error
	out := map[string]any{}

	maps.NewBucket[[]byte](m.data, pfx).ForEach(func(k string, v []byte) error {
		out[k], err = helpers.Decode[any](m.encoding, v)
		if err != nil {
			return err
		}
		return nil
	})

	log.PrintJSON(out)
	return err
}

func (m *InMemory) Bucket(bucket ...string) *storage.Bucket {
	return storage.NewBucket(m, m.encoding.DecodeBucket(bucket...)...)
}

func (m *InMemory) Set(k string, v any, op ...storage.Option) error {
	log.Debug("SET", k, v)
	data, err := m.encoding.EncodeValue(v)
	if err != nil {
		return err
	}
	m.data.Set(k, data)

	for _, opt := range op {
		switch opt := opt.(type) {
		case *options.TTLOption:
			go func(d time.Duration) {
				time.Sleep(d)
				m.data.Delete(k)
			}(opt.Value)

		default:
			log.Warning("Unsupported option: %T", opt)
		}
	}
	return nil
}

func (m *InMemory) Get(k string, v any) error {
	log.Debug("GET", k)
	data, exists := m.data.GetFull(k)
	if !exists {
		return fmt.Errorf("get %s: %w", k, storage.ErrNotFound)
	}
	return m.encoding.DecodeValue(data, v)
}

func (m *InMemory) Exists(k string) bool {
	log.Debug("EXISTS", k)
	return m.data.Exists(k)
}

func (m *InMemory) Delete(k string) error {
	log.Debug("DELETE", k)
	m.data.Delete(k)
	return nil
}

func (m *InMemory) Len(pfx string) (int, error) {
	log.Debug("LEN", pfx)
	return maps.NewBucket[[]byte](m.data, pfx).Len(), nil
}

func (m *InMemory) Keys(pfx string) ([]string, error) {
	log.Debug("KEYS", pfx)
	return maps.NewBucket[[]byte](m.data, pfx).Keys(), nil
}

func (m *InMemory) Values(pfx string) ([][]byte, error) {
	log.Debug("VALUES", pfx)
	return maps.NewBucket[[]byte](m.data, pfx).Values(), nil
}

func (m *InMemory) Iter(ctx context.Context, pfx string) types.Iterator[string, []byte] {
	log.Debug("ITER", pfx)
	out := make(chan types.Item[string, []byte])
	go func() {
		defer close(out)
		for item := range maps.NewBucket[[]byte](m.data, pfx).Iter() {
			out <- types.Item[string, []byte]{
				Key:   item.Key,
				Value: item.Value,
			}
		}
	}()
	return out
}

func (m *InMemory) Watch(ctx context.Context, pfx string) types.Watcher[string, []byte] {
	log.Debug("WATCH", pfx)
	out := make(chan types.WatchMsg[string, []byte])
	go func() {
		defer close(out)
		for event := range maps.NewBucket[[]byte](m.data, pfx).Watch(ctx) {
			out <- types.WatchMsg[string, []byte]{
				Event: event.Event,
				Item: types.Item[string, []byte]{
					Key: event.Key, Value: event.Value,
				},
			}
		}
	}()
	return out
}

func (m *InMemory) Tx(pfx string, fn func(tx storage.Transactioner) error) error {
	log.Debug("TX", pfx)

	var mtx sync.Locker
	m.pfxMutex.Commit(func(data map[string]sync.Locker) {
		var exists bool
		mtx, exists = data[pfx]
		if !exists {
			mtx = &sync.Mutex{}
			data[pfx] = mtx
		}
	})

	mtx.Lock()
	defer mtx.Unlock()
	return fn(m.Bucket(pfx))
}
