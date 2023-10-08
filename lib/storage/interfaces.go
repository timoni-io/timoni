package storage

import (
	"context"
	"errors"
	"lib/storage/encoding"
	"lib/utils/types"
)

var (
	ErrNotFound = errors.New("obj not found")
)

type Option interface{}

type Setter interface {
	Set(k string, v any, op ...Option) error
}

type Getter interface {
	Exists(k string) bool
	Get(k string, v any) error
}

type Deleter interface {
	Delete(k string) error
}

type Iterator interface {
	// Returns Raw unmarshaled bytes. Recomended to use with helpers.NewIter[T]
	Iter(ctx context.Context, pfx string) types.Iterator[string, []byte]
}

type Watcher interface {
	// Returns Raw unmarshaled bytes. Recomended to use with helpers.NewWatcher[T]
	Watch(ctx context.Context, pfx string) types.Watcher[string, []byte]
}

type Transactioner interface {
	Getter
	Setter
	Deleter
	Iterator
	Encoding() encoding.Coder
}

type Bucketer interface {
	Getter
	Setter
	Deleter

	Watcher
	Iterator
	Transactioner

	Prefix() string
	Bucket(bucket ...string) *Bucket
	Encoding() encoding.Coder
	Tx(fn func(tx Transactioner) error) error

	Len() (int, error)
	Keys() ([]string, error)
	PrintDebug() error

	// Returns Raw unmarshaled bytes. Recomended to use with helpers.Values[T]
	Values() ([][]byte, error)
}

type Connection interface {
	Getter
	Setter
	Deleter

	Watcher
	Iterator
	Transactioner

	// Close connection
	Close()

	Bucket(bucket ...string) *Bucket
	Encoding() encoding.Coder
	Tx(pfx string, fn func(tx Transactioner) error) error

	Len(pfx string) (int, error)
	Keys(pfx string) ([]string, error)
	PrintDebug(pfx string) error

	// Returns Raw unmarshaled bytes. Recomended to use with helpers.Values[T]
	Values(pfx string) ([][]byte, error)
}
