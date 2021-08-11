package gateway

import (
	"github.com/apache/thrift/lib/go/thrift"
	"gw"
	"strings"
)

type Pass struct {
	// 转发地址
	url string
	// 客户端
	client thrift.TClient
	// 转发客户端
	passClient *gw.PassClient
}

func (p Pass) init() {
	thrift.NewTTransportFactory()
	var client thrift.TTransport
	var err error

	if strings.HasPrefix(p.url, "http") {
		client, err = thrift.NewTHttpClient(p.url)

	} else {
		client, err = thrift.NewTSocketConf(p.url, nil)
	}

	if err != nil {
		// todo
	}

	proto := thrift.NewTCompactProtocolConf(client, nil)
	p.client = thrift.NewTStandardClient(proto, proto)
	p.passClient = gw.NewPassClient(p.client)
}

type Passes struct {
	// 转发列表
	passes []Pass
}

type PassesReg struct {
	// 转发地图
	passesMap map[string]Passes
}

var passesMap = map[string]Passes{}

func reg(name string, url string) {

}
