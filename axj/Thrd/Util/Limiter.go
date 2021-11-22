package Util

import "sync"

type Limiter interface {
	Add()
	Done()
	Wait()
	Limit() int
	StrictAs(limit int) bool
}

type LimiterLocker struct {
	limit  int
	add    int
	locker sync.Locker
	cond   *sync.Cond
}

func (that *LimiterLocker) Lock() sync.Locker {
	return that.locker
}

func (that *LimiterLocker) Add() {
	that.locker.Lock()
	if that.add >= that.limit {
		that.cond.Wait()
	}

	that.add++
	that.locker.Unlock()
}

func (that *LimiterLocker) Done() {
	that.locker.Lock()
	that.add--
	if that.add <= 0 {
		that.cond.Signal()
	}

	that.locker.Unlock()
}

func (that *LimiterLocker) Wait() {
	that.locker.Lock()
	for {
		if that.add > 0 {
			that.cond.Wait()
		}

		break
	}

	that.locker.Unlock()
}

func (that *LimiterLocker) Limit() int {
	return that.limit
}

func (that *LimiterLocker) StrictAs(limit int) bool {
	return that.add == 0 && that.limit == limit
}

func NewLimiterLocker(limit int, locker sync.Locker) *LimiterLocker {
	that := new(LimiterLocker)
	that.limit = limit
	that.add = 0
	if locker == nil {
		locker = new(sync.Mutex)
	}

	that.locker = locker
	that.cond = sync.NewCond(that.locker)
	return that
}
