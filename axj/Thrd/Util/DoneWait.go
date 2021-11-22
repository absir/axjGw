package Util

import "sync"

type DoneWait struct {
	wait   int
	locker sync.Locker
	cond   *sync.Cond
}

func (that *DoneWait) Add() {
	that.locker.Lock()
	that.wait++
	that.locker.Unlock()
}

func (that *DoneWait) Done() {
	that.locker.Lock()
	that.wait--
	if that.wait <= 0 {
		that.cond.Signal()
	}

	that.locker.Unlock()
}

func (that *DoneWait) Wait() {
	that.locker.Lock()
	if that.wait > 0 {
		that.cond.Wait()
	}

	that.locker.Unlock()
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
