package otk

import (
	"context"
	"errors"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

type TaskFunc = func(ctx context.Context, t time.Time) error

var (
	// ErrStopTask can be returned to stop the run loop
	ErrStopTask = errors.New("STOP")
)

func NewScheduler(pctx context.Context) *Scheduler {
	ctx, cfn := context.WithCancel(pctx)

	return &Scheduler{
		tasks: make(map[string]*task),

		ctx: ctx,
		cfn: cfn,
	}
}

type Scheduler struct {
	mux   sync.Mutex
	tasks map[string]*task

	ctx context.Context
	cfn context.CancelFunc
}

func (c *Scheduler) Start(id string, fn TaskFunc, startIn, thenEvery time.Duration) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.tasks[id] != nil {
		return xerrors.Errorf("task %q already exists", id)
	}

	ctx, cfn := context.WithCancel(c.ctx)
	tsk := &task{
		ctx:   ctx,
		cfn:   cfn,
		tk:    time.NewTicker(startIn),
		every: thenEvery,

		fn: fn,
	}

	c.tasks[id] = tsk
	go func() {
		defer c.Stop(id)
		tsk.run()
	}()

	return nil
}

func (c *Scheduler) Stop(id string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	t := c.tasks[id]
	if t == nil {
		return xerrors.Errorf("task %q does not exist", id)
	}
	t.stop()
	delete(c.tasks, id)
	return nil

}

func (c *Scheduler) StopAll() {
	c.cfn()
}

type task struct {
	ctx   context.Context
	cfn   context.CancelFunc
	tk    *time.Ticker
	every time.Duration

	fn TaskFunc
}

func (t *task) run() {
	defer t.tk.Stop()
	hitFirst := false
	for {
		select {
		case now := <-t.tk.C:
			if t.fn(t.ctx, now) == ErrStopTask {
				return
			}

			if !hitFirst {
				hitFirst = true
				t.tk.Stop()
				t.tk = time.NewTicker(t.every)
			}

		case <-t.ctx.Done():
			return
		}
	}
}

func (t *task) stop() {
	t.cfn()
}

// TimeUntil is a tiny helper to return the duration until hour:min:sec
// if the duration is in the past and nextDay is true, it'll add 24 hours.
func TimeUntil(t time.Time, hour, min, sec int, nextDay bool) time.Duration {
	y, m, d := t.Date()
	nt := time.Date(y, m, d, hour, min, sec, 0, t.Location())

	if nextDay && nt.Before(t) {
		nt = nt.AddDate(0, 0, 1)
	}

	return nt.Sub(t)
}
