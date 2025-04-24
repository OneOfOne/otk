package otk

import (
	"strconv"
	"time"
)

// UniqueSlice returns all unique keys in `in` by modifying it in place
func UniqueSlice(in []string) (out []string) {
	set := make(Set, len(in))
	out = in[:0]

	for _, s := range in {
		if set.AddIfNotExists(s) {
			out = append(out, s)
		}
	}

	return
}

// TryParseTime will try to parse input in all provided go time formats.
// Accepts: U for unix ts, UN for unix nano.
func TryParseTime(ts string, layouts []string, tz *time.Location) time.Time {
	const ms = 10000000000
	var (
		t   time.Time
		err error
	)

	if tz == nil {
		tz = time.UTC
	}

	for _, l := range layouts {
		if l == "U" || l == "UN" {
			var n int64
			if n, err = strconv.ParseInt(ts, 10, 64); err != nil {
				continue
			}
			if l == "U" {
				if n > ms { // javascript ts in ms
					n = n / 1000
				}
				t = time.Unix(n, 0)
			} else {
				t = time.Unix(0, n)
			}
			return t.In(tz)
		}

		if t, err = time.ParseInLocation(l, ts, tz); err == nil {
			return t
		}
	}
	return time.Time{}
}

type M = map[string]interface{}

func MergeMap(dst, src M) M {
	if dst == nil {
		return src
	}
	for k, v := range src {
		// Check if the value is a nested map
		if vMap, ok := v.(M); ok {
			if len(vMap) == 0 {
				// Skip empty map
				continue
			}
			if dv, ok := dst[k].(M); ok {
				MergeMap(dv, vMap)
				continue
			}
		}
		// If value is not nil, or empty set it in dst
		if !isEmpty(v) {
			dst[k] = v
		}
	}
	return dst
}

func isEmpty(v any) bool {
	switch v := v.(type) {
	case nil:
		return true
	case int, int8, int16, int32, int64:
		return v == 0
	case uint, uint8, uint16, uint32, uint64:
		return v == 0
	case float32, float64:
		return v == 0
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case M:
		return len(v) == 0
	default:
		return false
	}
}
