package main

import (
	"axj/APro"
	"axj/Kt"
	"axj/KtCvt"
	"axj/KtJson"
	"axjGW/pkg/gateway"
	"context"
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
	Kt.Err(err, true)
	server := thrift.NewTSimpleServer4(processor, transport, thrift.NewTTransportFactory(), thrift.NewTCompactProtocolFactoryConf(nil))
	go func() {
		err := server.Serve()
		Kt.Err(err, true)
	}()

	go func() {
		time.Sleep(1 * time.Second)
		transport, err := thrift.NewTSocketConf("127.0.0.1:8181", nil)
		Kt.Err(err, true)
		err = transport.Open()
		Kt.Err(err, true)
		proto := thrift.NewTCompactProtocolConf(transport, nil)
		client := thrift.NewTStandardClient(proto, proto)
		pass := gw.NewGatewayClient(client)
		r, err := pass.Req(context.Background(), 0, "123", "abc", nil)
		Kt.Err(err, true)
		Kt.Log(r)
	}()

	APro.Signal()
}
