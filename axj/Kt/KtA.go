package Kt

import (
	"container/list"
	"fmt"
)

const (
	Develop int8 = 0
	Debug   int8 = 1
	Test    int8 = 2
	Product int8 = 3
)

var Env = Develop

var Active = true

var Started = true

// 错误提示
func Err(err error) {
	if err == nil {
		return
	}

	fmt.Errorf(err.Error())
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
	len := len(array)
	for i := 0; i < len; i++ {
		if IsEquals(el, array[i], equals) {
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
func ToList(array []interface{}) *list.List {
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
