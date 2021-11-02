package Util

import (
	"axj/Thrd/AZap"
	"go.uber.org/zap"
	"sync"
)

type Pool interface {
	PInit()
	PRelease() bool
}

type AllocPool struct {
	new  func() interface{}
	pool *sync.Pool
}

func NewAllocPool(pool bool, alloc func() Pool) *AllocPool {
	that := new(AllocPool)
	that.new = func() interface{} {
		p := alloc()
		p.PInit()
		return p
	}

	that.SetPool(pool)
	return that
}

func (that *AllocPool) SetPool(pool bool) {
	if pool {
		that.pool = new(sync.Pool)
		that.pool.New = that.new

	} else {
		that.pool = nil
	}
}

func (that *AllocPool) Get() interface{} {
	if that.pool == nil {
		return that.new()
	}

	return that.pool.Get()
}

func (that *AllocPool) Put(p Pool, recover bool) {
	if that.pool != nil {
		if recover {
			defer that.recover()
		}

		if p.PRelease() {
			that.pool.Put(p)
		}
	}
}

func (that *AllocPool) recover() {
	if reason := recover(); reason != nil {
		AZap.Logger.Warn("AllocPool recover", zap.Reflect("reason", reason))
	}
}
