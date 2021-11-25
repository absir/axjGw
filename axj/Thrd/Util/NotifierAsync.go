package Util

import (
	"sync"
	"time"
)

type NotifierAsync struct {
	run     func()
	locker  sync.Locker
	runCond func() bool
	runTime int64
	running bool
}

func NewNotifierAsync(run func(), locker sync.Locker, runCond func() bool) *NotifierAsync {
	that := new(NotifierAsync)
	that.run = run
	if locker == nil {
		locker = new(sync.Mutex)
	}

	that.locker = locker
	that.runCond = runCond
	return that
}

func (that *NotifierAsync) Start(run func()) {
	that.StartLock(run, true)
}

func (that *NotifierAsync) StartLock(run func(), lock bool) {
	// run nil 拦截
	if run == nil && that.run == nil {
		return
	}

	// 运行时间
	runTime := time.Now().UnixNano()
	if lock {
		// 加锁
		that.locker.Lock()
	}

	// run函数判断
	if run != nil {
		that.run = run
	}

	// 保证runTime递增
	if that.runTime < runTime {
		that.runTime = runTime

	} else {
		that.runTime++
	}

	if !that.running {
		if that.runCond == nil || that.runCond() {
			GoSubmit(that.runDo)
		}
	}

	if lock {
		// 解锁
		that.locker.Unlock()
	}
}

func (that *NotifierAsync) runIn() bool {
	that.locker.Lock()
	if that.running {
		that.locker.Unlock()
		return false
	}

	that.running = true
	that.locker.Unlock()
	return true
}

func (that *NotifierAsync) runOut(runTime int64) {
	that.locker.Lock()
	that.running = false
	if that.runTime > runTime {
		if that.runCond == nil || that.runCond() {
			GoSubmit(that.runDo)
		}

		that.locker.Unlock()

	} else {
		that.locker.Unlock()
	}
}

func (that *NotifierAsync) runDone(runTime int64) bool {
	that.locker.Lock()
	done := that.runTime <= runTime
	that.locker.Unlock()
	return done
}

func (that *NotifierAsync) runDo() {
	if !that.runIn() {
		return
	}

	runTime := that.runTime
	defer that.runOut(runTime)
	for {
		runTime = that.runTime
		that.run()
		if that.runDone(runTime) {
			break
		}
	}
}
