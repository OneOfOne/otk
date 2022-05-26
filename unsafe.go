package otk

import (
	"unsafe"
)

// UnsafeString can spawn dragons out of bytes
func UnsafeString(p []byte) string {
	return *(*string)(unsafe.Pointer(&p))
}

type stringCap struct {
	string
	int
}

// UnsafeBytes can spawn even bigger dragons, summon wisely
func UnsafeBytes(s string) (out []byte) {
	return *(*[]byte)(unsafe.Pointer(&stringCap{s, len(s)}))
}
