package types

type Iterator[K any, V any] <-chan Item[K, V]

type Item[K any, V any] struct {
	Key   K
	Value V
}

type Marshaler interface {
	Encode(v any) error
}

type Unmarshaler struct {
	Decode func(v any) error
}
