package otk

func NewSem(size int) *Sem {
	if size == 0 {
		size = 1
	}

	return &Sem{ch: make(chan struct{}, size)}
}

type Sem struct {
	ch chan struct{}
}

func (s *Sem) Acquire() {
	s.ch <- struct{}{}
}

func (s *Sem) Release() {
	<-s.ch
}

func (s *Sem) Go(fn func()) {
	s.Acquire()
	go func() {
		defer s.Release()
		fn()
	}()
}
