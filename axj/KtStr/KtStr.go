package KtStr

import (
	"axj/Kt"
	"container/list"
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

func IndexOf(str []rune, chr rune) int {
	len := len(str)
	for i := 0; i < len; i++ {
		c := str[i]
		if c == chr {
			return i
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

// 字符串分隔
func SplitStr(str string, sps string, trim bool, start int) []interface{} {
	return SplitStrBr(str, sps, trim, start, false, 0).([]interface{})
}

func SplitStrBr(str string, sps string, trim bool, start int, br bool, typ int) interface{} {
	strs := list.New()
	if br {
		brc := Kt.If(typ == 1, '}', ']').(rune)
		splitStrBrC([]rune(str), []rune(sps), trim, start, br, brc, strs)
		return splitStrM(strs, brc)

	} else {
		splitStrBrC([]rune(str), []rune(sps), trim, start, br, '{', strs)
	}

	return Kt.ToArray(strs)
}

func splitStrM(strs *list.List, brc rune) interface{} {
	if brc == '}' {
		maps := map[interface{}]interface{}{}
		if strs.Len() == 1 {
			maps["value"] = strs.Front().Value

		} else {
			var key interface{} = nil
			for el := strs.Front(); el != nil; el = el.Next() {
				if key == nil {
					key = el.Value

				} else {
					maps[key] = el.Value
					key = nil
				}
			}
		}

		return maps
	}

	return Kt.ToArray(strs)
}

func splitStrBrC(str []rune, sps []rune, trim bool, start int, br bool, brc rune, strs *list.List) int {
	if str == nil || sps == nil || strs == nil {
		return start
	}

	length := len(str)
	var chr rune
	si := start
	ei := -1
	for ; start < length; start++ {
		chr = str[start]
		if br {
			if chr == '{' || chr == '[' {
				sts := list.New()
				chr = Kt.If(chr == '{', '}', ']').(rune)
				start = splitStrBrC(str, sps, trim, start+1, br, chr, sts)
				strs.PushBack(splitStrM(sts, chr))
				ei = -2
				continue

			} else if chr == brc {
				if si < ei {
					strs.PushBack(string(str[si:ei]))

				} else if ei == -1 {
					strs.PushBack("")
					ei = -2
				}

				return start
			}
		}

		if IndexOf(sps, chr) < 0 {
			if trim && chr == ' ' {
				if si == start {
					si++
				}

			} else {
				ei = start + 1
			}

			continue
		}

		if si < ei {
			strs.PushBack(string(str[si:ei]))
			ei = -1

		} else if ei >= -1 {
			strs.PushBack("")
			ei = -1

		} else {
			ei = start
		}

		si = start + 1
	}

	if si < ei {
		strs.PushBack(string(str[si:ei]))
	}

	return start
}
