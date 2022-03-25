package otk

import (
	"encoding/json"
	"sort"
	"unsafe"
)

type MultiSet[T1, T2 comparable] map[T1]Set[T2]

func (s MultiSet[T1, T2]) init() MultiSet[T1, T2] {
	if s == nil {
		s = MultiSet[T1, T2]{}
	}
	return s
}

func (s *MultiSet[T1, T2]) Set(key T1, values ...T2) MultiSet[T1, T2] {
	ss := s.Add(key, values...)
	if *s == nil {
		*s = ss
	}
	return ss
}

func (s MultiSet[T1, T2]) Add(key T1, values ...T2) MultiSet[T1, T2] {
	s = s.init()
	s[key] = s[key].Add(values...)
	return s
}

// AddIfNotExists returns true if the key was added, false if it already existed
func (s MultiSet[T1, T2]) AddIfNotExists(key T1, value T2) bool {
	if m := s[key]; m != nil {
		return m.AddIfNotExists(value)
	}
	s[key] = SetOf(value)
	return true
}

func (s MultiSet[T1, T2]) Clone() MultiSet[T1, T2] {
	ns := make(MultiSet[T1, T2], len(s))
	for k, v := range s {
		ns[k] = v.Clone()
	}
	return ns
}

func (s MultiSet[T1, T2]) Values(key T1) Set[T2] {
	return s[key]
}

func (s MultiSet[T1, T2]) Merge(o MultiSet[T1, T2]) MultiSet[T1, T2] {
	s = s.init()
	for k, v := range o {
		s[k] = s[k].Merge(v)
	}
	return s
}

func (s MultiSet[T1, T2]) MergeSet(key T1, o Set[T2]) MultiSet[T1, T2] {
	s = s.init()
	s[key] = s[key].Merge(o)
	return s
}

func (s MultiSet[T1, T2]) Delete(keys ...T1) MultiSet[T1, T2] {
	for _, k := range keys {
		delete(s, k)
	}
	return s
}

func (s MultiSet[T1, T2]) DeleteValues(key T1, values ...T2) MultiSet[T1, T2] {
	m := s[key]
	for _, v := range values {
		delete(m, v)
	}
	if len(m) == 0 {
		delete(s, key)
	}
	return s
}

func (s MultiSet[T1, T2]) Has(key T1, sub T2) bool {
	return s[key].Has(sub)
}

func (s MultiSet[T1, T2]) Match(fn func(key T1, s Set[T2]) bool, all bool) bool {
	for k, v := range s {
		b := fn(k, v)
		if b && !all {
			return true
		}

		if !b && all {
			return false
		}
	}

	return all
}

func (s MultiSet[T1, T2]) Equal(os MultiSet[T1, T2]) bool {
	if len(os) != len(s) {
		return false
	}

	for k, ss := range s {
		if !os[k].Equal(ss) {
			return false
		}
	}
	return true
}

func (s MultiSet[T1, T2]) Len() int {
	return len(s)
}

func (s MultiSet[T1, T2]) Keys() []T1 {
	keys := make([]T1, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MultiSet[T1, T2]) SortedKeys() []T1 {
	if s == nil {
		return nil
	}
	keys := s.Keys()
	cmpFn := Lesser[T1]()
	sort.Slice(keys, func(i, j int) bool {
		return cmpFn(keys[i], keys[j])
	})
	return keys
}

func (s MultiSet[T1, T2]) String() string {
	v, _ := json.Marshal(s)
	return *(*string)(unsafe.Pointer(&v))
}
