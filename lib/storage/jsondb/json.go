package jsondb

import (
	"context"
	"encoding/json"
	"fmt"
	"lib/storage"
	"lib/storage/encoding"
	"lib/storage/encoding/key"
	"lib/storage/encoding/value"
	"lib/storage/helpers"
	"lib/storage/options"
	"lib/utils"
	"lib/utils/iter"
	"lib/utils/maps"
	"lib/utils/types"
	"os"
	"sync"
	"time"

	log "lib/tlog"
)

var _ storage.Connection = (*JsonDB)(nil)

type JsonDB struct {
	path     string                             // path to db file
	data     maps.EventfulMaper[string, []byte] // database data
	ticker   *time.Ticker                       // ticker for db file sync
	encoding encoding.Coder                     // db key/value encoder

	// prefix mutex map
	pfxMutex maps.Maper[string, sync.Locker]

	// cancel for data map event watcher/hub
	cancel context.CancelFunc
}

func New(dataPath string) (storage.Connection, error) {
	var data map[string]any
	if utils.PathExists(dataPath) {
		// Load db
		file, err := os.Open(dataPath)
		if err != nil {
			return nil, err
		}

		err = json.NewDecoder(file).Decode(&data)
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	j := &JsonDB{
		data: maps.New(
			iter.MapMap(data, func(v any) []byte {
				b, _ := json.Marshal(v)
				return b
			}),
		).Eventful(ctx, 10),
		path:     dataPath,
		ticker:   time.NewTicker(time.Second),
		encoding: encoding.NewCoder(key.Simple, value.JSON),

		pfxMutex: maps.New[string, sync.Locker](nil).Safe(),
		cancel:   cancel,
	}

	go func() {
		for range j.ticker.C {
			err := j.Save()
			if err != nil {
				log.Error(err)
			}
		}
	}()

	return j, nil
}

func (j *JsonDB) Encoding() encoding.Coder {
	return j.encoding
}

func (j *JsonDB) PrintDebug(pfx string) error {
	var err error
	out := map[string]any{}

	maps.NewBucket[[]byte](j.data, pfx).ForEach(func(k string, v []byte) error {
		out[k], err = helpers.Decode[any](j.encoding, v)
		if err != nil {
			return err
		}
		return nil
	})

	log.PrintJSON(out)
	return err
}

func (j *JsonDB) Save() error {
	// unmarshal every value to map
	out := map[string]any{}
	j.data.ForEach(func(k string, v []byte) error {
		var val any
		err := json.Unmarshal(v, &val)
		out[k] = val
		return err
	})

	data, err := json.MarshalIndent(out, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(j.path, data, 0665)
	if err != nil {
		return err
	}
	return nil
}

func (j *JsonDB) Close() {
	j.cancel()
	j.ticker.Stop()
	err := j.Save()
	if err != nil {
		log.Error(err)
	}
}

func (j *JsonDB) Bucket(bucket ...string) *storage.Bucket {
	return storage.NewBucket(j, j.encoding.DecodeBucket(bucket...)...)
}

func (j *JsonDB) Set(k string, v any, op ...storage.Option) error {
	log.Debug("SET", k, v)
	data, err := j.encoding.EncodeValue(v)
	if err != nil {
		return err
	}
	j.data.Set(k, data)

	for _, opt := range op {
		switch opt := opt.(type) {
		case *options.TTLOption:
			go func(d time.Duration) {
				time.Sleep(d)
				j.data.Delete(k)
			}(opt.Value)

		default:
			log.Warning("Unsupported option: %T", opt)
		}
	}

	return nil
}

func (j *JsonDB) Get(k string, v any) error {
	log.Debug("GET", k)
	data, exists := j.data.GetFull(k)
	if !exists {
		return fmt.Errorf("get %s: %w", k, storage.ErrNotFound)
	}
	return j.encoding.DecodeValue(data, v)
}

func (j *JsonDB) Exists(k string) bool {
	log.Debug("EXISTS", k)
	return j.data.Exists(k)
}

func (j *JsonDB) Delete(k string) error {
	log.Debug("DELETE", k)
	j.data.Delete(k)
	return nil
}

func (j *JsonDB) Len(pfx string) (int, error) {
	log.Debug("LEN", pfx)
	return maps.NewBucket[[]byte](j.data, pfx).Len(), nil
}

func (j *JsonDB) Keys(pfx string) ([]string, error) {
	log.Debug("KEYS", pfx)
	return maps.NewBucket[[]byte](j.data, pfx).Keys(), nil
}

func (j *JsonDB) Values(pfx string) ([][]byte, error) {
	log.Debug("VALUES", pfx)
	return maps.NewBucket[[]byte](j.data, pfx).Values(), nil
}

func (j *JsonDB) Iter(ctx context.Context, pfx string) types.Iterator[string, []byte] {
	log.Debug("ITER", pfx)
	out := make(chan types.Item[string, []byte])
	go func() {
		defer close(out)
		for item := range maps.NewBucket[[]byte](j.data, pfx).Iter() {
			out <- types.Item[string, []byte]{
				Key:   item.Key,
				Value: item.Value,
			}
		}
	}()
	return out
}

func (j *JsonDB) Watch(ctx context.Context, pfx string) types.Watcher[string, []byte] {
	log.Debug("WATCH", pfx)
	out := make(chan types.WatchMsg[string, []byte])
	go func() {
		defer close(out)
		for event := range maps.NewBucket[[]byte](j.data, pfx).Watch(ctx) {
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

func (j *JsonDB) Tx(pfx string, fn func(tx storage.Transactioner) error) error {
	log.Debug("TX", pfx)

	var mtx sync.Locker
	j.pfxMutex.Commit(func(data map[string]sync.Locker) {
		var exists bool
		mtx, exists = data[pfx]
		if !exists {
			mtx = &sync.Mutex{}
			data[pfx] = mtx
		}
	})

	mtx.Lock()
	defer mtx.Unlock()
	return fn(j.Bucket(pfx))
}
