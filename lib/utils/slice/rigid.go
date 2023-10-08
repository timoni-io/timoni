package slice

import (
	"lib/utils"
	"sync"

	"golang.org/x/exp/constraints"
)

type Rigid[T any, S constraints.Unsigned] struct {
	lock utils.Locker

	data []T
	size S
}

func NewRigid[T any, S constraints.Unsigned](size S) *Rigid[T, S] {
	if size == 0 {
		size = 1
	}

	return &Rigid[T, S]{
		data: make([]T, 0, size),
		size: size,
		lock: &utils.FakeLock{},
	}
}

func (r *Rigid[T, S]) Safe() *Rigid[T, S] {
	r.lock = &sync.RWMutex{}
	return r
}

func (r *Rigid[T, S]) Add(x ...T) (removed []T) {
	if r == nil {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	r.data = append(r.data, x...)
	if S(len(r.data)) > r.size {
		removed = r.data[:S(len(r.data))-r.size]
		r.data = r.data[S(len(r.data))-r.size:]
	}

	return removed
}

func (r *Rigid[T, S]) GetAll() []T {
	r.lock.RLock()
	defer r.lock.RUnlock()
	return r.data
}

func (r *Rigid[T, S]) Get(idx S) *T {
	if r == nil {
		return nil
	}

	r.lock.RLock()
	defer r.lock.RUnlock()

	if idx > S(len(r.data)) {
		idx = S(len(r.data) - 1)
	}
	return &r.data[idx]
}

func (r *Rigid[T, S]) GetLast(amount S) []T {
	if r == nil {
		return nil
	}

	r.lock.RLock()
	defer r.lock.RUnlock()

	if amount > r.size {
		amount = r.size
	}
	return r.data[S(len(r.data))-amount:]
}

func (r *Rigid[T, S]) Take() []T {
	if r == nil {
		return nil
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	v := r.data
	r.data = make([]T, 0, r.size)
	return v
}
func (s *Rigid[T, S]) Len() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.data)
}
