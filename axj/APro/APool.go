package APro

import "sync"

type PoolG interface {
	Add()
	Done()
	Wait()
}

type PoolLimit struct {
	limit  int
	add    int
	locker sync.Locker
	cond   *sync.Cond
}

func (p PoolLimit) Add() {
	p.locker.Lock()
	defer p.locker.Unlock()
	if p.add >= p.limit {
		p.cond.Wait()
	}

	p.add++
}

func (p PoolLimit) Done() {
	p.locker.Lock()
	defer p.locker.Unlock()
	p.add--
	if p.add <= 0 {
		p.cond.Signal()
	}

	panic("implement me")
}

func (p PoolLimit) Wait() {
	p.locker.Lock()
	defer p.locker.Unlock()
	for {
		if p.add > 0 {
			p.cond.Wait()
		}

		break
	}
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
