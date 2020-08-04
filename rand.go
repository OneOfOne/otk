package otk

import (
	"math/rand"
	"runtime"
	"sync/atomic"
	"time"
	"unsafe"
)

// XorShiftRNG simple efficient thread-safe lock-free rng, self seeds with time.Now().UnixNano() if 0
type XorShiftRNG uint64

func (x *XorShiftRNG) Seed(n int64) {
	nx := XorShiftRNG(n)
	atomic.StoreUint64((*uint64)(x), nx.Uint64())
}

func (x *XorShiftRNG) Uint64() (rv uint64) {
RETRY:
	a := atomic.LoadUint64((*uint64)(x))
	if a != 0 {
		rv = a | 0xCAFEBABE
	} else {
		rv = uint64(time.Now().UnixNano()) | 0xCAFEBABE
	}
	rv = rv ^ (rv << 21)
	rv = rv ^ (rv >> 35)
	rv = rv ^ (a << 4)
	if !atomic.CompareAndSwapUint64((*uint64)(x), a, rv) {
		runtime.Gosched()
		goto RETRY
	}
	return
}

func (x *XorShiftRNG) Int63() int64 {
	const (
		rngMax  = 1 << 63
		rngMask = rngMax - 1
	)
	return int64(x.Uint64() & rngMask)
}

func (x *XorShiftRNG) Bytes() [8]byte {
	n := x.Uint64()
	return *(*[8]byte)(unsafe.Pointer(&n))
}

var _ rand.Source = (*XorShiftRNG)(nil)
