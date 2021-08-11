package main

import (
	"axj/APro"
	"axj/KtCvt"
	"axj/KtJson"
	"fmt"
	"runtime"
)

func main() {
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
	fmt.Println(APro.Path())
	cfg := APro.Cfg(nil, "config.properties")
	fmt.Println(KtJson.ToJsonStr(KtCvt.Safe(cfg.Map())))
}
