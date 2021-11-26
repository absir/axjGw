package Util

import (
	"container/list"
	"sync"
)

type ListChan struct {
	size   int
	list   *list.List
	locker sync.Locker
	cond   *sync.Cond
	closed bool
}

func NewListChan(locker sync.Locker) *ListChan {
	that := new(ListChan)
	that.list = new(list.List)
	if locker == nil {
		locker = new(sync.Mutex)
	}

	that.locker = locker
	that.cond = sync.NewCond(locker)
	return that
}

func (that *ListChan) Size() int {
	return that.size
}

func (that *ListChan) Close(lock bool) {
	if that.closed {
		return
	}

	if lock {
		that.locker.Lock()
	}

	if !that.closed {
		that.closed = true
		that.cond.Signal()
	}

	if lock {
		that.locker.Unlock()
	}
}

func (that *ListChan) Push(data interface{}, lock bool) {
	if lock {
		that.locker.Lock()
	}

	that.list.PushBack(data)
	that.size++
	that.cond.Signal()
	if lock {
		that.locker.Unlock()
	}
}

func (that *ListChan) Peek(lock bool) interface{} {
	for {
		if lock {
			that.locker.Lock()
		}

		el := that.list.Front()
		if el == nil {
			that.cond.Wait()
			if lock {
				that.locker.Unlock()
			}

		} else {
			that.size--
			that.list.Remove(el)
			if lock {
				that.locker.Unlock()
			}

			return el.Value
		}
	}
}
