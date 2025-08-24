package util

import "sync"

type Collector[T any] struct {
	inner []T
	lock  sync.Mutex
}

func (c *Collector[T]) Add(value T) {
	c.lock.Lock()
	c.inner = append(c.inner, value)
	c.lock.Unlock()
}

func (c *Collector[T]) Values() []T {
	return c.inner
}

func NewCollector[T any]() *Collector[T] {
	return &Collector[T]{}
}
