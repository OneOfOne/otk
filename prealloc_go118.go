package otk

import "sync"

type ptrGen[T any] struct {
	raw []T
	cap int
}

func (p *ptrGen[T]) next() (v *T) {
	if len(p.raw) == 0 {
		p.raw = make([]T, p.cap)
	}

	v = &p.raw[0]
	p.raw = p.raw[1:]

	return
}

func PtrGen[T any](cap int, safe bool) func() *T {
	if cap < 1 {
		cap = 64
	}

	p := ptrGen[T]{cap: cap}

	if !safe {
		return p.next
	}

	var mux sync.Mutex
	return func() *T {
		mux.Lock()
		v := p.next()
		mux.Unlock()
		return v
	}
}

func ValuesToPtrs[T any](vals []T, noCopy bool) []*T {
	out := make([]*T, 0, len(vals))
	for i := range vals {
		v := &vals[i]
		if noCopy {
			cp := *v
			v = &cp
		}
		out = append(out, v)
	}
	return out
}
