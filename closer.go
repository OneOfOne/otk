package otk

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.oneofone.dev/genh"
	"go.oneofone.dev/oerrs"
)

func NewCloser(onClose func(name string, took time.Duration)) *Closer {
	ctx, cfn := context.WithCancel(context.Background())
	if onClose == nil {
		onClose = func(name string, took time.Duration) {}
	}
	return &Closer{
		ctx:     ctx,
		cfn:     cfn,
		onClose: onClose,
	}
}

type closerFn struct {
	fn   func() error
	name string
}

type Closer struct {
	ctx     context.Context
	cfn     func()
	onClose func(name string, took time.Duration)
	fns     []closerFn
	fnsSync []closerFn
	mux     sync.Mutex
}

func (c *Closer) Add(name string, fn func() error, sync bool) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.cfn == nil {
		return os.ErrClosed
	}
	if sync {
		c.fnsSync = append(c.fnsSync, closerFn{name: name, fn: fn})
	} else {
		c.fns = append(c.fns, closerFn{name: name, fn: fn})
	}
	return nil
}

func (c *Closer) Delete(name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.cfn == nil {
		return os.ErrClosed
	}
	c.fns = genh.Filter(c.fns, func(cfn closerFn) (keep bool) {
		return cfn.name != name
	}, true)
	c.fnsSync = genh.Filter(c.fns, func(cfn closerFn) (keep bool) {
		return cfn.name != name
	}, true)
	return nil
}

func (c *Closer) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.cfn == nil {
		return os.ErrClosed
	}
	c.cfn()
	c.cfn = nil
	var wg sync.WaitGroup
	errs := oerrs.NewSafeList(false)
	for _, cfn := range c.fnsSync {
		start := time.Now()
		if err := cfn.fn(); err != nil {
			errs.Errorf("error closing %s: %v", cfn.name, err)
		}

		c.onClose(cfn.name, time.Since(start))
	}
	for _, cfn := range c.fns {
		wg.Add(1)
		go func(cfn closerFn) {
			defer wg.Done()
			start := time.Now()
			if err := cfn.fn(); err != nil {
				errs.Errorf("error closing %s: %v", cfn.name, err)
			}
			c.onClose(cfn.name, time.Since(start))
		}(cfn)
	}
	wg.Wait()

	return errs.Err()
}

func (c *Closer) WaitSignal(signals ...os.Signal) error {
	ctx, _ := signal.NotifyContext(c.ctx, signals...)
	return c.Wait(ctx)
}

func (c *Closer) Wait(ctx context.Context) error {
	select {
	case <-ctx.Done():
	case <-c.ctx.Done():
	}
	return c.Close()
}
