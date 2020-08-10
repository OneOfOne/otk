package reallyunsafe

import "unsafe"

// Slice is the runtime representation of a slice.
// It cannot be used safely or portably and its representation may
// change in a later release.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// String is the runtime representation of a string.
// It cannot be used safely or portably and its representation may
// change in a later release.
//
// Unlike reflect.StringHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type String struct {
	Data unsafe.Pointer
	Len  int
}

func StringToBytes(s string) (out []byte) {
	ss := (*String)(unsafe.Pointer(&s))
	bs := &Slice{ss.Data, ss.Len, ss.Len}
	out = *(*[]byte)(unsafe.Pointer(bs))
	return
}

func BytesToString(s []byte) (out string) {
	ss := (*String)(unsafe.Pointer(&s))
	out = *(*string)(unsafe.Pointer(ss))
	return
}
