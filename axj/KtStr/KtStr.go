package KtStr

import (
	"io"
	"unsafe"
)

// 字符串转bytes
func ToBytes(s string) []byte {
	a := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{a[0], a[1], a[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

// bytes转字符串
func FromBytes(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// 版本比较
func CompareV(from, to string) int {
	return CompareVb(ToBytes(from), ToBytes(to))
}

func CompareVb(from, to []byte) int {
	fLen := len(from)
	tLen := len(to)
	mLen := 0
	if fLen < tLen {
		mLen = fLen

	} else {
		mLen = tLen
	}

	cmp := 0
	for i := 0; i < mLen; i++ {
		f := from[i]
		t := to[i]
		if cmp == 0 {
			if f > t {
				cmp = 1

			} else if f < t {
				cmp = -1
			}
		}

		if f == '.' {
			if t == '.' {
				if cmp != 0 {
					return cmp
				}

			} else {
				return -1
			}

		} else if t == '.' {
			return 1
		}
	}

	if fLen == tLen {
		return cmp
	}

	if cmp != 0 {
		if (fLen > tLen && from[mLen] == '.') || (fLen < tLen && to[mLen] == '.') {
			return cmp
		}
	}

	if fLen > tLen {
		return 1
	}

	return -1
}

// 字符串比较
func CompareBFt(from, to []byte, fLen, tLen int) int {
	mLen := 0
	if fLen < tLen {
		mLen = fLen

	} else {
		mLen = tLen
	}

	for i := 0; i < mLen; i++ {
		f := from[i]
		t := to[i]
		if f > t {
			return 1

		} else if f < t {
			return -1
		}
	}

	if fLen > tLen {
		return 1

	} else if fLen < tLen {
		return -1
	}

	return 0
}

// 字符串尾部比较
func CompareBEndFt(from, to []byte, fLen, tLen int) int {
	mLen := 0
	if fLen < tLen {
		mLen = fLen

	} else {
		mLen = tLen
	}

	for i := 1; i <= mLen; i++ {
		f := from[tLen-i]
		t := to[tLen-i]
		if f > t {
			return 1

		} else if f < t {
			return -1
		}
	}

	if fLen > tLen {
		return 1

	} else if fLen < tLen {
		return -1
	}

	return 0
}

// 查找从位置
func Index(s, substr string, from int) int {
	i := from
	if i < 0 {
		i = 0
	}

	subL := len(substr)
	if subL == 0 {
		return i
	}

	b := substr[0]
	max := len(s) - subL
	for ; i <= max; i++ {
		if s[i] == b {
			j := 1
			for ; j < subL; j++ {
				if s[i+j] != substr[j] {
					break
				}
			}

			if j >= subL {
				return i
			}
		}
	}

	return -1
}

// 查找尾部从位置
func LastIndex(s, substr string, from int) int {
	i := from
	max := len(s) - 1
	if i > max || i < 0 {
		i = max
	}

	min := len(substr)
	if min == 0 {
		return i
	}

	min = min - 1
	b := substr[min]
	for ; i >= min; i-- {
		if s[i] == b {
			j := 0
			for ; j < min; j++ {
				if s[i-min+j] != substr[j] {
					break
				}
			}

			if j >= min {
				return i - min
			}
		}
	}

	return -1
}

// 字符串拼接
func WriteBytes(write io.ByteWriter, bs []byte, off, len int) io.ByteWriter {
	for ; off < len; off++ {
		write.WriteByte(bs[off])
	}

	return write
}
