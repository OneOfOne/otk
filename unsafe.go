package otk

import "unsafe"

// UnsafeString can spawn dragons out of bytes
func UnsafeString(p []byte) string {
	return *(*string)(unsafe.Pointer(&p))
}
