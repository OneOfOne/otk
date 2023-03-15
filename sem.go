package otk

import (
	"errors"
	"sync"
)

var (
	ErrClosedSem      = errors.New("sem is already closed")
	ErrNegativeNumber = errors.New("n < 1")
	ErrSemIsFull      = errors.New("sem is full")
)

func NewSem(size int) *Sem {
	if size == 0 {
		size = 1
	}

	return &Sem{ch: make(chan struct{}, size)}
}

type Sem struct {
	ch chan struct{}
	m  sync.Mutex
}

func (s *Sem) Acquire(n int) error {
	if n < 1 {
		return ErrNegativeNumber
	}

	var e struct{}
	for i := 0; i < n; i++ {
		s.m.Lock()
		if s.ch == nil {
			s.m.Unlock()
			return ErrClosedSem
		}

		s.ch <- e
		s.m.Unlock()
	}

	return nil
}

func (s *Sem) Release(n int) error {
	if n < 1 {
		return ErrNegativeNumber
	}

	for i := 0; i < n; i++ {
		s.m.Lock()
		if s.ch == nil {
			s.m.Unlock()
			return ErrClosedSem
		}
		<-s.ch
		s.m.Unlock()
	}

	return nil
}

func (s *Sem) Go(fn func()) error {
	if err := s.Acquire(1); err != nil {
		return err
	}
	go func() {
		defer s.Release(1)
		fn()
	}()

	return nil
}

func (s *Sem) Close() error {
	s.m.Lock()
	defer s.m.Unlock()
	if s.ch == nil {
		return ErrClosedSem
	}
	close(s.ch)
	s.ch = nil
	return nil
}
