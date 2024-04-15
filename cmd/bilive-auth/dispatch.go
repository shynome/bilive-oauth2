package main

import (
	"sync"
)

type Dispatcher[T any] struct {
	listeners map[string]chan T
	locker    *sync.RWMutex
}

func NewDisptacher[T any]() *Dispatcher[T] {
	d := &Dispatcher[T]{
		listeners: make(map[string]chan T),
		locker:    &sync.RWMutex{},
	}
	return d
}

func (d Dispatcher[T]) Dispatch(item T) {
	d.locker.RLock()
	defer d.locker.RUnlock()
	for _, ch := range d.listeners {
		go func(ch chan T) {
			ch <- item
		}(ch)
	}
}
func (d Dispatcher[T]) Listen(vid string) <-chan T {
	ch := make(chan T)
	d.listeners[vid] = ch
	return ch
}

func (d Dispatcher[T]) Free(k string) {
	d.locker.Lock()
	defer d.locker.Unlock()
	ch := d.listeners[k]
	close(ch)
	delete(d.listeners, k)
}
