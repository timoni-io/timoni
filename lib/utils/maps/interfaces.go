package maps

import (
	"context"
	"errors"
	"lib/utils/channel"
	"lib/utils/types"
)

var (
	ErrNilMap      = errors.New("map is nil")
	ErrReadOnlyMap = errors.New("map is readonly")
)

// All maps must implement this interface themselves.
type MapConverter[K comparable, V any] interface {
	// Eventful map forces safe
	Eventful(ctx context.Context, buf int) *EventfulMap[K, V]

	// return SafeMap
	Safe() *SafeMap[K, V]

	// return ReadOnly Map
	ReadOnly() *ReadOnlyMap[K, V]
}

type Maper[K comparable, V any] interface {
	MapConverter[K, V]

	// return key existence
	Exists(k K) bool

	// return value for key
	Get(k K) V

	// return value and existence of key
	GetFull(k K) (obj V, exists bool)

	// set value for key
	Set(k K, v V)

	// delete key from Map
	Delete(k K)

	// run function with direct access to Map
	Commit(fn func(data map[K]V))

	// return iterator for safe iterating over Map
	Iter() types.Iterator[K, V]

	// range over Map
	ForEach(fn func(k K, v V) error) error

	// return all Map keys
	Keys() (keys []K)

	// return all Map values
	Values() (values []V)

	// return Map length
	Len() int

	// return underlying Map copy (writable, non-safe or eventful)
	Copy() Maper[K, V]

	// returns original map. Use copy before calling this method.
	Raw() map[K]V

	MarshalJSON() ([]byte, error)
	UnmarshalJSON(data []byte) error

	MarshalCBOR() ([]byte, error)
	UnmarshalCBOR(data []byte) error

	// Toml needs Map[K,V] to be in struct/slice/map

	MarshalText() ([]byte, error)
	UnmarshalText(data []byte) error

	String() string
}

type EventfulMaper[K comparable, V any] interface {
	Maper[K, V]

	Register(context.Context) channel.Client[types.WatchMsg[K, V]]
}

type OrderedValue interface {
	Less(x any) bool
}
