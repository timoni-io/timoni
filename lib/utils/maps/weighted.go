package maps

import (
	"bytes"
	"context"
	"encoding/json"
	"lib/utils/types"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/exp/constraints"
)

var _ Maper[string, Weighted[string]] = (*WeightedMap[string, string])(nil)

type Weighted[V any] struct {
	Value  V
	Weight uint32
}

func (w Weighted[V]) Less(other any) bool {
	o, _ := other.(Weighted[V])
	return w.Weight > o.Weight
}

// WeightedMap is a map of Weighted values. Higher weight means higher priority (descending).
type WeightedMap[K constraints.Ordered, V any] struct {
	OrderedMap[K, Weighted[V]]
}

// NewWeighted returns new WeightedMap. Higher weight means higher priority (descending).
func NewWeighted[K constraints.Ordered, V any](data map[K]Weighted[V]) *WeightedMap[K, V] {
	if data == nil {
		data = map[K]Weighted[V]{}
	}
	return &WeightedMap[K, V]{
		OrderedMap: *NewOrderedByValue(data, nil),
	}
}

func NewWeightedMapFromSlice[K constraints.Ordered, V any](keys []K, data []V) *WeightedMap[K, V] {
	m := map[K]Weighted[V]{}
	for i, value := range data {
		m[keys[i]] = Weighted[V]{Value: value, Weight: uint32(i)}
	}

	return NewWeighted(m)
}

func (m *WeightedMap[K, V]) Eventful(ctx context.Context, buf int) *EventfulMap[K, Weighted[V]] {
	return NewEventful[K, Weighted[V]](m, ctx, buf)
}

func (m *WeightedMap[K, V]) ReadOnly() *ReadOnlyMap[K, Weighted[V]] {
	return NewReadOnlyMap[K, Weighted[V]](m)
}

func (m *WeightedMap[K, V]) Safe() *SafeMap[K, Weighted[V]] {
	return NewSafe[K, Weighted[V]](m)
}

func (m *WeightedMap[K, V]) marshal(enc types.Marshaler) error {
	if m == nil {
		return ErrNilMap
	}

	rawMap := map[K]V{}

	for k, v := range m.data {
		rawMap[k] = v.Value
	}

	return enc.Encode(rawMap)
}

func (m *WeightedMap[K, V]) unmarshal(dec types.Unmarshaler) error {
	if m == nil {
		return ErrNilMap
	}

	// unmarshal weightless values
	var rawMap map[K]V
	err := dec.Decode(&rawMap)
	if err != nil {
		return err
	}

	// assign existing weights
	finalMap := map[K]Weighted[V]{}
	for k, v := range rawMap {
		var weight uint32 = 0

		if x, exists := m.data[k]; exists {
			weight = x.Weight
		}

		finalMap[k] = Weighted[V]{Value: v, Weight: weight}
	}

	// save final map
	m.data = finalMap

	return nil
}

func (m *WeightedMap[K, V]) MarshalJSON() ([]byte, error) {
	data := &bytes.Buffer{}
	err := m.marshal(json.NewEncoder(data))
	return data.Bytes(), err
}

func (m *WeightedMap[K, V]) UnmarshalJSON(data []byte) error {
	return m.unmarshal(types.Unmarshaler{
		Decode: json.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

func (m *WeightedMap[K, V]) MarshalCBOR() ([]byte, error) {
	data := &bytes.Buffer{}
	err := m.marshal(cbor.NewEncoder(data))
	return data.Bytes(), err
}

func (m *WeightedMap[K, V]) UnmarshalCBOR(data []byte) error {
	return m.unmarshal(types.Unmarshaler{
		Decode: cbor.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

// Toml needs Map[K,V] to be in struct/slice/map
func (m *WeightedMap[K, V]) MarshalText() ([]byte, error) {
	data := &bytes.Buffer{}
	err := m.marshal(toml.NewEncoder(data))
	return data.Bytes(), err
}

// Toml needs Map[K,V] to be in struct/slice/map
func (m *WeightedMap[K, V]) UnmarshalText(data []byte) error {
	return m.unmarshal(types.Unmarshaler{
		Decode: func(v any) error {
			return toml.NewDecoder(bytes.NewReader(data)).Decode(v)
		},
	})
}

// Weighted helpers

func GetValue[K comparable, V any](data Maper[K, Weighted[V]], key K) V {
	return data.Get(key).Value
}

func SetValue[K comparable, V any](data Maper[K, Weighted[V]], key K, value V) {
	data.Set(key, Weighted[V]{Value: value, Weight: 0})
}

func WeightedIter[K comparable, V any](data Maper[K, Weighted[V]]) <-chan types.Iterator[K, V] {
	len := data.Len()
	weightChan := make(chan types.Iterator[K, V], len)

	go func() {
		lastWeight := -1
		var iter chan types.Item[K, V]

		for i := range data.Iter() {
			v := i.Value
			if lastWeight != int(v.Weight) {
				// Close previous
				if iter != nil {
					close(iter)
				}

				iter = make(chan types.Item[K, V], len)
				weightChan <- iter
				lastWeight = int(v.Weight)
			}

			iter <- types.Item[K, V]{Key: i.Key, Value: v.Value}
		}
		close(iter)
		close(weightChan)
	}()

	return weightChan
}
