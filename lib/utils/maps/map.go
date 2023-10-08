package maps

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"lib/utils"
	"lib/utils/types"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
)

var _ Maper[string, string] = (*Map[string, string])(nil)

type Map[K comparable, V any] struct {
	data map[K]V
}

func New[K comparable, V any](data map[K]V) *Map[K, V] {
	if data == nil {
		data = map[K]V{}
	}
	return &Map[K, V]{
		data: data,
	}
}

func (m *Map[K, V]) init() error {
	if m == nil {
		return ErrNilMap
	}
	if m.data == nil {
		m.data = map[K]V{}
	}
	return nil
}

func (m *Map[K, V]) Eventful(ctx context.Context, buf int) *EventfulMap[K, V] {
	return NewEventful[K, V](m, ctx, buf)
}

// return ReadOnly Map
func (m *Map[K, V]) ReadOnly() *ReadOnlyMap[K, V] {
	return NewReadOnlyMap[K, V](m)
}

// return Safe Map
func (m *Map[K, V]) Safe() *SafeMap[K, V] {
	return NewSafe[K, V](m)
}

// return key existence
func (m *Map[K, V]) Exists(k K) bool {
	if err := m.init(); err != nil {
		return false
	}

	_, exists := m.data[k]
	return exists
}

// return value for key
func (m *Map[K, V]) Get(k K) V {
	if err := m.init(); err != nil {
		return *new(V)
	}
	return m.data[k]
}

// return value and existence of key
func (m *Map[K, V]) GetFull(k K) (obj V, exists bool) {
	if err := m.init(); err != nil {
		return
	}

	obj, exists = m.data[k]
	return
}

// set value for key
func (m *Map[K, V]) Set(k K, v V) {
	if err := m.init(); err != nil {
		return
	}

	m.data[k] = v
}

// delete key from Map
func (m *Map[K, V]) Delete(k K) {
	if err := m.init(); err != nil {
		return
	}

	delete(m.data, k)
}

// run function with direct access to Map
func (m *Map[K, V]) Commit(fn func(data map[K]V)) {
	if err := m.init(); err != nil {
		return
	}

	fn(m.data)
}

// return iterator for safe iterating over Map
func (m *Map[K, V]) Iter() types.Iterator[K, V] {
	if err := m.init(); err != nil {
		c := make(chan types.Item[K, V])
		close(c)
		return c
	}

	iter := make(chan types.Item[K, V], len(m.data))

	go func() {
		for k, v := range m.data {
			iter <- types.Item[K, V]{Key: k, Value: v}
		}
		close(iter)
	}()

	return iter
}

// range over Map
func (m *Map[K, V]) ForEach(fn func(k K, v V) error) error {
	if err := m.init(); err != nil {
		return err
	}

	var err error
	for k, v := range m.data {
		err = fn(k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

// return all Map keys
func (m *Map[K, V]) Keys() (keys []K) {
	if err := m.init(); err != nil {
		return
	}

	i := 0
	keys = make([]K, len(m.data))

	for k := range m.data {
		keys[i] = k
		i++
	}

	return
}

// return all Map values
func (m *Map[K, V]) Values() (values []V) {
	if err := m.init(); err != nil {
		return
	}

	i := 0
	values = make([]V, len(m.data))

	for _, v := range m.data {
		values[i] = v
		i++
	}

	return
}

// return Map length
func (m *Map[K, V]) Len() int {
	return len(m.data)
}

// return Map copy
func (m *Map[K, V]) Copy() Maper[K, V] {
	if err := m.init(); err != nil {
		return nil
	}

	copy := utils.DeepCopy(m.data)
	if copy == nil {
		return nil
	}
	return New(*copy)
}

// returns original map. Use copy before calling this method.
func (m *Map[K, V]) Raw() map[K]V {
	if err := m.init(); err != nil {
		return nil
	}
	return m.data
}

func (m *Map[K, V]) marshal(enc types.Marshaler) error {
	if err := m.init(); err != nil {
		return err
	}
	return enc.Encode(m.data)
}

func (m *Map[K, V]) unmarshal(dec types.Unmarshaler) error {
	if err := m.init(); err != nil {
		return nil
	}
	return dec.Decode(&m.data)
}

func (m *Map[K, V]) MarshalJSON() ([]byte, error) {
	data := &bytes.Buffer{}
	err := m.marshal(json.NewEncoder(data))
	return data.Bytes(), err
}

func (m *Map[K, V]) UnmarshalJSON(data []byte) error {
	return m.unmarshal(types.Unmarshaler{
		Decode: json.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

func (m *Map[K, V]) MarshalCBOR() ([]byte, error) {
	data := &bytes.Buffer{}
	err := m.marshal(cbor.NewEncoder(data))
	return data.Bytes(), err
}

func (m *Map[K, V]) UnmarshalCBOR(data []byte) error {
	return m.unmarshal(types.Unmarshaler{
		Decode: cbor.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

// Toml needs Map[K,V] to be in struct/slice/map
func (m *Map[K, V]) MarshalText() ([]byte, error) {
	data := &bytes.Buffer{}
	err := m.marshal(toml.NewEncoder(data))
	return data.Bytes(), err
}

// Toml needs Map[K,V] to be in struct/slice/map
func (m *Map[K, V]) UnmarshalText(data []byte) error {
	return m.unmarshal(types.Unmarshaler{
		Decode: func(v any) error {
			return toml.NewDecoder(bytes.NewReader(data)).Decode(v)
		},
	})
}

func (m *Map[K, V]) String() string {
	return fmt.Sprintf("%v", m.data)
}
