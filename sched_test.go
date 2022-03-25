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
	sch.Start("w00t", func(ctx context.Context, now time.Time) error {
		count++
		return nil
	}, time.Millisecond, time.Millisecond)

	time.Sleep(time.Second / 4)
	sch.Stop("w00t")
	time.Sleep(time.Second / 4)
	if count < 200 || count > 250 {
		t.Fatalf("expected 249-250, got %d", count)
	}
}
