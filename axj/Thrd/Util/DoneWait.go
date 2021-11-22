package Util

import "sync"

type DoneWait struct {
	wait   int
	locker sync.Locker
	cond   *sync.Cond
}

func (that *DoneWait) Add() {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.wait++
}

func (that *DoneWait) Done() {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.wait--
	if that.wait <= 0 {
		that.cond.Signal()
	}
}

func (that *DoneWait) Wait() {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.wait > 0 {
		that.cond.Wait()
	}
}

func NewWaitDone(locker sync.Locker) *DoneWait {
	that := new(DoneWait)
	if locker == nil {
		locker = new(sync.Mutex)
	}

	that.locker = locker
	that.cond = sync.NewCond(that.locker)
	return that
}
