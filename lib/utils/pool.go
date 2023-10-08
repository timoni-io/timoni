package utils

type Pool[T any] struct {
	ch chan T
}

func NewPool[T any](size int, objs ...T) *Pool[T] {
	if len(objs) > size {
		panic("invalid objects len")
	}

	ch := make(chan T, size)
	for _, obj := range objs {
		ch <- obj
	}
	return &Pool[T]{ch: ch}
}

func (p *Pool[T]) Add(v T) {
	p.ch <- v
}

func (p *Pool[T]) Get() T {
	return <-p.ch
}

func (p *Pool[T]) GetNoWait() T {
	select{
	case v := <-p.ch:
		return v
	default:
		return *new(T)
	}
}

func (p *Pool[T]) Len() int {
	return len(p.ch)
}
