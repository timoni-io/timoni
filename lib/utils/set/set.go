package set

import (
	"bytes"
	"encoding/json"
	"fmt"
	"lib/utils/types"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
)

var _ Seter[string] = (*Set[string])(nil)

// --- Set implemented using map ---
type Set[T comparable] struct {
	data map[T]void
}

func New[T comparable](data ...T) *Set[T] {
	set := &Set[T]{
		data: map[T]void{},
	}
	set.Add(data...)
	return set
}

func (set *Set[T]) Safe() *Safe[T] {
	return NewSafe[T](set)
}

func (set *Set[T]) Add(values ...T) {
	if set == nil {
		return
	}

	for _, value := range values {
		set.data[value] = void{}
	}
}

func (set *Set[T]) Delete(values ...T) {
	if set == nil {
		return
	}

	for _, v := range values {
		delete(set.data, v)
	}
}

func (set *Set[T]) Exists(value T) bool {
	if set == nil {
		return false
	}

	_, exists := set.data[value]
	return exists
}

func (set *Set[T]) Iter() <-chan T {
	out := make(chan T, len(set.data))

	go func() {
		for value := range set.data {
			out <- value
		}
	}()

	return out
}

func (set *Set[T]) List() (list []T) {
	if set == nil {
		return
	}

	i := 0
	list = make([]T, len(set.data))
	for value := range set.data {
		list[i] = value
		i++
	}
	return
}

func (set *Set[T]) Len() int {
	if set == nil {
		return 0
	}

	return len(set.data)
}

func (set *Set[T]) String() string {
	if set == nil {
		return "[]"
	}

	return fmt.Sprint(set.List())
}

func (set *Set[T]) marshal(enc types.Marshaler) error {
	if set == nil {
		return ErrNilSet
	}

	return enc.Encode(set.List())
}

func (set *Set[T]) unmarshal(dec types.Unmarshaler) error {
	if set == nil {
		return ErrNilSet
	}

	var values []T
	err := dec.Decode(&values)
	if err != nil {
		return err
	}

	set.data = make(map[T]void)
	set.Add(values...)

	return nil
}

func (set *Set[T]) MarshalJSON() ([]byte, error) {
	data := &bytes.Buffer{}
	err := set.marshal(json.NewEncoder(data))
	return data.Bytes(), err
}

func (set *Set[T]) UnmarshalJSON(data []byte) error {
	return set.unmarshal(types.Unmarshaler{
		Decode: json.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

func (set *Set[T]) MarshalCBOR() ([]byte, error) {
	data := &bytes.Buffer{}
	err := set.marshal(cbor.NewEncoder(data))
	return data.Bytes(), err
}

func (set *Set[T]) UnmarshalCBOR(data []byte) error {
	return set.unmarshal(types.Unmarshaler{
		Decode: cbor.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

// Toml needs Map[K,V] to be in struct/slice/map
func (set *Set[T]) MarshalText() ([]byte, error) {
	data := &bytes.Buffer{}
	err := set.marshal(toml.NewEncoder(data))
	return data.Bytes(), err
}

// Toml needs Map[K,V] to be in struct/slice/map
func (set *Set[T]) UnmarshalText(data []byte) error {
	return set.unmarshal(types.Unmarshaler{
		Decode: func(v any) error {
			return toml.NewDecoder(bytes.NewReader(data)).Decode(v)
		},
	})
}
