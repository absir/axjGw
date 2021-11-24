package Kt

import (
	"container/list"
)

type Map interface {
	Range(fun func(key interface{}, val interface{}) bool)
	Get(key interface{}) interface{}
	Put(key interface{}, val interface{}) interface{}
	Remove(key interface{}) interface{}
}

type LinkedEl struct {
	val interface{}
	el  *list.Element
}

func (that *LinkedEl) Get() interface{} {
	return that.val
}

type LinkedMap struct {
	lst *list.List
	mp  map[interface{}]*LinkedEl
}

func (that *LinkedMap) Init() *LinkedMap {
	that.lst = list.New()
	that.mp = map[interface{}]*LinkedEl{}
	return that
}

func (that *LinkedMap) Val() interface{} {
	return that.mp
}

func (that *LinkedMap) Front() *list.Element {
	return that.lst.Front()
}

func (that *LinkedMap) Range(fun func(key interface{}, val interface{}) bool) {
	if fun == nil {
		return
	}

	for front := that.lst.Front(); front != nil; front = front.Next() {
		key := front.Value
		el := that.mp[front.Value]
		if el == nil {
			rm := front
			front = front.Next()
			that.lst.Remove(rm)

		} else {
			if !fun(key, el.val) {
				break
			}
		}
	}
}

func (that *LinkedMap) Get(key interface{}) interface{} {
	val, _ := that.GetC(key)
	return val
}

func (that *LinkedMap) Has(key interface{}) bool {
	_, has := that.GetC(key)
	return has
}

func (that *LinkedMap) GetC(key interface{}) (interface{}, bool) {
	el := that.mp[key]
	if el == nil {
		return nil, false
	}

	return el.val, true
}

func (that *LinkedMap) GetVal(val interface{}, equals Equals) (interface{}, bool) {
	for k, v := range that.mp {
		if IsEquals(val, v.val, equals) {
			return k, true
		}
	}

	return nil, false
}

func (that *LinkedMap) Put(key interface{}, val interface{}) interface{} {
	el := that.mp[key]
	if el == nil {
		el = new(LinkedEl)
		el.val = val
		el.el = that.lst.PushBack(key)
		that.mp[key] = el
		return nil
	}

	_val := el.val
	el.val = val
	that.lst.MoveToBack(el.el)
	return _val
}

func (that *LinkedMap) Remove(key interface{}) interface{} {
	el := that.mp[key]
	if el == nil {
		return nil
	}

	delete(that.mp, key)
	that.lst.Remove(el.el)
	return el.val
}

func (that *LinkedMap) Clear() {
	that.lst.Init()
	that.mp = map[interface{}]*LinkedEl{}
}

func (that *LinkedMap) IsEmpty() bool {
	return that.lst.Front() == nil
}
