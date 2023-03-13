package otk

import (
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"go.oneofone.dev/genh"
)

type closable struct {
	i atomic.Int64
}

func (c *closable) Close() error {
	c.i.Add(-1)
	return nil
}

func TestCloser(t *testing.T) {
	var cc closable
	var closed genh.LSlice[string]
	c := NewCloser(func(name string, took time.Duration) {
		closed.Append(name)
		// t.Logf("closed %s in %s", name, took)
	})
	for i := 0; i < 10; i++ {
		cc.i.Add(1)
		c.Add("closable:"+strconv.Itoa(i), cc.Close, i%2 == 0)
	}
	if err := c.Close(); err != nil {
		t.Fatal(err)
	}
	syncExpected := []string{"closable:0", "closable:2", "closable:4", "closable:6", "closable:8"}
	for i, v := range syncExpected {
		if closed.Get(i) != v {
			t.Errorf("expected %s at %d, got %s", v, i, closed.Get(i))
		}
	}
}
