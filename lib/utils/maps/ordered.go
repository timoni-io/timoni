package maps

import (
	"context"
	"lib/utils/iter"
	"lib/utils/types"
	"sort"

	"golang.org/x/exp/constraints"
)

var _ Maper[string, string] = (*OrderedMap[string, string])(nil)

type SortFunction[K comparable, V any] func(data map[K]V) []types.Item[K, V]

func defaultSortFunction[K constraints.Ordered, V any](data map[K]V) []types.Item[K, V] {
	items := make([]types.Item[K, V], 0, len(data))
	for k, v := range data {
		items = append(items, types.Item[K, V]{Key: k, Value: v})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Key < items[j].Key
	})
	return items
}

type OrderedMap[K constraints.Ordered, V any] struct {
	Map[K, V]
	sorter SortFunction[K, V]

	sorted bool
	cache  []types.Item[K, V]
}

func NewOrdered[K constraints.Ordered, V any](data map[K]V, sorter SortFunction[K, V]) *OrderedMap[K, V] {
	if sorter == nil {
		sorter = defaultSortFunction[K, V]
	}

	if data == nil {
		data = map[K]V{}
	}

	return &OrderedMap[K, V]{
		Map: Map[K, V]{
			data: data,
		},
		sorter: sorter,
	}
}

func NewOrderedFromMap[K constraints.Ordered, V any](m *Map[K, V], sorter SortFunction[K, V]) *OrderedMap[K, V] {
	return NewOrdered(m.data, sorter)
}

func NewOrderedByValue[K constraints.Ordered, V OrderedValue](data map[K]V, sorter SortFunction[K, V]) *OrderedMap[K, V] {
	if sorter == nil {
		// default value sorter
		sorter = func(data map[K]V) []types.Item[K, V] {
			items := make([]types.Item[K, V], 0, len(data))
			for k, v := range data {
				items = append(items, types.Item[K, V]{Key: k, Value: v})
			}
			sort.Slice(items, func(i, j int) bool {
				return items[i].Value.Less(items[j].Value)
			})
			return items
		}
	}

	if data == nil {
		data = map[K]V{}
	}

	return &OrderedMap[K, V]{
		Map: Map[K, V]{
			data: data,
		},
		sorter: sorter,
	}
}

func (m *OrderedMap[K, V]) init() error {
	if m == nil {
		return ErrNilMap
	}
	if m.sorter == nil {
		m.sorter = defaultSortFunction[K, V]
	}
	return nil
}

func (m *OrderedMap[K, V]) Eventful(ctx context.Context, buf int) *EventfulMap[K, V] {
	return NewEventful[K, V](m, ctx, buf)
}

func (m *OrderedMap[K, V]) ReadOnly() *ReadOnlyMap[K, V] {
	return NewReadOnlyMap[K, V](m)
}

func (m *OrderedMap[K, V]) Safe() *SafeMap[K, V] {
	return NewSafe[K, V](m)
}

func (m *OrderedMap[K, V]) sort() {
	if err := m.init(); err != nil {
		return
	}

	// skip sort if sorted
	if m.sorted {
		return
	}

	m.cache = m.sorter(m.data)
	m.sorted = true
}

// set value for key
func (m *OrderedMap[K, V]) Set(k K, v V) {
	if err := m.init(); err != nil {
		return
	}

	m.data[k] = v
	m.sorted = false
}

// delete key from Map
func (m *OrderedMap[K, V]) Delete(k K) {
	if err := m.init(); err != nil {
		return
	}

	delete(m.data, k)
	m.sorted = false
}

// run function with direct access to Map
func (m *OrderedMap[K, V]) Commit(fn func(data map[K]V)) {
	if err := m.init(); err != nil {
		return
	}

	fn(m.data)
	m.sorted = false
}

// return iterator for safe iterating over Map
func (m *OrderedMap[K, V]) Iter() types.Iterator[K, V] {
	if err := m.init(); err != nil {
		c := make(chan types.Item[K, V])
		close(c)
		return c
	}

	// sort before returning
	m.sort()

	iter := make(chan types.Item[K, V], len(m.cache))

	go func() {
		for _, v := range m.cache {
			iter <- v
		}
		close(iter)
	}()

	return iter
}

// range over Map
func (m *OrderedMap[K, V]) ForEach(fn func(k K, v V) error) error {
	if err := m.init(); err != nil {
		return err
	}

	// sort before returning
	m.sort()

	var err error
	for _, v := range m.cache {
		err = fn(v.Key, v.Value)
		if err != nil {
			return err
		}
	}
	return nil
}

// return all Map keys
func (m *OrderedMap[K, V]) Keys() (keys []K) {
	if err := m.init(); err != nil {
		return
	}

	// sort before returning
	m.sort()

	return iter.MapSlice(m.cache, func(i types.Item[K, V]) K { return i.Key })
}

// return all Map values
func (m *OrderedMap[K, V]) Values() (values []V) {
	if err := m.init(); err != nil {
		return
	}

	// sort before returning
	m.sort()

	return iter.MapSlice(m.cache, func(i types.Item[K, V]) V { return i.Value })
}

func (m *OrderedMap[K, V]) UnmarshalJSON(data []byte) error {
	return m.Map.UnmarshalJSON(data)
}

func (m *OrderedMap[K, V]) UnmarshalCBOR(data []byte) error {
	return m.Map.UnmarshalCBOR(data)
}

func (m *OrderedMap[K, V]) UnmarshalText(data []byte) error {
	return m.Map.UnmarshalText(data)
}
