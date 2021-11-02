package Util

import "sync"

type Limiter interface {
	Add()
	Done()
	Wait()
	StrictAs(limit int) bool
}

type LimiterLocker struct {
	limit  int
	add    int
	locker sync.Locker
	cond   *sync.Cond
}

func (that *LimiterLocker) Add() {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.add >= that.limit {
		that.cond.Wait()
	}

	that.add++
}

func (that *LimiterLocker) Done() {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.add--
	if that.add <= 0 {
		that.cond.Signal()
	}
}

func (that *LimiterLocker) Wait() {
	that.locker.Lock()
	defer that.locker.Unlock()
	for {
		if that.add > 0 {
			that.cond.Wait()
		}

		break
	}
}

func (that *LimiterLocker) StrictAs(limit int) bool {
	return that.add == 0 && that.limit == limit
}

func NewLimiterLocker(limit int) *LimiterLocker {
	pl := new(LimiterLocker)
	pl.limit = limit
	pl.add = 0
	pl.locker = new(sync.Mutex)
	pl.cond = sync.NewCond(pl.locker)
	return pl
}

var LimiterOne = NewLimiterLocker(1)
