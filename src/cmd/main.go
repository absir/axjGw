package main

import (
	"axj/APro"
	"axj/Kt"
	"axj/KtCvt"
	"axj/KtJson"
	"axjGW/pkg/gateway"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"gw"
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

	processor := gw.NewGatewayProcessor(gateway.Remote{})
	transport, err := thrift.NewTServerSocket("0.0.0.0:8181")
	Kt.Err(err)
	server := thrift.NewTSimpleServer2(processor, transport)
	go func() {
		err := server.Serve()
		Kt.Err(err)
	}()

	go func() {
		time.Sleep(5 * time.Second)
		transport, err := thrift.NewTSocketConf("127.0.0.1:8181", nil)
		Kt.Err(err)
		proto := thrift.NewTCompactProtocolConf(transport, nil)
		client := thrift.NewTStandardClient(proto, proto)
		pass := gw.NewGatewayClient(client)
		r, err := pass.Req(nil, 0, "", "", nil)
		Kt.Err(err)
		Kt.Log(r)
	}()

	APro.Signal()
}
