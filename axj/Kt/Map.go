package Kt

import (
	"container/list"
)

type Map interface {
	Get(key interface{}) interface{}
	Put(key interface{}, val interface{}) interface{}
	Remove(key interface{}) interface{}
}

type LinkedMap struct {
	lst *list.List
	mp  map[interface{}]*list.Element
}

func (l *LinkedMap) Init() *LinkedMap {
	l.lst = list.New()
	l.mp = map[interface{}]*list.Element{}
	return l
}

func (l *LinkedMap) Front() *list.Element {
	return l.lst.Front()
}

func (l *LinkedMap) Get(key interface{}) interface{} {
	val, _ := l.GetC(key)
	return val
}

func (l *LinkedMap) Has(key interface{}) bool {
	_, has := l.GetC(key)
	return has
}

func (l *LinkedMap) GetC(key interface{}) (interface{}, bool) {
	el := l.mp[key]
	if el == nil {
		return nil, false
	}

	return el.Value, true
}

func (l *LinkedMap) GetVal(val interface{}, equals Equals) (interface{}, bool) {
	for k, v := range l.mp {
		if IsEquals(val, v.Value, equals) {
			return k, true
		}
	}

	return nil, false
}

func (l *LinkedMap) Put(key interface{}, val interface{}) interface{} {
	el := l.mp[key]
	if el == nil {
		l.mp[key] = l.lst.PushBack(key)
		return nil
	}

	_val := el.Value
	el.Value = val
	l.lst.MoveToBack(el)
	return _val
}

func (l *LinkedMap) Remove(key interface{}) interface{} {
	el := l.mp[key]
	if el == nil {
		return nil
	}

	delete(l.mp, key)
	l.lst.Remove(el)
	return el.Value
}

func (l *LinkedMap) Clear() {
	l.lst.Init()
	l.mp = map[interface{}]*list.Element{}
}

func (l *LinkedMap) IsEmpty() bool {
	return l.lst.Front() == nil
}
