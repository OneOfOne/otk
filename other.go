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
