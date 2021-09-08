package Kt

import (
	"container/list"
)

type Map interface {
	Get(key interface{}) interface{}
	Put(key interface{}, val interface{}) interface{}
	Remove(key interface{}) interface{}
}

type LinkedEl struct {
	val interface{}
	el  *list.Element
}

func (l LinkedEl) Get() interface{} {
	return l.val
}

type LinkedMap struct {
	lst *list.List
	mp  map[interface{}]*LinkedEl
}

func (l *LinkedMap) Init() *LinkedMap {
	l.lst = list.New()
	l.mp = map[interface{}]*LinkedEl{}
	return l
}

func (l *LinkedMap) Val() interface{} {
	return l.mp
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

	return el.val, true
}

func (l *LinkedMap) GetVal(val interface{}, equals Equals) (interface{}, bool) {
	for k, v := range l.mp {
		if IsEquals(val, v.val, equals) {
			return k, true
		}
	}

	return nil, false
}

func (l *LinkedMap) Put(key interface{}, val interface{}) interface{} {
	el := l.mp[key]
	if el == nil {
		el = new(LinkedEl)
		el.val = val
		el.el = l.lst.PushBack(key)
		l.mp[key] = el
		return nil
	}

	_val := el.val
	el.val = val
	l.lst.MoveToBack(el.el)
	return _val
}

func (l *LinkedMap) Remove(key interface{}) interface{} {
	el := l.mp[key]
	if el == nil {
		return nil
	}

	delete(l.mp, key)
	l.lst.Remove(el.el)
	return el.val
}

func (l *LinkedMap) Clear() {
	l.lst.Init()
	l.mp = map[interface{}]*LinkedEl{}
}

func (l *LinkedMap) IsEmpty() bool {
	return l.lst.Front() == nil
}
