package gateway

import (
	"axj/APro"
	"github.com/apache/thrift/lib/go/thrift"
	"time"
)

type config struct {
	WorkId      int32
	WorkHash    int
	CompressMin int                    // 最短压缩
	DataMax     int32                  // 最大数据(请求)
	CheckDrt    int64                  // 客户端检查间隔
	IdleDrt     int64                  // 空闲检测间隔
	ConnDrt     int64                  // 连接检查间隔
	KickDrt     time.Duration          // 踢出间隔
	GwProd      string                 // 网关服务名
	AclProd     string                 // Acl服务名
	TConfig     *thrift.TConfiguration // thrift server 配置
}

var Config *config

func initConfig(workId int32) {
	Config = &config{
		CompressMin: 1024,
		DataMax:     1024 << 10,
		CheckDrt:    3000,
		IdleDrt:     30000,
		ConnDrt:     60000,
		KickDrt:     6000,
		GwProd:      "gw",
		AclProd:     "acl",
		TConfig:     nil,
	}
	APro.SubCfgBind("gateway", Config)
	Config.WorkId = workId
	Config.WorkHash = int(workId)
}
