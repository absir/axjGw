package Kt

import (
	"container/list"
	"fmt"
	"hash/crc32"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
	"time"
)

const (
	Develop int8 = 0
	Debug   int8 = 1
	Test    int8 = 2
	Product int8 = 3
)

type IVal interface {
	Get() interface{}
}

var Env = Develop

var Active = true

var Started = false

var LogOut = log.New(os.Stdout, "", log.LstdFlags)

var LogErr = log.New(os.Stderr, "", log.LstdFlags)

// 日志
func Log(v ...interface{}) {
	LogOut.Output(2, fmt.Sprintln(v...))
}

// 信息日志
func Info(info ...interface{}) {
	if info == nil {
		return
	}

	last := len(info)
	if last <= 0 {
		return
	}

	if last == 1 {
		LogOut.Output(2, time.Now().Format(time.RFC3339)+" Info "+fmt.Sprintln(info[0]))
		return
	}

	// 多个参数打印
	LogOut.Output(2, time.Now().Format(time.RFC3339)+" Info "+fmt.Sprint(info[0]))
	for i := 1; i < last; i++ {
		LogOut.Output(2, ", ")
		LogOut.Output(2, fmt.Sprint(info[i]))
	}

	LogOut.Output(2, ", ")
	LogOut.Output(2, fmt.Sprintln(info[last]))
}

// 错误提示
func Err(err error, stack bool) {
	if err == nil {
		return
	}

	LogOut.Output(2, fmt.Sprintln(err))
	if stack {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			fun := runtime.FuncForPC(pc)
			LogOut.Output(2, fmt.Sprintf("\tat %s:%d (Method %s)\nCause by: %s\n", file, line, fun.Name(), err.Error()))

		} else {
			LogOut.Output(2, "get call stack err ...\n")
		}
	}
}

func Error(info string, err error) {
	ErrorS(info, err, false)
}

func ErrorS(info string, err error, stack bool) {
	if info != "" {
		LogOut.Output(2, time.Now().Format(time.RFC3339)+" Error "+info+"\n")
	}

	Err(err, stack)
}

func Panic(err error) {
	if err == nil {
		return
	}

	panic(err)
}

// 三元表达式
func If(a bool, b, c interface{}) interface{} {
	if a {
		return b
	}

	return c
}

func Min(a, b, c int) int {
	if a < b {
		if c < a {
			return c
		}

		return a
	}

	if c < b {
		return c
	}

	return b
}

// 等于方法
type Equals func(interface{}, interface{}) bool

func IsEquals(from, to interface{}, equals Equals) bool {
	return from == to || (equals != nil && equals(from, to))
}

// 数据查找
func IndexOf(array []interface{}, el interface{}, equals Equals) int {
	aLen := len(array)
	for i := 0; i < aLen; i++ {
		if IsEquals(el, array[i], equals) {
			return i
		}
	}

	return -1
}

// 转数组
func ToArray(lst *list.List) []interface{} {
	if lst == nil {
		return nil
	}

	array := make([]interface{}, lst.Len())
	i := 0
	for el := lst.Front(); el != nil; el = el.Next() {
		array[i] = el.Value
		i++
	}

	return array
}

// 转列表
func ToList(array []interface{}) *list.List {
	if array == nil {
		return nil
	}

	lst := list.New()
	lenA := len(array)
	for i := 0; i < lenA; i++ {
		lst.PushBack(array[i])
	}

	return lst
}

func HashCode(bs []byte) int {
	v := int(crc32.ChecksumIEEE(bs))
	if v >= 0 {
		return v
	}

	if -v >= 0 {
		return -v
	}

	return 0
}

func IpAddr(addr net.Addr) string {
	if tAddr, ok := addr.(*net.TCPAddr); ok {
		if tAddr.Zone != "" {
			return tAddr.Zone
		}

		return tAddr.IP.String()
	}

	if uAddr, ok := addr.(*net.UDPAddr); ok {
		if uAddr.Zone != "" {
			return uAddr.Zone
		}

		return uAddr.IP.String()
	}

	return IpAddrStr(addr.String())
}

func IpAddrStr(str string) string {
	idx := strings.IndexByte(str, ':')
	if idx >= 0 {
		return str[:idx]
	}

	return str
}

func PrintStacks() {
	var buf [2 << 10]byte
	fmt.Println(string(buf[:runtime.Stack(buf[:], true)]))
}

type ErrReason struct {
	reason string
}

func (e ErrReason) Error() string {
	return e.reason
}

func NewErrReason(reason string) error {
	that := new(ErrReason)
	that.reason = reason
	return that
}
