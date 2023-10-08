package set

import (
	"sync"
)

var _ Seter[string] = (*Safe[string])(nil)

type Safe[T comparable] struct {
	Seter[T]
	lock sync.RWMutex
}

func NewSafe[T comparable](set Seter[T]) *Safe[T] {
	if set == nil {
		set = New[T]()
	}
	return &Safe[T]{
		Seter: set,
	}
}

func (set *Safe[T]) init() error {
	if set == nil {
		return ErrNilSet
	}
	if set.Seter == nil {
		set.Seter = New[T]()
	}
	return nil
}

func (set *Safe[T]) Add(values ...T) {
	if err := set.init(); err != nil {
		return
	}
	set.lock.Lock()
	defer set.lock.Unlock()
	set.Seter.Add(values...)
}

func (set *Safe[T]) Delete(values ...T) {
	if err := set.init(); err != nil {
		return
	}
	set.lock.Lock()
	defer set.lock.Unlock()
	set.Seter.Delete(values...)
}

func (set *Safe[T]) Exists(value T) bool {
	if err := set.init(); err != nil {
		return false
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.Exists(value)
}

func (set *Safe[T]) Iter() <-chan T {
	if err := set.init(); err != nil {
		c := make(chan T)
		close(c)
		return c
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.Iter()
}

func (set *Safe[T]) List() (list []T) {
	if err := set.init(); err != nil {
		return nil
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.List()
}

func (set *Safe[T]) Len() int {
	if err := set.init(); err != nil {
		return -1
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.Len()
}

func (set *Safe[T]) MarshalJSON() ([]byte, error) {
	if err := set.init(); err != nil {
		return nil, err
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.MarshalJSON()
}

func (set *Safe[T]) UnmarshalJSON(data []byte) error {
	if err := set.init(); err != nil {
		return err
	}
	set.lock.Lock()
	defer set.lock.Unlock()
	return set.Seter.UnmarshalJSON(data)
}

func (set *Safe[T]) MarshalCBOR() ([]byte, error) {
	if err := set.init(); err != nil {
		return nil, err
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.MarshalCBOR()
}
func (set *Safe[T]) UnmarshalCBOR(data []byte) error {
	if err := set.init(); err != nil {
		return err
	}
	set.lock.Lock()
	defer set.lock.Unlock()
	return set.Seter.UnmarshalCBOR(data)
}

func (set *Safe[T]) MarshalText() ([]byte, error) {
	if err := set.init(); err != nil {
		return nil, err
	}
	set.lock.RLock()
	defer set.lock.RUnlock()
	return set.Seter.MarshalText()
}

func (set *Safe[T]) UnmarshalText(data []byte) error {
	if err := set.init(); err != nil {
		return err
	}
	set.lock.Lock()
	defer set.lock.Unlock()
	return set.Seter.UnmarshalText(data)
}
