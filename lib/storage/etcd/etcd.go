package etcd

import (
	"context"
	"errors"
	"fmt"
	"lib/storage"
	"lib/storage/encoding"
	"lib/storage/encoding/key"
	"lib/storage/encoding/value"
	"lib/storage/helpers"
	"lib/storage/options"
	"lib/utils/conv"
	"lib/utils/iter"
	"lib/utils/types"
	"time"

	log "lib/tlog"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type Etcd struct {
	// Client fields

	// context
	ctx    context.Context
	cancel context.CancelFunc

	endpoints []string
	cfg       clientv3.Config
	client    *clientv3.Client

	// Embedded server
	server *etcdEmbed

	// Storage driver encoding
	encoding encoding.Coder
}

func New(opts ...EtcdOpts) (storage.Connection, error) {
	etcd := &Etcd{encoding: encoding.NewCoder(key.Simple, value.JSON)}

	// Apply options
	for _, opt := range opts {
		err := opt(etcd)
		if err != nil {
			return nil, err
		}
	}

	if len(etcd.endpoints) == 0 {
		return nil, errors.New("endpoints not set. Use Embed or Endpoints option")
	}

	if etcd.ctx == nil {
		// Add default context
		Context(context.Background())(etcd)
	}

	etcd.cfg = clientv3.Config{
		Endpoints:   etcd.endpoints,
		DialTimeout: time.Second * 5,
		// Logger:      log.GetZapLogger(zapcore.ErrorLevel),
	}

	var err error
	etcd.client, err = clientv3.New(etcd.cfg)
	if err != nil {
		return nil, err
	}

	log.Debug(fmt.Sprintf("%+v", etcd.cfg))

	return etcd, nil
}

func (e *Etcd) Close() {
	e.cancel()
	e.client.Close()
	if e.server != nil {
		e.server.embed.Close()
	}
}

func (e *Etcd) applyOptions(ops []storage.Option) []clientv3.OpOption {
	out := []clientv3.OpOption{}

	for _, opt := range ops {
		switch opt := opt.(type) {

		case *options.TTLOption:
			lease, err := e.client.Lease.Grant(e.ctx, (int64)(opt.Value/time.Second))
			if err != nil {
				log.Error(err)
				continue
			}
			out = append(out, clientv3.WithLease(lease.ID))

		default:
			log.Warning("Unsupported option: %T", opt)
		}

	}

	return out
}

func (e *Etcd) Bucket(bucket ...string) *storage.Bucket {
	return storage.NewBucket(e, e.encoding.DecodeBucket(bucket...)...)
}

func (e *Etcd) Encoding() encoding.Coder {
	return e.encoding
}

func (e *Etcd) PrintDebug(pfx string) error {
	kv := e.client.KV
	resp, err := kv.Get(e.ctx, pfx, clientv3.WithPrefix())
	if err != nil {
		return fmt.Errorf("etcd: %w", err)
	}

	out := map[string]any{}
	for _, kv := range resp.Kvs {
		k := string(kv.Key)
		out[k], err = helpers.Decode[any](e.encoding, kv.Value)
		if err != nil {
			log.Warning("decode", "err", err, "key", k, "value", string(kv.Value))
			out[k] = string(kv.Value)
		}

		if m, ok := out[k].(map[interface{}]interface{}); ok {
			// fix type for json marshal
			out[k] = conv.FixMap(m)
		}

	}
	log.PrintJSON(out)
	return nil
}

func (e *Etcd) Set(k string, v any, op ...storage.Option) error {
	log.Debug("SET", k, v)
	kv := e.client.KV

	data, err := e.encoding.EncodeValue(v)
	if err != nil {
		return fmt.Errorf("encoder: %w", err)
	}

	_, err = kv.Put(e.ctx, k, string(data), e.applyOptions(op)...)
	return err
}

func (e *Etcd) Get(k string, v any) error {
	log.Debug("GET", k)
	kv := e.client.KV

	resp, err := kv.Get(e.ctx, k)
	if err != nil {
		return fmt.Errorf("etcd: %w", err)
	}
	if len(resp.Kvs) <= 0 {
		return fmt.Errorf("get %s: %w", k, storage.ErrNotFound)
	}

	return e.Encoding().DecodeValue(resp.Kvs[0].Value, v)
}

func (e *Etcd) Exists(k string) bool {
	log.Debug("EXISTS", k)
	kv := e.client.KV

	resp, err := kv.Get(e.ctx, k, clientv3.WithKeysOnly())
	if err != nil {
		return false
	}

	return resp.Count > 1
}

func (e *Etcd) Delete(k string) error {
	log.Debug("DELETE", k)
	kv := e.client.KV

	_, err := kv.Delete(e.ctx, k)
	return err
}

func (e *Etcd) Len(pfx string) (int, error) {
	log.Debug("LEN", pfx)
	kv := e.client.KV

	resp, err := kv.Get(e.ctx, pfx, clientv3.WithPrefix(), clientv3.WithCountOnly())
	if err != nil {
		return 0, fmt.Errorf("etcd: %w", err)

	}

	return int(resp.Count), nil
}

func (e *Etcd) Keys(pfx string) ([]string, error) {
	log.Debug("KEYS", pfx)
	kv := e.client.KV

	resp, err := kv.Get(e.ctx, pfx, clientv3.WithPrefix(), clientv3.WithKeysOnly())
	if err != nil {
		return nil, fmt.Errorf("etcd: %w", err)
	}

	return iter.MapSlice(resp.Kvs, func(kv *mvccpb.KeyValue) string {
		return string(kv.Key)
	}), nil
}

func (e *Etcd) Values(pfx string) ([][]byte, error) {
	log.Debug("VALUES", pfx)
	kv := e.client.KV

	resp, err := kv.Get(e.ctx, pfx, clientv3.WithPrefix())
	if err != nil {
		return nil, fmt.Errorf("etcd: %w", err)
	}

	return iter.MapSlice(resp.Kvs, func(kv *mvccpb.KeyValue) []byte {
		return kv.Value
	}), nil
}

func (e *Etcd) Iter(ctx context.Context, pfx string) types.Iterator[string, []byte] {
	log.Debug("ITER", pfx)
	out := make(chan types.Item[string, []byte])
	kv := e.client.KV

	go func() {
		defer close(out)

		resp, err := kv.Get(ctx, pfx, clientv3.WithPrefix())
		if err != nil {
			log.Error(err)
			return
		}

		for _, keyval := range resp.Kvs {
			out <- types.Item[string, []byte]{Key: string(keyval.Key), Value: keyval.Value}
		}
	}()

	return out
}

func (e *Etcd) Watch(ctx context.Context, pfx string) types.Watcher[string, []byte] {
	log.Debug("WATCH", pfx)

	out := make(chan types.WatchMsg[string, []byte])

	go func() {
		defer close(out)
		watcher := e.client.Watch(ctx, pfx, clientv3.WithPrefix())

		for resp := range watcher {
			for _, event := range resp.Events {
				out <- types.WatchMsg[string, []byte]{
					Event: types.EventType(event.Type.String()),
					Item: types.Item[string, []byte]{
						Key: string(event.Kv.Key), Value: event.Kv.Value,
					},
				}
			}
		}
	}()

	return out
}

func (e *Etcd) Tx(pfx string, fn func(tx storage.Transactioner) error) error {
	log.Debug("TX", pfx)

	sess, err := concurrency.NewSession(e.client)
	if err != nil {
		return fmt.Errorf("etcd: %w", err)
	}

	mutex := concurrency.NewLocker(sess, e.encoding.EncodeKey(pfx, e.encoding.Symbols().TransactionKey))

	mutex.Lock()
	defer mutex.Unlock()

	return fn(e.Bucket(pfx))
}
