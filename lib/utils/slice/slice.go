package slice

import (
	"bytes"
	"encoding/json"
	"errors"
	"lib/utils"
	"lib/utils/types"
	"sync"

	"github.com/fxamacker/cbor/v2"
	"github.com/pelletier/go-toml/v2"
)

var ErrNilSlice = errors.New("slice is nil")

type Slice[T any] struct {
	lock     utils.Locker
	data     []T
	capacity int
}

func NewSlice[T any](capacity int) *Slice[T] {
	return &Slice[T]{
		data:     make([]T, 0, capacity),
		capacity: capacity,
		lock:     &utils.FakeLock{},
	}
}

func (s *Slice[T]) Safe() *Slice[T] {
	s.lock = &sync.RWMutex{}
	return s
}

func (s *Slice[T]) Add(x ...T) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = append(s.data, x...)
}

func (s *Slice[T]) GetAll() []T {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.data
}

func (s *Slice[T]) Len() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.data)
}

func (s *Slice[T]) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = make([]T, 0, s.capacity)
}

func (s *Slice[T]) Get(idx int) *T {
	s.lock.Lock()
	defer s.lock.Unlock()

	if idx < 0 || idx >= len(s.data) {
		return nil
	}

	return &s.data[idx]
}

func (s *Slice[T]) Commit(fn func(data *[]T, capacity int)) {
	s.lock.Lock()
	defer s.lock.Unlock()
	fn(&s.data, s.capacity)
}

func (s *Slice[T]) Take() []T {
	s.lock.Lock()
	defer s.lock.Unlock()
	v := s.data
	s.data = make([]T, 0, s.capacity)
	return v
}

func (s *Slice[T]) marshal(enc types.Marshaler) error {
	if s == nil {
		return ErrNilSlice
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	return enc.Encode(s.data)
}

func (s *Slice[T]) unmarshal(dec types.Unmarshaler) error {
	if s == nil {
		return ErrNilSlice
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	var values []T
	err := dec.Decode(&values)
	if err != nil {
		return err
	}

	s.data = nil
	s.Add(values...)

	return nil
}

func (s *Slice[T]) MarshalJSON() ([]byte, error) {
	data := &bytes.Buffer{}
	err := s.marshal(json.NewEncoder(data))
	return data.Bytes(), err
}

func (s *Slice[T]) UnmarshalJSON(data []byte) error {
	return s.unmarshal(types.Unmarshaler{
		Decode: json.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

func (s *Slice[T]) MarshalCBOR() ([]byte, error) {
	data := &bytes.Buffer{}
	err := s.marshal(cbor.NewEncoder(data))
	return data.Bytes(), err
}

func (s *Slice[T]) UnmarshalCBOR(data []byte) error {
	return s.unmarshal(types.Unmarshaler{
		Decode: cbor.NewDecoder(bytes.NewReader(data)).Decode,
	})
}

// Toml needs Map[K,V] to be in struct/slice/map
func (s *Slice[T]) MarshalText() ([]byte, error) {
	data := &bytes.Buffer{}
	err := s.marshal(toml.NewEncoder(data))
	return data.Bytes(), err
}

// Toml needs Map[K,V] to be in struct/slice/map
func (s *Slice[T]) UnmarshalText(data []byte) error {
	return s.unmarshal(types.Unmarshaler{
		Decode: func(v any) error {
			return toml.NewDecoder(bytes.NewReader(data)).Decode(v)
		},
	})
}
