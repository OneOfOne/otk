package otk

import (
	"context"
	"time"
)

// Retry is an alias for RetryCtx(context.Background(), fn, attempts, delay, backoffMod)
func Retry(fn func() error, attempts uint, delay time.Duration, backoffMod float64) error {
	return RetryCtx(context.Background(), fn, attempts, delay, backoffMod)
}

// RetryCtx calls fn every (delay * backoffMod) until it returns nil, the passed ctx is done or attempts are reached.
func RetryCtx(ctx context.Context, fn func() error, attempts uint, delay time.Duration, backoffMod float64) error {
	if delay == 0 {
		delay = time.Second
	}

	if attempts == 0 {
		attempts = 1
	}

	if backoffMod == 0 {
		backoffMod = 1
	}

	ret := make(chan error, 1)

	go func() {
		var err error
		for ; attempts > 0; attempts-- {
			if err = fn(); err == nil {
				break
			}
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * backoffMod)
		}
		ret <- err
	}()

	select {
	case err := <-ret:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
