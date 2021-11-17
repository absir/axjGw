package Util

import (
	"sync"
	"time"
)

type NotifierAsync struct {
	run     func()
	locker  sync.Locker
	runTime int64
	running bool
}

func NewNotifierAsync(run func(), locker sync.Locker) *NotifierAsync {
	that := new(NotifierAsync)
	that.run = run
	if locker == nil {
		locker = new(sync.Mutex)
	}

	that.locker = locker
	return that
}

func (that *NotifierAsync) Start(run func()) {
	if run == nil && that.run == nil {
		return
	}

	runTime := time.Now().UnixNano()
	that.locker.Lock()
	defer that.locker.Unlock()
	if run != nil {
		that.run = run
	}

	if that.runTime < runTime {
		that.runTime = runTime

	} else {
		that.runTime++
	}

	if !that.running {
		go that.runDo()
	}
}

func (that *NotifierAsync) runIn() bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.running {
		return false
	}

	that.running = true
	return true
}

func (that *NotifierAsync) runOut(runTime int64) {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.running = false
	if that.runTime > runTime {
		go that.runDo()
	}
}

func (that *NotifierAsync) runDone(runTime int64) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	return that.runTime <= runTime
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
