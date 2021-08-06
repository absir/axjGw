package Kt

import (
	"container/list"
	"os"
	"os/signal"
	"syscall"
)

// 三元表达式
func If(a bool, b, c interface{}) interface{} {
	if a {
		return b
	}

	return c
}

// 等于方法
type Equals func(interface{}, interface{}) bool

// 数据查找
func IndexOf(array []interface{}, el interface{}, equals Equals) int {
	len := len(array)
	for i := 0; i < len; i++ {
		t := array[i]
		if t == el || (equals != nil && equals(t, el)) {
			return i
		}
	}

	return -1
}

// 转数组
func ToArray(list *list.List) []interface{} {
	if list == nil {
		return nil
	}

	array := make([]interface{}, list.Len())
	i := 0
	for el := list.Front(); el != nil; el = el.Next() {
		array[i] = el.Value
		i++
	}

	return array
}

// 转列表
func toList(array []interface{}) *list.List {
	if array == nil {
		return nil
	}

	list := list.New()
	len := len(array)
	for i := 0; i < len; i++ {
		list.PushBack(array[i])
	}

	return list
}

// 关闭信号
func Signal() os.Signal {
	c := make(chan os.Signal, 0)
	signal.Notify(c, syscall.SIGTERM)
	return <-c
}
