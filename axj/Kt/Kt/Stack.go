package Kt

import "container/list"

type Stack struct {
	lst *list.List
}

func (s *Stack) Init() *Stack {
	s.lst = list.New()
	return s
}

func (s *Stack) Push(val interface{}) {
	s.lst.PushBack(s)
}

func (s *Stack) Pop() interface{} {
	el := s.lst.Back()
	if el == nil {
		return nil
	}
	s.lst.Remove(el)
	return el.Value
}

func (s *Stack) Peek() (interface{}, bool) {
	el := s.lst.Back()
	if el == nil {
		return nil, false
	}

	return el.Value, true
}

func (s *Stack) Clear() {
	s.lst.Init()
}

func (s *Stack) IsEmpty() bool {
	return s.lst.Front() == nil
}
