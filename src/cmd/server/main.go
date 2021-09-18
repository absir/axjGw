package main

import (
	"axj/APro"
	Kt2 "axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtJson"
	"axjGW/pkg/gateway"
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"gw"
	"runtime"
	"time"
	"unsafe"
)

type Test struct {
	Name string
	desc string

	Timeout time.Duration
}

type Test2 struct {
	Test
	Name2 string
}

func main() {

	test2 := Test2{}
	test2.Name2 = "abc"
	test := test2.Test
	test2 = *(*Test2)(unsafe.Pointer(&test))
	println(test2.Name2)

	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
	fmt.Println(APro.Path())
	cfg := APro.Load(nil, "config.properties")
	fmt.Println(KtJson.ToJsonStr(KtCvt.Safe(cfg.Map())))


	processor := gw.NewGatewayProcessor(gateway.Remote{})
	transport, err := thrift.NewTServerSocket("0.0.0.0:8181")
	Kt2.Err(err, true)
	server := thrift.NewTSimpleServer4(processor, transport, thrift.NewTTransportFactory(), thrift.NewTCompactProtocolFactoryConf(nil))
	go func() {
		err := server.Serve()
		Kt2.Err(err, true)
	}()

	go func() {
		time.Sleep(1 * time.Second)
		transport, err := thrift.NewTSocketConf("127.0.0.1:8181", nil)
		Kt2.Err(err, true)
		err = transport.Open()
		Kt2.Err(err, true)
		proto := thrift.NewTCompactProtocolConf(transport, nil)
		client := thrift.NewTStandardClient(proto, proto)
		pass := gw.NewGatewayClient(client)
		r, err := pass.Req(context.Background(), 0, "123", "abc", nil)
		Kt2.Err(err, true)
		Kt2.Log(r)
	}()

	APro.Signal()
}
