package Util

import "axj/Kt/Kt"

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
		Kt.Panic(err)
		return
	}

	go task()
}
