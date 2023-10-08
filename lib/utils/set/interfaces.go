package set

import "errors"

var ErrNilSet = errors.New("set is nil")

type void struct{}

type Seter[T comparable] interface {
	Safe() *Safe[T]

	Add(values ...T)
	Delete(values ...T)
	Exists(value T) bool
	Iter() <-chan T
	List() (list []T)
	Len() int

	MarshalJSON() ([]byte, error)
	UnmarshalJSON([]byte) error
	MarshalCBOR() ([]byte, error)
	UnmarshalCBOR([]byte) error
	MarshalText() ([]byte, error)
	UnmarshalText([]byte) error
}
