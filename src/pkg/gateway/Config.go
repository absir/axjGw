package gateway

import (
	"axj/APro"
	"github.com/apache/thrift/lib/go/thrift"
)

type config struct {
	GwProd      string
	AclProd     string
	TConfig     *thrift.TConfiguration
	CompressMin int
	DataMax     int32
}

var Config = config{}

func init() {
	Config.GwProd = "gw"
	Config.AclProd = "acl"
	Config.TConfig = nil
	Config.CompressMin = 1024
	Config.DataMax = 1024 << 10
	APro.SubCfgBind("gateway", Config)
}
