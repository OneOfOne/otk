package sets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"unsafe"
)

func boolp(b bool) *bool {
	return &b
}

var empty struct{}

func SetOf(keys ...string) Set {
	s := Set{}
	s.Add(keys...)
	return s
}

// Set is a simple set.
type Set map[string]struct{}

func (s Set) init() Set {
	if s == nil {
		s = Set{}
	}
	return s
}

func (s *Set) Set(keys ...string) Set {
	ss := s.Add(keys...)
	if *s == nil {
		*s = ss
	}
	return ss
}

func (s Set) Add(keys ...string) Set {
	s = s.init()
	for _, k := range keys {
		s[k] = empty
	}
	return s
}

// AddIfNotExists returns true if the key was added, false if it already existed
func (s *Set) AddIfNotExists(key string) bool {
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

func (s Set) Clone() Set {
	ns := make(Set, len(s))
	for k, v := range s {
		ns[k] = v
	}
	return ns
}

func (s Set) Merge(os ...Set) Set {
	s = s.init()
	for _, o := range os {
		for k := range o {
			s[k] = empty
		}
	}
	return s
}

func (s Set) Delete(keys ...string) Set {
	for _, k := range keys {
		delete(s, k)
	}
	return s
}

func (s Set) Has(key string) bool {
	_, ok := s[key]
	return ok
}

func (s Set) Equal(os Set) bool {
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

func (s Set) Len() int {
	return len(s)
}

func (s Set) Keys() []string {
	if s == nil {
		return nil
	}
	keys := make([]string, 0, len(s))
	for k := range s {
		keys = append(keys, k)
	}
	return keys
}

func (s Set) SortedKeys() []string {
	keys := s.Keys()
	sort.Strings(keys)
	return keys
}

func (s Set) strLen() int {
	ln := 2
	for k := range s {
		ln += len(k) + 3
	}
	return ln - 1
}

func (s Set) append(buf *bytes.Buffer, sorted, checkLen bool) *bytes.Buffer {
	if len(s) == 0 {
		buf.WriteString("[]")
		return buf
	}

	var keys []string
	if sorted {
		keys = s.SortedKeys()
	} else {
		keys = s.Keys()
	}

	if checkLen {
		ln := 2 + (len(keys) * 3) // [] + len(keys) * `""`
		for _, k := range keys {
			ln += len(k)
		}

		if n := buf.Len() + ln - 1; n > buf.Cap() {
			buf.Grow(n)
		}
	}
	buf.WriteByte('[')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		fmt.Fprintf(buf, "%q", k)
	}
	buf.WriteByte(']')

	return buf
}

func (s Set) String() string {
	if len(s) == 0 {
		return "[]"
	}
	buf := s.append(&bytes.Buffer{}, true, true).Bytes()
	buf = buf[:len(buf):len(buf)]
	return *(*string)(unsafe.Pointer(&buf))
}

func (s Set) MarshalJSON() ([]byte, error) {
	buf := s.append(&bytes.Buffer{}, true, true).Bytes()
	buf = buf[:len(buf):len(buf)]
	return buf, nil
}

func (s *Set) UnmarshalJSON(data []byte) (err error) {
	var keys []string
	if err = json.Unmarshal(data, &keys); err == nil {
		s.Set(keys...)
	}
	return
}
