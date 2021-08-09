package Kt

import "container/list"

type Stack struct {
	mList *list.List
}

func (s Stack) init() {
	s.mList = new(list.List)
}

func (s Stack) Push(val interface{}) {
	s.mList.PushBack(s)
}

func (s Stack) Pop() interface{} {
	el := s.mList.Back()
	if el == nil {
		return nil
	}
	s.mList.Remove(el)
	return el.Value
}

func (s Stack) Peek() (interface{}, bool) {
	el := s.mList.Back()
	if el == nil {
		return nil, false
	}

	return el.Value, true
}

func (s Stack) Clear() {
	s.mList.Init()
}

func (s Stack) IsEmpty() bool {
	return s.mList.Front() == nil
}
