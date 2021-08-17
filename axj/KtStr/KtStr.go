package KtStr

import (
	"axj/Kt"
	"axj/KtCvt"
	"axj/KtUnsafe"
	"container/list"
	"io"
	"reflect"
	"regexp"
	"strings"
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

func UpperCase(chr byte) byte {
	if chr >= 'a' && chr <= 'z' {
		return (chr - 32)
	}

	return chr
}

func LowerCase(chr byte) byte {
	if chr >= 'A' && chr <= 'Z' {
		return chr + 32
	}

	return chr
}

func Cap(str string) string {
	chr := str[0]
	if chr >= 'a' && chr <= 'z' {
		b := []byte(str)
		b[0] -= 32
		str = KtUnsafe.BytesToString(b)
	}

	return str
}

func UnCap(str string, strict bool) string {
	chr := str[0]
	if chr >= 'A' && chr <= 'Z' {
		if strict && len(str) > 1 {
			chr = str[1]
			if chr >= 'A' && chr <= 'Z' {
				return str
			}
		}

		b := []byte(str)
		b[0] += 32
		str = KtUnsafe.BytesToString(b)
	}

	return str
}

func Cmp(str string, to string, m int, n int) int {
	if m == 0 {
		return n
	}

	if n == 0 {
		return m
	}

	nL := n + 1
	mtx := make([][]int, m+1)
	for i := 0; i <= m; i++ {
		mtx[i] = make([]int, nL)
	}

	for i := 0; i <= m; i++ {
		mtx[i][0] = i
	}

	for j := 0; j <= n; j++ {
		mtx[0][j] = j
	}

	for i := 0; i < m; i++ {
		chr := str[i]
		for j := 0; j < n; j++ {
			mtx[i+1][j+1] = Kt.Min(mtx[i][j+1]+1, mtx[i+1][j]+1,
				mtx[i][j]+Kt.If(chr == to[j], 0, 1).(int))
		}
	}

	return mtx[m][n]
}

func Sim(str, to string) float32 {
	if str == to {
		return 1
	}

	if str == "" {
		return 0
	}

	m := len(str)
	n := len(to)
	if m == 0 && n == 0 {
		return 1
	}

	return 1.0 - float32(Cmp(str, to, m, n))/float32(Kt.If(m > n, m, n).(int))
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

func IndexByte(str string, chr byte, from int) int {
	i := from
	if i < 0 {
		i = 0
	}

	lenS := len(str)
	for ; i < lenS; i++ {
		c := str[i]
		if c == chr {
			return i
		}
	}

	return -1
}

func IndexBytes(str string, chrs []byte, from int) int {
	i := from
	if i < 0 {
		i = 0
	}

	lenC := len(chrs)
	lenS := len(str)
	for ; i < lenS; i++ {
		c := str[i]
		for j := 0; j < lenC; j++ {
			if c == chrs[j] {
				return i
			}
		}
	}

	return -1
}

func IndexRune(str []rune, chr rune, from int) int {
	i := from
	if i < 0 {
		i = 0
	}

	lenS := len(str)
	for ; i < lenS; i++ {
		c := str[i]
		if c == chr {
			return i
		}
	}

	return -1
}

func IndexRunes(str []rune, chrs []rune, from int) int {
	i := from
	if i < 0 {
		i = 0
	}

	lenC := len(chrs)
	lenS := len(str)
	for ; i < lenS; i++ {
		c := str[i]
		for j := 0; j < lenC; j++ {
			if c == chrs[j] {
				return i
			}
		}
	}

	return -1
}

func CountByte(str string, chr byte, start int, max int) int {
	count := 0
	for {
		start = IndexByte(str, chr, start)
		if start < 0 {
			return count
		}

		count++
		if max > 0 && max <= count {
			return count
		}

		start++
	}

	return count
}

func SplitByte(str string, chr byte, trim bool, start int, max int) []string {
	return SplitByteType(str, chr, trim, start, max, KtCvt.String).([]string)
}

func SplitByteType(str string, chr byte, trim bool, start int, max int, typ reflect.Type) interface{} {
	is := KtCvt.ForArrayIs(typ)
	if is == nil {
		return nil
	}

	last := CountByte(str, chr, start, max)
	strs := is.New(last + 1)
	last--
	for i := 0; i < last; i++ {
		end := IndexByte(str, chr, start)
		s := str[start:end]
		if trim {
			s = strings.TrimSpace(s)
		}

		is.Set(strs, i, KtCvt.ToType(s, typ))
		start = end + 1
	}

	s := str[start:]
	if trim {
		s = strings.TrimSpace(s)
	}

	is.Set(strs, last, KtCvt.ToType(s, typ))
	return strs
}

// 字符串分隔
func SplitStr(str string, sps string, trim bool, start int) []interface{} {
	return Kt.ToArray(SplitStrBr(str, sps, trim, start, false, 0, false).(*list.List))
}

func SplitStrS(str string, sps string, trim bool, start int, strict bool) []interface{} {
	return Kt.ToArray(SplitStrBr(str, sps, trim, start, false, 0, strict).(*list.List))
}

func SplitStrBr(str string, sps string, trim bool, start int, br bool, typ int, strict bool) interface{} {
	strs := list.New()
	if br {
		brc := Kt.If(typ == 1, '}', ']').(rune)
		splitStrBrC(KtUnsafe.StringToRunes(str), KtUnsafe.StringToRunes(sps), trim, start, br, brc, strs, strict)
		return splitStrM(strs, brc)

	} else {
		splitStrBrC(KtUnsafe.StringToRunes(str), KtUnsafe.StringToRunes(sps), trim, start, br, '{', strs, strict)
	}

	return strs
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

	return strs
}

func splitStrBrC(str []rune, sps []rune, trim bool, start int, br bool, brc rune, strs *list.List, strict bool) int {
	if str == nil || sps == nil || strs == nil {
		return start
	}

	length := len(str)
	var chr rune
	si := start
	ei := -2
	for ; start < length; start++ {
		chr = str[start]
		if br {
			if chr == '{' || chr == '[' {
				sts := list.New()
				chr = Kt.If(chr == '{', '}', ']').(rune)
				start = splitStrBrC(str, sps, trim, start+1, br, chr, sts, strict)
				strs.PushBack(splitStrM(sts, chr))
				ei = -2
				continue

			} else if chr == brc {
				if si < ei {
					strs.PushBack(KtUnsafe.RunesToString(str[si:ei]))

				} else if ei == -1 {
					if strict {
						strs.PushBack("")
					}
				}

				return start
			}
		}

		if IndexRune(sps, chr, 0) < 0 {
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
			strs.PushBack(KtUnsafe.RunesToString(str[si:ei]))
			ei = -1

		} else {
			if strict {
				strs.PushBack("")
			}

			ei = -1
		}

		si = start + 1
	}

	if si < ei {
		strs.PushBack(KtUnsafe.RunesToString(str[si:ei]))

	} else if ei == -1 {
		if strict {
			strs.PushBack("")
		}
	}

	return start
}

// 字符串拼接
func WriteBytes(write io.ByteWriter, bs []byte, off, len int) io.ByteWriter {
	for ; off < len; off++ {
		write.WriteByte(bs[off])
	}

	return write
}

func transArg(sb *strings.Builder, chr byte) {
	switch chr {
	case 't':
		sb.WriteString("\t")
		break
	case 'r':
		sb.WriteString("\r")
		break
	case 'n':
		sb.WriteString("\n")
		break
	case '"':
		sb.WriteByte('"')
		break
	case '\'':
		sb.WriteByte('\'')
		break
	default:
		sb.WriteByte('\\')
		sb.WriteByte(chr)
		break
	}
}

func ToArg(str string) string {
	str = strings.TrimSpace(str)
	last := len(str)
	if last <= 1 {
		return str
	}

	sb := &strings.Builder{}
	qt := 0
	chr := str[0]
	if chr == '"' {
		qt = 1

	} else {
		sb.WriteByte(chr)
	}

	last--
	trans := false
	for i := 1; i < last; i++ {
		chr = str[i]
		if trans {
			trans = false
			transArg(sb, chr)

		} else if chr == '\\' {
			trans = true

		} else {
			sb.WriteByte(chr)
		}
	}

	chr = str[last]
	if trans {
		transArg(sb, chr)

	} else {
		// "" quotation
		if qt == 1 && chr == '"' {
			qt = 2

		} else {
			sb.WriteByte(chr)
		}
	}

	if qt == 1 {
		return "\"" + sb.String()
	}

	return sb.String()
}

func ToArgs(str string) *list.List {
	lenS := len(str)
	if lenS <= 0 {
		return nil
	}

	lst := list.New()

	var chr byte
	qt := false
	trans := false
	s := 0
	for i := 0; i < lenS; i++ {
		chr = str[i]
		if qt {
			if trans {
				trans = false

			} else if chr == '/' {
				trans = true

			} else if chr == '"' {
				lst.PushBack(str[s:i])
				qt = false
				s = i
			}

		} else {
			if chr == '"' {
				qt = true
				if s < i {
					lst.PushBack(str[s:i])
					s = i
				}

			} else if chr == ' ' {
				if s < i {
					lst.PushBack(str[s:i])
				}

				s = i
			}
		}
	}

	if s < lenS {
		if qt && s > 0 {
			lst.PushBack(str[s-1 : lenS])

		} else {
			lst.PushBack(str[s:lenS])
		}
	}

	return lst
}

func Quote(str string) string {
	lenS := len(str)
	if lenS <= 0 {
		return str
	}

	sb := &strings.Builder{}
	sb.WriteByte('"')

	var chr byte
	for i := 0; i < lenS; i++ {
		chr = str[i]
		if chr == '\\' || chr == '"' {
			sb.WriteByte('\\')
		}

		sb.WriteByte(chr)
	}

	sb.WriteByte('"')
	return sb.String()
}

type Matcher struct {
	Typ    int8
	Match  string
	matchO interface{}
}

func (m *Matcher) cover(to *Matcher) bool {
	if to == nil || m.Typ == 0 {
		return true
	}

	if to.Typ == m.Typ {
		switch to.Typ {
		case 1:
			return strings.HasPrefix(to.Match, m.Match)
		case 2:
			return strings.HasSuffix(to.Match, m.Match)
		case 3:
			return strings.Index(to.Match, m.Match) >= 0
		default:
			return to.Match == m.Match
		}
	}

	return false
}

func (m *Matcher) match(str string) bool {
	if m.matchO == nil {
		m.matchO = mathO(m.Typ, m.Match, m.matchO)
	}

	return matchStr(str, m.Typ, m.Match, m.matchO)
}

func mathO(typ int8, match string, matchO interface{}) interface{} {
	if matchO == nil && match != "" {
		switch typ {
		case 4:
			var err error
			matchO, _ = regexp.Compile(match[1:])
			if err != nil {
				Kt.Err(err, true)
			}
			break
		case 5:
			matchO = SplitByte(match, '*', true, 0, 0)
			break
		default:
			matchO = match
		}
	}

	return matchO
}

func matchStr(str string, typ int8, match string, matchO interface{}) bool {
	if len(str) == 0 {
		return typ == 0
	}

	switch typ {
	case 0:
		return true
	case 1:
		return strings.HasSuffix(str, match)
	case 2:
		return strings.HasPrefix(str, match)
	case 3:
		return strings.Index(str, match) >= 0
	case 4:
		return matchO != nil && matchO.(*regexp.Regexp).MatchString(str)
	case 5:
		return MatchSplits(str, matchO.([]string))
	case 9:
		return false
	default:
		return str == match
	}
}

func MatchSplits(str string, splits []string) bool {
	if splits == nil || len(splits) == 0 {
		return true
	}

	if str == "" {
		return false
	}

	i := 0
	for _, split := range splits {
		if len(split) == 0 {
			continue
		}

		i = Index(str, split, i)
		if i < 0 {
			return false
		}

		i++
	}

	return true
}

var AllMatcher = &Matcher{0, "", nil}

func matcherMs(exp string, reg bool, m bool, str string) *Matcher {
	if exp == "" {
		return nil
	}

	if m && str == "" {
		return nil
	}

	var match string
	var matchO interface{}
	var typ int8 = -1
	for {
		if typ != -1 {
			break
		}

		typ = 0
		last := len(exp) - 1
		if last >= 0 {
			if reg && exp[0] == '$' {
				matchO = mathO(4, match, nil)
				if matchO != nil {
					typ = 4
					break
				}
			}

			ms := CountByte(exp, '*', 0, 3)
			if exp[0] == '*' {
				if last == 0 {
					break
				}

				if exp[last] == '*' {
					if last == 1 {
						break
					}

					if ms == 2 {
						match = exp[1:last]
						typ = 3
						break
					}

				} else if ms == 1 {
					match = exp[1:]
					typ = 1
					break
				}

			} else if ms == 1 && exp[last] == '*' {
				match = exp[0:last]
				typ = 2
				break
			}

			if ms > 0 {
				// 多星号匹配
				match = str
				matchO = SplitByte(exp, '*', false, 0, 0)
				typ = 5
				break
			}

		} else {
			typ = 9
			break
		}

		match = exp
		typ = 8
		break
	}

	if m {
		if matchStr(str, typ, match, matchO) {
			return AllMatcher
		}

		return nil

	} else {
		if typ == 0 {
			return AllMatcher
		}

		return &Matcher{typ, match, matchO}
	}
}

func Match(exp string, reg bool, str string) bool {
	return matcherMs(exp, reg, true, str) != nil
}

func ForMatcher(exp string, reg bool) *Matcher {
	return matcherMs(exp, reg, false, "")
}
