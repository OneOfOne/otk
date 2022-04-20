package sets

import (
	"encoding/json"
	"sort"
	"sync"
)

func SafeSetOf(keys ...string) *SafeSet {
	s := SetOf(keys...)
	return &SafeSet{s: s}
}

type SafeSet struct {
	mux sync.RWMutex
	s   Set
}

func (ss *SafeSet) Set(keys ...string) *SafeSet {
	return ss.Add(keys...)
}

func (ss *SafeSet) Add(keys ...string) *SafeSet {
	ss.mux.Lock()
	ss.s.Set(keys...)
	ss.mux.Unlock()
	return ss
}

func (ss *SafeSet) AddIfNotExists(key string) bool {
	ss.mux.Lock()
	added := ss.s.AddIfNotExists(key)
	ss.mux.Unlock()
	return added
}

func (ss *SafeSet) Clone() *SafeSet {
	ss.mux.RLock()
	ns := ss.s.Clone()
	ss.mux.RUnlock()
	return &SafeSet{s: ns}
}

func (ss *SafeSet) MergeSafe(o *SafeSet) *SafeSet {
	ss.mux.Lock()
	o.mux.Lock()
	ss.s.Merge(o.s)
	o.mux.Unlock()
	ss.mux.Unlock()
	return ss
}

func (ss *SafeSet) Merge(o Set) *SafeSet {
	ss.mux.Lock()
	ss.s.Merge(o)
	ss.mux.Unlock()
	return ss
}

func (ss *SafeSet) Delete(keys ...string) *SafeSet {
	ss.mux.Lock()
	ss.s.Delete(keys...)
	ss.mux.Unlock()
	return ss
}

func (ss *SafeSet) Has(key string) bool {
	ss.mux.RLock()
	ok := ss.s.Has(key)
	ss.mux.RUnlock()
	return ok
}

func (ss *SafeSet) Len() int {
	ss.mux.RLock()
	ln := ss.s.Len()
	ss.mux.RUnlock()
	return ln
}

func (ss *SafeSet) Keys() []string {
	ss.mux.RLock()
	keys := ss.s.Keys()
	ss.mux.RUnlock()
	return keys
}

func (ss *SafeSet) SortedKeys() []string {
	keys := ss.Keys()
	sort.Strings(keys)
	return keys
}

func (ss *SafeSet) MarshalJSON() ([]byte, error) {
	keys := ss.Keys()
	return json.Marshal(keys)
}

func (ss *SafeSet) UnmarshalJSON(data []byte) error {
	var s Set
	if err := s.UnmarshalJSON(data); err != nil {
		return err
	}
	ss.mux.Lock()
	ss.s = s
	ss.mux.Unlock()
	return nil
}
