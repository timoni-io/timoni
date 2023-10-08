package maps

type ReadOnlyMap[K comparable, V any] struct {
	Maper[K, V]
}

func NewReadOnlyMap[K comparable, V any](m Maper[K, V]) *ReadOnlyMap[K, V] {
	return &ReadOnlyMap[K, V]{
		Maper: m,
	}
}

func (ReadOnlyMap[K, V]) Set(k K, v V) {
}

func (ReadOnlyMap[K, V]) Delete(k K) {
}

func (ReadOnlyMap[K, V]) Commit(fn func(data map[K]V)) {
}

func (ReadOnlyMap[K, V]) UnmarshalJSON(data []byte) error {
	return ErrReadOnlyMap
}
func (ReadOnlyMap[K, V]) UnmarshalCBOR(data []byte) error {
	return ErrReadOnlyMap
}
func (ReadOnlyMap[K, V]) UnmarshalText(data []byte) error {
	return ErrReadOnlyMap
}
