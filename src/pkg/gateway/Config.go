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

var Config *config

func init() {
	Config = &config{
		GwProd:      "gw",
		AclProd:     "acl",
		TConfig:     nil,
		CompressMin: 1024,
		DataMax:     1024 << 10,
	}
	APro.SubCfgBind("gateway", Config)
}
