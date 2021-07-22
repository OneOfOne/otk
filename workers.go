package otk

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrClosedPool = errors.New("the pool is closed, go swim somewhere else")
	ErrPoolIsFull = errors.New("the pool is full")
)

func NewWorkers(ctx context.Context, initial, increaseBy int) *Workers {
	ctx, cfn := context.WithCancel(ctx)
	w := &Workers{
		ch:  make(chan func(ctx context.Context), initial),
		ctx: ctx,
		cfn: cfn,
		inc: increaseBy,
	}
	go w.init(initial)
	return w
}

type Workers struct {
	mux   sync.Mutex
	ch    chan func(ctx context.Context)
	ctx   context.Context
	cfn   context.CancelFunc
	inc   int
	total int64
}

func (w *Workers) Exec(fn func(context.Context)) (_ error) {
	if err := w.ctx.Err(); err != nil {
		return ErrClosedPool
	}

	for i := 0; i < 3; i++ {
		select {
		case w.ch <- fn:
			return
		case <-w.ctx.Done():
			return ErrClosedPool
		default:
			w.spawn(w.inc)
			runtime.Gosched()
		}
	}

	log.Printf("workers: we are being overrun :(, count: %v, chan: %v", atomic.LoadInt64(&w.total), len(w.ch))
	return ErrPoolIsFull
}

func (w *Workers) Close() error {
	if w.ctx.Err() != nil {
		return ErrClosedPool
	}
	w.cfn()
	close(w.ch)
	return nil
}

func (w *Workers) init(n int) {
	w.spawn(n)
	tk := time.NewTicker(time.Minute * 5)
	defer tk.Stop()
	for {
		select {
		case <-tk.C:
			// skip if the channel isn't empty or the number of workers is less than the initial
			if len(w.ch) > 0 || atomic.LoadInt64(&w.total) <= int64(cap(w.ch)) {
				continue
			}
			for i := 0; i < w.inc; i++ {
				select {
				case <-w.ctx.Done():
					return
				case w.ch <- nil:
				}
			}
			atomic.AddInt64(&w.total, -int64(w.inc))
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Workers) spawn(n int) {
	for i := 0; i < n; i++ {
		go w.worker()
	}
	atomic.AddInt64(&w.total, int64(n))
}

func (w *Workers) worker() {
	for {
		select {
		case fn := <-w.ch:
			if fn == nil {
				return
			}
			fn(w.ctx)
		case <-w.ctx.Done():
			return
		}
	}
}
