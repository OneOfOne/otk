package otk

import (
	"reflect"
	"unsafe"
)

// UnsafeString can spawn dragons out of bytes
func UnsafeString(p []byte) string {
	return *(*string)(unsafe.Pointer(&p))
}

func UnsafeBytes(s string) (out []byte) {
	ss := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bs := (*reflect.SliceHeader)(unsafe.Pointer(&out))
	bs.Data, bs.Len, bs.Cap = ss.Data, ss.Len, ss.Len
	return
}
