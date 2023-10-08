package storage

import (
	"context"
	"lib/storage/encoding"
	"lib/utils/iter"
	"lib/utils/types"
)

var _ Bucketer = (*Bucket)(nil)

type Bucket struct {
	conn    Connection `cbor:"-"`
	buckets []string
}

func NewBucket(conn Connection, keys ...string) *Bucket {
	return &Bucket{
		conn:    conn,
		buckets: keys,
	}
}
func (b Bucket) EncodeValue(c encoding.ValueCoder) ([]byte, error) {
	return c.EncodeValue(b.buckets)
}

func (b *Bucket) DecodeValue(c encoding.ValueCoder, data []byte) error {
	return c.DecodeValue(data, &b.buckets)
}

func (b Bucket) Bucket(bucket ...string) *Bucket {
	b.buckets = append(b.buckets, bucket...)
	return &b
}

func (b Bucket) Prefix() string {
	return b.conn.Encoding().EncodeBucket(b.buckets...)
}

func (b Bucket) Encoding() encoding.Coder {
	return b.conn.Encoding()
}

func (b Bucket) PrintDebug() error {
	return b.conn.PrintDebug(b.Prefix())
}

func (b Bucket) Len() (int, error) {
	return b.conn.Len(b.Prefix())
}

func (b Bucket) Keys() ([]string, error) {
	keys, err := b.conn.Keys(b.Prefix())
	if err != nil {
		return nil, err
	}
	return iter.MapSlice(keys, func(key string) string {
		keys := b.conn.Encoding().DecodeKey(key)
		if len(keys) == 0 {
			return ""
		}
		return keys[len(keys)-1]
	}), nil
}

func (b Bucket) Values() ([][]byte, error) {
	return b.conn.Values(b.Prefix())
}

func (b Bucket) Set(k string, v any, op ...Option) error {
	return b.conn.Set(b.conn.Encoding().EncodeKey(b.Prefix(), k), v, op...)
}

func (b Bucket) Get(k string, v any) error {
	return b.conn.Get(b.conn.Encoding().EncodeKey(b.Prefix(), k), v)
}

func (b Bucket) Exists(k string) bool {
	return b.conn.Exists(b.conn.Encoding().EncodeKey(b.Prefix(), k))
}

func (b Bucket) Delete(k string) error {
	return b.conn.Delete(b.conn.Encoding().EncodeKey(b.Prefix(), k))
}

func (b Bucket) Iter(ctx context.Context, pfx string) types.Iterator[string, []byte] {
	out := make(chan types.Item[string, []byte])
	go func() {
		defer close(out)
		for item := range b.conn.Iter(ctx, b.conn.Encoding().EncodeKey(b.Prefix(), pfx)) {
			keys := b.conn.Encoding().DecodeKey(item.Key)
			if len(keys) == 0 {
				item.Key = ""
			} else {
				item.Key = keys[len(keys)-1]
			}

			out <- item
		}
	}()
	return out
}

func (b Bucket) Watch(ctx context.Context, k string) types.Watcher[string, []byte] {
	out := make(chan types.WatchMsg[string, []byte])
	go func() {
		defer close(out)
		for item := range b.conn.Watch(ctx, b.conn.Encoding().EncodeKey(b.Prefix(), k)) {
			keys := b.conn.Encoding().DecodeKey(item.Key)
			if len(keys) == 0 {
				item.Key = ""
			} else {
				item.Key = keys[len(keys)-1]
			}

			out <- item
		}
	}()
	return out
}

func (b Bucket) Tx(fn func(tx Transactioner) error) error {
	return b.conn.Tx(b.Prefix(), fn)
}
