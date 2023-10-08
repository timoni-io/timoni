package set

import (
	"lib/utils/slice"

	"golang.org/x/exp/constraints"
)

var _ Seter[string] = (*Rigid[string, uint8])(nil)

type Rigid[T comparable, S constraints.Unsigned] struct {
	Set[T]
	Rigid slice.Rigid[T, S]
}

func NewRigid[T comparable, S constraints.Unsigned](size S) *Rigid[T, S] {
	return &Rigid[T, S]{
		Set:   *New[T](),
		Rigid: *slice.NewRigid[T](size),
	}
}

func (r *Rigid[T, S]) Safe() *Safe[T] {
	return NewSafe[T](r)
}

func (r *Rigid[T, S]) Add(x ...T) {
	toAdd := make([]T, 0, len(x))

	for _, v := range x {
		if !r.Set.Exists(v) {
			toAdd = append(toAdd, v)
		}
	}

	r.Set.Add(toAdd...)
	r.Set.Delete(r.Rigid.Add(toAdd...)...)
}
