package APro

import "sync"

type PoolG interface {
	Add()
	Done()
	Wait()
	StrictAs(limit int) bool
}

type PoolLimit struct {
	limit  int
	add    int
	locker sync.Locker
	cond   *sync.Cond
}

func (that PoolLimit) Add() {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.add >= that.limit {
		that.cond.Wait()
	}

	that.add++
}

func (that PoolLimit) Done() {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.add--
	if that.add <= 0 {
		that.cond.Signal()
	}
}

func (that PoolLimit) Wait() {
	that.locker.Lock()
	defer that.locker.Unlock()
	for {
		if that.add > 0 {
			that.cond.Wait()
		}

		break
	}
}

func (that PoolLimit) StrictAs(limit int) bool {
	return that.add == 0 && that.limit == limit
}

func NewPoolLimit(limit int) *PoolLimit {
	pl := new(PoolLimit)
	pl.limit = limit
	pl.add = 0
	pl.locker = new(sync.Mutex)
	pl.cond = sync.NewCond(pl.locker)
	return pl
}

var PoolOne = NewPoolLimit(1)
