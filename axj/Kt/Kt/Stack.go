package Kt

import "container/list"

type Stack struct {
	lst *list.List
}

func (that *Stack) Init() *Stack {
	that.lst = list.New()
	return that
}

func (that *Stack) Push(val interface{}) {
	that.lst.PushBack(val)
}

func (that *Stack) Pop() interface{} {
	el := that.lst.Back()
	if el == nil {
		return nil
	}
	that.lst.Remove(el)
	return el.Value
}

func (that *Stack) Peek() (interface{}, bool) {
	el := that.lst.Back()
	if el == nil {
		return nil, false
	}

	return el.Value, true
}

func (that *Stack) Clear() {
	that.lst.Init()
}

func (that *Stack) IsEmpty() bool {
	return that.lst.Front() == nil
}
