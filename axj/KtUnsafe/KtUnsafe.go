package KtUnsafe

import (
	"unsafe"
)

func StringToBytes(s string) []byte {
	return *(*[]byte)(unsafe.Pointer(&s))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
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
