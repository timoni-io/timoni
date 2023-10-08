package maps

import (
	"context"
	"fmt"
	"lib/utils/types"
	"sync"
)

var _ Maper[string, string] = (*SafeMap[string, string])(nil)

type SafeMap[K comparable, V any] struct {
	Maper[K, V]
	lock sync.RWMutex
}

func NewSafe[K comparable, V any](m Maper[K, V]) *SafeMap[K, V] {
	if m == nil {
		m = New[K, V](nil)
	}
	return &SafeMap[K, V]{
		Maper: m,
	}
}

func (m *SafeMap[K, V]) init() error {
	if m == nil {
		return ErrNilMap
	}
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	return nil
}

func (m *SafeMap[K, V]) Eventful(ctx context.Context, buf int) *EventfulMap[K, V] {
	if err := m.init(); err != nil {
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	return m.Maper.Eventful(ctx, buf)
}

// return ReadOnly Map
func (m *SafeMap[K, V]) ReadOnly() *ReadOnlyMap[K, V] {
	if err := m.init(); err != nil {
		return nil
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	return m.Maper.ReadOnly()
}

// return Safe Map
func (m *SafeMap[K, V]) Safe() *SafeMap[K, V] {
	return m
}

// return key existence
func (m *SafeMap[K, V]) Exists(k K) bool {
	if err := m.init(); err != nil {
		return false
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.Exists(k)
}

// return value for key
func (m *SafeMap[K, V]) Get(k K) V {
	if err := m.init(); err != nil {
		return *new(V)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.Get(k)
}

// return value and existence of key
func (m *SafeMap[K, V]) GetFull(k K) (obj V, exists bool) {
	if err := m.init(); err != nil {
		return *new(V), false
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.GetFull(k)
}

// set value for key
func (m *SafeMap[K, V]) Set(k K, v V) {
	if err := m.init(); err != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.Maper.Set(k, v)
}

// delete key from Map
func (m *SafeMap[K, V]) Delete(k K) {
	if err := m.init(); err != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.Maper.Delete(k)
}

// run function with direct access to Map
func (m *SafeMap[K, V]) Commit(fn func(data map[K]V)) {
	if err := m.init(); err != nil {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	m.Maper.Commit(fn)
}

// return iterator for safe iterating over Map
func (m *SafeMap[K, V]) Iter() types.Iterator[K, V] {
	if err := m.init(); err != nil {
		c := make(chan types.Item[K, V])
		close(c)
		return c
	}

	m.lock.RLock()
	iter := make(chan types.Item[K, V], m.Maper.Len())

	go func() {
		defer m.lock.RUnlock()
		for i := range m.Maper.Iter() {
			iter <- i
		}
		close(iter)
	}()

	return iter
}

// range over Map
func (m *SafeMap[K, V]) ForEach(fn func(k K, v V) error) error {
	if err := m.init(); err != nil {
		return err
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.ForEach(fn)
}

// return all Map keys
func (m *SafeMap[K, V]) Keys() (keys []K) {
	if err := m.init(); err != nil {
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.Keys()
}

// return all Map values
func (m *SafeMap[K, V]) Values() (values []V) {
	if err := m.init(); err != nil {
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.Values()
}

// return Map length
func (m *SafeMap[K, V]) Len() int {
	if err := m.init(); err != nil {
		return -1
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.Len()
}

func (m *SafeMap[K, V]) Copy() Maper[K, V] {
	if err := m.init(); err != nil {
		return nil
	}

	m.lock.RLock()
	defer m.lock.RUnlock()

	copy := m.Maper.Copy()
	if copy == nil {
		return nil
	}

	return copy
}

func (m *SafeMap[K, V]) MarshalJSON() ([]byte, error) {
	if err := m.init(); err != nil {
		return nil, err
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.MarshalJSON()
}

func (m *SafeMap[K, V]) UnmarshalJSON(data []byte) error {
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.Maper.UnmarshalJSON(data)
}

func (m *SafeMap[K, V]) MarshalCBOR() ([]byte, error) {
	if err := m.init(); err != nil {
		return nil, err
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.MarshalCBOR()
}

func (m *SafeMap[K, V]) UnmarshalCBOR(data []byte) error {
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.Maper.UnmarshalCBOR(data)
}

func (m *SafeMap[K, V]) MarshalText() ([]byte, error) {
	if err := m.init(); err != nil {
		return nil, err
	}
	m.lock.RLock()
	defer m.lock.RUnlock()
	return m.Maper.MarshalText()
}

func (m *SafeMap[K, V]) UnmarshalText(data []byte) error {
	if m.Maper == nil {
		m.Maper = New[K, V](nil)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.Maper.UnmarshalText(data)
}

func (m *SafeMap[K, V]) String() string {
	if err := m.init(); err != nil {
		return fmt.Sprintf("SafeMap{%s}", err.Error())
	}
	m.lock.RLock()
	defer m.lock.RUnlock()

	return fmt.Sprintf("SafeMap{%s}", m.Maper.String())
}
