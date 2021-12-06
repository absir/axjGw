package proxy

import (
	"axj/Kt/KtBytes"
	"context"
	"time"
)

type config struct {
	SocketAddr    string // socket服务地址
	SocketOut     bool   // socket流写入
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
	CloseDelay    time.Duration     // 关闭延迟
	Servs         map[string]*Serv  // 协议服务
	ClientKeys    map[string]string // 客户端授权码
	Acl           string            // Acl服务地址
	AclTry        time.Duration     // Acl服务地址
	AclTimeout    time.Duration     // Acl调用超时
}

type Serv struct {
	Addr  string
	Proto string
	Cfg   map[string]string
}

var Config *config

func initConfig() {
	Config = &config{
		SocketAddr:    ":8783",
		CompressMin:   256,
		DataMax:       256 << 10,
		Encrypt:       true,
		CheckDrt:      3000,
		IdleDrt:       30000,
		KickDrt:       6000,
		AdapCheckDrt:  3000,
		AdapCheckBuff: 128,
		AdapTimeout:   60000,
		AdapMaxId:     KtBytes.VINT_3_MAX,
		CloseDelay:    30,
		Servs:         map[string]*Serv{},
		ClientKeys:    map[string]string{},
		AclTry:        3000,
		AclTimeout:    30000,
	}

	Config.CheckDrt *= time.Millisecond
	Config.AdapTimeout *= int64(time.Millisecond)
	Config.AclTimeout *= time.Millisecond
	Config.CloseDelay *= time.Second
}

func (that *config) AclCtx() context.Context {
	if that.AclTimeout <= 0 {
		return context.TODO()
	}

	ctx, _ := context.WithTimeout(context.TODO(), that.AclTimeout)
	return ctx
}
