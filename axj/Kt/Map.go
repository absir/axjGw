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
	mList *list.List
	mMap  map[interface{}]*list.Element
}

func (l LinkedMap) init() {
	l.mList = new(list.List)
}

func (l LinkedMap) Front() *list.Element {
	return l.mList.Front()
}

func (l LinkedMap) Get(key interface{}) interface{} {
	val, _ := l.GetC(key)
	return val
}

func (l LinkedMap) Has(key interface{}) bool {
	_, has := l.GetC(key)
	return has
}

func (l LinkedMap) GetC(key interface{}) (interface{}, bool) {
	el := l.mMap[key]
	if el == nil {
		return nil, false
	}

	return el.Value, true
}

func (l LinkedMap) GetVal(val interface{}, equals Equals) (interface{}, bool) {
	for k, v := range l.mMap {
		if IsEquals(val, v.Value, equals) {
			return k, true
		}
	}

	return nil, false
}

func (l LinkedMap) Put(key interface{}, val interface{}) interface{} {
	el := l.mMap[key]
	if el == nil {
		l.mMap[key] = l.mList.PushBack(key)
		return nil
	}

	_val := el.Value
	el.Value = val
	l.mList.MoveToBack(el)
	return _val
}

func (l LinkedMap) Remove(key interface{}) interface{} {
	el := l.mMap[key]
	if el == nil {
		return nil
	}

	delete(l.mMap, key)
	l.mList.Remove(el)
	return el.Value
}

func (l LinkedMap) Clear() {
	l.mList.Init()
	l.mMap = *new(map[interface{}]*list.Element)
}

func (l LinkedMap) IsEmpty() bool {
	return l.mList.Front() == nil
}
