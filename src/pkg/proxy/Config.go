package proxy

import (
	"axj/APro"
	"axj/Kt/KtBytes"
	"axj/Kt/KtStr"
	"context"
	"time"
)

type config struct {
	SocketAddr    string // socket服务地址
	CompressMin   int    // 最短压缩
	DataMax       int32  // 最大数据(请求)
	Encrypt       bool   // 通讯加密
	CheckDrt      time.Duration
	IdleDrt       int64
	KickDrt       time.Duration
	AdapCheckDrt  time.Duration     // 客户端检查间隔
	AdapCheckBuff int               // 客户端Buff
	AdapTimeout   int64             // 适配超时
	AdapMaxId     int32             // 适配最大编号
	CloseDelay    int               // 关闭延迟秒
	DialTimeout   time.Duration     // 连接超时
	Servs         map[string]*Serv  // 协议服务
	ClientKeys    map[string]string // 客户端授权码
	Acl           string            // Acl服务地址
	AclTry        time.Duration     // Acl服务地址
	AclTimeout    time.Duration     // Acl调用超时
	GrpcAddr      string            // grpc服务地址
	GrpcIps       []string          // grpc调用Ip白名单，支持*通配
}

type Serv struct {
	Addr       string
	Proto      string
	TrafficDrt time.Duration
	Cfg        map[string]string
}

var Config *config

func initConfig() {
	Config = &config{
		SocketAddr:    ":8783",
		CompressMin:   256,
		DataMax:       256 << 10,
		Encrypt:       true,
		CheckDrt:      3,
		IdleDrt:       30,
		KickDrt:       6,
		AdapCheckDrt:  3,
		AdapCheckBuff: 128,
		AdapTimeout:   60,
		AdapMaxId:     KtBytes.VINT_3_MAX,
		CloseDelay:    30,
		DialTimeout:   10,
		Servs:         map[string]*Serv{},
		ClientKeys:    map[string]string{},
		AclTry:        3,
		AclTimeout:    30,
		GrpcAddr:      "0.0.0.0:8082",
		GrpcIps:       KtStr.SplitByte("*", ',', true, 0, 0),
	}

	APro.SubCfgBind("proxy", Config)
	Config.AclTimeout = Config.AclTimeout * time.Second
}

func (that *config) AclCtx() context.Context {
	if that.AclTimeout <= 0 {
		return context.TODO()
	}

	ctx, _ := context.WithTimeout(context.TODO(), that.AclTimeout)
	return ctx
}
