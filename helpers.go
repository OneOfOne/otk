package otk

import (
	"bytes"
	"fmt"
	"reflect"
)

func Lesser[T comparable]() func(a, b T) bool {
	var zero T
	tt := reflect.TypeOf(zero)
	switch tt.Kind() {
	case reflect.String:
		return func(a, b T) bool {
			av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
			return av.String() < bv.String()
		}
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return func(a, b T) bool {
			av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
			return av.Int() < bv.Int()
		}
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return func(a, b T) bool {
			av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
			return av.Uint() < bv.Uint()
		}
	case reflect.Float64, reflect.Float32:
		return func(a, b T) bool {
			av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
			return av.Float() < bv.Float()
		}
	case reflect.Slice, reflect.Array:
		if tt.Elem().Kind() == reflect.Uint8 {
			return func(a, b T) bool {
				av, bv := reflect.ValueOf(a), reflect.ValueOf(b)
				return bytes.Compare(av.Bytes(), bv.Bytes()) < 0
			}
		}
	}

	panic("invalid")
}

func fmtChar[T comparable]() string {
	var zero T
	if k := reflect.TypeOf(zero).Kind(); k == reflect.String {
		return "%q"
	}
	return "%v"
}

func sliceToString[S ~[]E, E comparable](s S) []string {
	out, ch := make([]string, 0, len(s)), fmtChar[E]()
	for _, v := range s {
		out = append(out, fmt.Sprintf(ch, v))
	}
	return out
}
