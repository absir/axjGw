package Util

import (
	"axj/Thrd/AZap"
	"go.uber.org/zap"
)

// 协程池接口
type goPool interface {
	Submit(task func()) error
}

// 协程池默认
var GoPool goPool

func GoSubmit(task func()) {
	GoPoolSubmit(GoPool, task)
}

func GoPoolSubmit(pool goPool, task func()) {
	if pool != nil {
		err := pool.Submit(task)
		if err == nil {
			return
		}

		AZap.Logger.Warn("GoPoolSubmit err", zap.Error(err))
	}

	go task()
}
