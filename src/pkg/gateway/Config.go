package gateway

import (
	"axj/APro"
	"time"
)

type config struct {
	WorkId       int32
	WorkHash     int
	CompressMin  int               // 最短压缩
	DataMax      int32             // 最大数据(请求)
	Encrypt      bool              // 通讯加密
	CheckDrt     time.Duration     // 客户端检查间隔
	IdleDrt      int64             // 空闲检测间隔
	ConnDrt      int64             // 连接检查间隔
	KickDrt      time.Duration     // 踢出间隔
	ConnLimit    int               // 连接检查协程限制
	LiveLimit    int               // 存活检查协程限制
	GwProd       string            // 网关服务名
	GwMetas      map[string]string // 网关额外数据
	GwAclReg     time.Duration     // 网关Acl注册
	AclProd      string            // Acl服务名
	PassProd     string            // Pass服务名
	ProdTimeout  time.Duration     // 服务超时时间
	ProdCheckDrt time.Duration     // 服务发现检测时间
	TeamMax      int               // 群组最大缓存
	ClientPMax   int               // 客户端Map遍历最大预增缓冲
	zDevAcl      bool              // 网关服务测试模式
	ProdPort     int               // 微服务默认服务端口
	ProdPortKey  string            // 微服务默认服务端口发现KEY
}

var Config *config

func initConfig(workId int32) {
	Config = &config{
		CompressMin:  256,
		DataMax:      256 << 10,
		Encrypt:      true,
		CheckDrt:     3000,
		IdleDrt:      30000,
		ConnDrt:      60000,
		KickDrt:      10 * time.Second,
		GwProd:       "gw",
		AclProd:      "acl",
		ProdTimeout:  30 * time.Second,
		ProdCheckDrt: 30 * time.Second,
		TeamMax:      65535,
		ClientPMax:   16,
		ProdPort:     8083,
		ProdPortKey:  "gwPort",
	}

	Config.PassProd = Config.AclProd
	APro.SubCfgBind("gateway", Config)
	Config.WorkId = workId
	Config.WorkHash = int(workId)
}
