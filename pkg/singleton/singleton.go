package singleton

import "sync"

type Singleton[T any] struct {
	once sync.Once
	obj  T
	ctor func() T
}

func NewSingleton[T any](ctor func() T) *Singleton[T] {
	return &Singleton[T]{
		ctor: ctor,
	}
}

func (s *Singleton[T]) New() T {
	s.once.Do(func() {
		if s.ctor == nil {
			s.obj = *new(T)
		} else {
			s.obj = s.ctor()
		}
	})
	return s.obj
}
