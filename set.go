package otk

import (
	"encoding/json"
	"sort"
	"unsafe"
)

var empty struct{}

func SetOf[T comparable](keys ...T) Set[T] {
	s := make(Set[T], len(keys))
	s.Add(keys...)
	return s
}

// Set is a simple set.
type Set[T comparable] map[T]struct{}

func (s Set[T]) init() Set[T] {
	if s == nil {
		s = Set[T]{}
	}
	return s
}

func (s *Set[T]) Set(keys ...T) Set[T] {
	ss := s.Add(keys...)
	if *s == nil {
		*s = ss
	}
	return ss
}

func (s Set[T]) Add(keys ...T) Set[T] {
	s = s.init()
	for _, k := range keys {
		s[k] = empty
	}
	return s
}

// AddIfNotExists returns true if the key was added, false if it already existed
func (s *Set[T]) AddIfNotExists(key T) bool {
	sm := s.init()
	if *s == nil {
		*s = sm
	}
	if _, ok := sm[key]; ok {
		return false
	}

	sm[key] = empty
	return true
}

func (s Set[T]) Clone() Set[T] {
	ns := make(Set[T], len(s))
	for k, v := range s {
		ns[k] = v
	}
	return ns
}

func (s Set[T]) Merge(os ...Set[T]) Set[T] {
	s = s.init()
	for _, o := range os {
		for k := range o {
			s[k] = empty
		}
	}
	return s
}

func (s Set[T]) Delete(keys ...T) Set[T] {
	for _, k := range keys {
		delete(s, k)
	}
	return s
}

func (s Set[T]) Has(key T) bool {
	_, ok := s[key]
	return ok
}

func (s Set[T]) Equal(os Set[T]) bool {
	if len(os) != len(s) {
		return false
	}

	for k := range s {
		if _, ok := os[k]; !ok {
			return false
		}
	}
	return true
}

func (s Set[T]) Len() int {
	return len(s)
}

func (s Set[T]) Keys() []T {
	if s == nil {
		return nil
	}
	keys := make([]T, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s Set[T]) SortedKeys() []T {
	if s == nil {
		return nil
	}
	keys := s.Keys()
	cmpFn := Lesser[T]()
	sort.Slice(keys, func(i, j int) bool {
		return cmpFn(keys[i], keys[j])
	})
	return keys
}

func (s Set[T]) String() string {
	v, _ := json.Marshal(s)
	return *(*string)(unsafe.Pointer(&v))
}

func (s Set[T]) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.SortedKeys())
}

func (s *Set[T]) UnmarshalJSON(data []byte) (err error) {
	var keys []T
	if err = json.Unmarshal(data, &keys); err == nil {
		s.Set(keys...)
	}
	return
}
