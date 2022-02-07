package sets

import (
	"bytes"
	"fmt"
	"sort"
	"unsafe"
)

type MultiSet map[string]Set

func (s MultiSet) init() MultiSet {
	if s == nil {
		s = MultiSet{}
	}
	return s
}

func (s *MultiSet) Set(key string, values ...string) MultiSet {
	ss := s.Add(key, values...)
	if *s == nil {
		*s = ss
	}
	return ss
}

func (s MultiSet) Add(key string, values ...string) MultiSet {
	s = s.init()
	s[key] = s[key].Add(values...)
	return s
}

// AddIfNotExists returns true if the key was added, false if it already existed
func (s MultiSet) AddIfNotExists(key, value string) bool {
	if m := s[key]; m == nil {
		return m.AddIfNotExists(value)
	}
	s[key] = SetOf(value)
	return true
}

func (s MultiSet) Clone() MultiSet {
	ns := make(MultiSet, len(s))
	for k, v := range s {
		ns[k] = v.Clone()
	}
	return ns
}

func (s MultiSet) Values(key string) Set {
	return s[key]
}

func (s MultiSet) Merge(o MultiSet) MultiSet {
	s = s.init()
	for k, v := range o {
		s[k] = s[k].Merge(v)
	}
	return s
}

func (s MultiSet) MergeSet(key string, o Set) MultiSet {
	s = s.init()
	s[key] = s[key].Merge(o)
	return s
}

func (s MultiSet) Delete(keys ...string) MultiSet {
	for _, k := range keys {
		delete(s, k)
	}
	return s
}

func (s MultiSet) DeleteValues(key string, values ...string) MultiSet {
	m := s[key]
	for _, v := range values {
		delete(m, v)
	}
	if len(m) == 0 {
		delete(s, key)
	}
	return s
}

func (s MultiSet) Has(key, sub string) bool {
	return s[key].Has(sub)
}

func (s MultiSet) Match(fn func(key string, s Set) bool, all bool) bool {
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

func (s MultiSet) Equal(os MultiSet) bool {
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

func (s MultiSet) Len() int {
	return len(s)
}

func (s MultiSet) Keys() []string {
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s MultiSet) SortedKeys() []string {
	keys := s.Keys()
	sort.Strings(keys)
	return keys
}

func (s MultiSet) append(buf *bytes.Buffer, sorted bool) *bytes.Buffer {
	if len(s) == 0 {
		buf.WriteString("{}")
		return buf
	}

	var keys []string
	if sorted {
		keys = s.SortedKeys()
	} else {
		keys = s.Keys()
	}

	ln := 2 + (len(keys) * 4) // [] + len(keys) * `",:"`
	for _, k := range keys {
		ln += len(k) + s[k].strLen() // at least "[]"
	}

	if n := buf.Len() + ln - 1; n > buf.Cap() {
		buf.Grow(n)
	}

	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, "%q:", k)
		buf = s[k].append(buf, sorted, false)
	}
	buf.WriteByte('}')
	return buf
}

func (s MultiSet) String() string {
	if len(s) == 0 {
		return "{}"
	}
	buf := s.append(&bytes.Buffer{}, true).Bytes()
	buf = buf[:len(buf):len(buf)]
	return *(*string)(unsafe.Pointer(&buf))
}

func (s MultiSet) MarshalJSON() ([]byte, error) {
	buf := s.append(&bytes.Buffer{}, true).Bytes()
	buf = buf[:len(buf):len(buf)]
	return buf, nil
}
