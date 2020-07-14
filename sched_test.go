package otk

import (
	"context"
	"testing"
	"time"
)

func TestScheduler(t *testing.T) {
	sch := NewScheduler(context.Background())
	defer sch.StopAll()

	count := 0
	sch.Start("1 sec", func(ctx context.Context, now time.Time) error {
		count++
		return nil
	}, time.Millisecond, time.Millisecond)

	time.Sleep(time.Second / 2)
	if count != 499 {
		t.Fatalf("expected 499, got %d", count)
	}
}
