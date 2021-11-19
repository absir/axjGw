package main

import (
	"axj/APro"
	"runtime"
)

type Config struct {
	Proxy string // 代理地址
}

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../agent")
	APro.Load(nil, "config.ini")
	APro.Load(nil, "config.yml")


}
