package gateway

import (
	"axj/APro"
	"github.com/apache/thrift/lib/go/thrift"
)

type config struct {
	id          int32 // 服务编号
	gwProd      string
	accProd     string
	TConfig     *thrift.TConfiguration
	CompressMin int
	DataMax     int32
}

var Config = config{}

func init() {
	Config.gwProd = "gw"
	Config.accProd = "acc"
	Config.TConfig = nil
	Config.CompressMin = 1024
	Config.DataMax = 1024 << 10
	APro.SubCfgBind("gateway", Config)
}
