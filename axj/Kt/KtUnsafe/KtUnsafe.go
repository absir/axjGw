package KtUnsafe

import (
	"reflect"
	"unsafe"
)

func StringToBytes(s string) (b []byte) {
	pB := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	pS := (*reflect.StringHeader)(unsafe.Pointer(&s))
	pB.Data = pS.Data
	pB.Len = pS.Len
	pB.Cap = pS.Len
	return
}

func BytesToString(b []byte) (s string) {
	//pB := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	//pS := (*reflect.StringHeader)(unsafe.Pointer(&s))
	//pS.Data = pB.Data
	//pS.Len = pB.Len
	s = *(*string)(unsafe.Pointer(&b))
	return
}

func StringToRunes(s string) []rune {
	return []rune(s)
}

func RunesToString(b []rune) string {
	return string(b)
}

func IndexByte(bs []byte, b byte, start int, end int) int {
	for ; start < end; start++ {
		if bs[start] == b {
			return start
		}
	}

	return -1
}

func PointerHash(b interface{}) int {
	return *(*int)(unsafe.Pointer(&b))
}
