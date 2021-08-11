package main

import (
	"axj/APro"
	"axj/KtCvt"
	"axj/KtJson"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

type Test struct {
	Name string
	desc string

	Timeout time.Duration
}

func main() {
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
	fmt.Println(APro.Path())
	cfg := APro.Cfg(nil, "config.properties")
	fmt.Println(KtJson.ToJsonStr(KtCvt.Safe(cfg.Map())))
	test := Test{}
	test.Name = "2"
	field := reflect.ValueOf(&test).Elem().FieldByName("Name")
	fmt.Println(field.CanSet())
	KtCvt.BindMap(reflect.ValueOf(&test), cfg.Map())
	fmt.Println(KtJson.ToJsonStr(test))
}
