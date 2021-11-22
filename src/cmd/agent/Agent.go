package main

import (
	"axj/APro"
	"axj/Thrd/cmap"
	"runtime"
)

type Config struct {
	Proxy     string // 代理地址
	ClientKey string // 客户端Key
}

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../agent")
	APro.Load(nil, "config.ini")
	APro.Load(nil, "config.yml")

	cmap := cmap.NewCMapInit()
	for i := 0; i < 1000; i++ {
		cmap.Store(i, i)
		println(cmap.SizeBuckets())
	}
}
