package utils

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/barkimedes/go-deepcopy"
	jsonpatch "github.com/evanphx/json-patch"
)

func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func DeepCopy[T any](src T) *T {
	copy, err := deepcopy.Anything(&src)
	if err != nil {
		return nil
	}
	return copy.(*T)
}

// WaitWithTimeout waits for the waitgroup for the specified max timeout.
// Returns error if waiting timed out.
func WaitWithTimeout(wg *sync.WaitGroup, timeout time.Duration) error {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-c:
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

func All[T any](val []T, fn func(x T) bool) bool {
	for _, v := range val {
		if !fn(v) {
			return false
		}
	}

	return true
}

func Any[T any](val []T, fn func(x T) bool) bool {
	for _, v := range val {
		if fn(v) {
			return true
		}
	}

	return false
}

func Must[T any](out T, err error) T {
	if err != nil {
		panic(err)
	}
	return out
}

func First[T any](in []T) T {
	if len(in) > 0 {
		return in[0]
	}
	return *new(T)
}

func PanicOnNil(v any) {
	if v == nil {
		panic("value is nil")
	}
}

func Ternary[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}

func Patch(obj1, obj2 any) ([]byte, error) {
	objb1, err := json.Marshal(obj1)
	if err != nil {
		return nil, err
	}
	objb2, err := json.Marshal(obj2)
	if err != nil {
		return nil, err
	}

	return jsonpatch.CreateMergePatch(objb1, objb2)
}
