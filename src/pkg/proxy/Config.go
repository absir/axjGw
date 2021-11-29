package proxy

import (
	"axj/Kt/KtBytes"
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
	AdapCheckDrt  time.Duration // 客户端检查间隔
	AdapCheckBuff int           // 客户端Buff
	AdapTimeout   int64         // 适配超时
	AdapMaxId     int32         // 适配最大编号
}

var Config *config

func initConfig() {
	Config = &config{
		SocketAddr:    ":8683",
		CompressMin:   256,
		DataMax:       256 << 10,
		Encrypt:       true,
		CheckDrt:      3000,
		IdleDrt:       30000,
		KickDrt:       6000,
		AdapCheckDrt:  3000,
		AdapCheckBuff: 128,
		AdapTimeout:   30000,
		AdapMaxId:     KtBytes.VINT_2_MAX,
	}

	Config.CheckDrt *= time.Millisecond
	Config.AdapTimeout *= int64(time.Millisecond)
}
