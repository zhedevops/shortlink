package main

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	items []T
}

func NewPool[T Resettable]() *Pool[T] {
	return &Pool[T]{}
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.items = append(p.items, obj)
}

func (p *Pool[T]) Get() T {
	var empty T

	n := len(p.items)
	if n == 0 {
		return empty
	}

	obj := p.items[n-1]
	p.items = p.items[:n-1]

	return obj
}
