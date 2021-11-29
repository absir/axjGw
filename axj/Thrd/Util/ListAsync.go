package Util

import (
	"container/list"
	"sync"
)

type ListAsync struct {
	size    int
	list    *list.List
	run     func(interface{})
	locker  sync.Locker
	running bool
}

func NewListAsync(run func(interface{}), locker sync.Locker) *ListAsync {
	that := new(ListAsync)
	that.list = new(list.List)
	that.run = run
	if locker == nil {
		locker = new(sync.Mutex)
	}

	that.locker = locker
	return that
}

func (that *ListAsync) Size() int {
	return that.size
}

func (that *ListAsync) SetRun(run func(interface{})) {
	if run != nil {
		that.run = run
	}
}

func (that *ListAsync) Clear(lock bool) {
	if lock {
		that.locker.Lock()
	}

	if that.list.Front() != nil {
		that.list.Init()
	}

	if lock {
		that.locker.Unlock()
	}
}

func (that *ListAsync) Start() {
	if that.running {
		return
	}

	// 加锁
	that.locker.Lock()
	if !that.running && that.list.Front() != nil {
		GoSubmit(that.runDo)
	}

	that.locker.Unlock()
}

func (that *ListAsync) Submit(el interface{}) {
	that.SubmitLock(el, true)
}

func (that *ListAsync) SubmitLock(el interface{}, lock bool) {
	if lock {
		// 加锁
		that.locker.Lock()
	}

	//  添加到尾部
	that.size++
	that.list.PushBack(el)
	if that.run != nil && !that.running {
		GoSubmit(that.runDo)
	}

	if lock {
		// 解锁
		that.locker.Unlock()
	}
}

func (that *ListAsync) runIn() (bool, interface{}) {
	var el interface{}
	that.locker.Lock()
	if that.running {
		that.locker.Unlock()
		return false, nil
	}

	front := that.list.Front()
	if front == nil {
		that.locker.Unlock()
		return false, nil
	}

	that.running = true
	el = front.Value
	that.size--
	that.list.Remove(front)
	that.locker.Unlock()
	return true, el
}

func (that *ListAsync) runOut() {
	that.locker.Lock()
	that.running = false
	if that.size > 0 {
		GoSubmit(that.runDo)
		that.locker.Unlock()

	} else {
		that.locker.Unlock()
	}
}

func (that *ListAsync) runDone() (bool, interface{}) {
	that.locker.Lock()
	front := that.list.Front()
	if front == nil {
		that.locker.Unlock()
		return true, nil
	}

	el := front.Value
	that.size--
	that.list.Remove(front)
	that.locker.Unlock()
	return false, el
}

func (that *ListAsync) runDo() {
	in, el := that.runIn()
	if !in {
		return
	}

	defer that.runOut()
	var done bool
	for {
		that.run(el)
		done, el = that.runDone()
		if done {
			break
		}
	}
}
