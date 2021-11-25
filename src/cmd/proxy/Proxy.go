package main

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/AZap/AZapIst"
	"axjGW/pkg/gateway"
	"axjGW/pkg/proxy"
	"go.uber.org/zap"
	"net"
	"runtime"
	"strings"
)

type Config struct {
	SocketAddr  string // socket服务地址
	CompressMin int    // 最短压缩
	DataMax     int32  // 最大数据(请求)
	Encrypt     bool   // 通讯加密
	CheckDrt    int64  // 客户端检查间隔
	IdleDrt     int64  // 空闲检测间隔
}

var PCfg = Config{
	SocketAddr:  ":8683",
	CompressMin: 256,
	DataMax:     256 << 10,
	Encrypt:     true,
	CheckDrt:    3000,
	IdleDrt:     30000,
}

var PWorkHash int

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../resources")
	APro.Load(nil, "config.yml")

	// 默认配置
	{
		KtCvt.BindInterface(&PCfg, APro.Cfg)
		PWorkHash = int(APro.WorkId())
	}

	// 代理服务初始化

	proxy.PrxServ.Init(APro.WorkId(), APro.Cfg, new(proxy.PrxAcl))
	// 代理服务开启
	proxy.PrxServ.Start()

	// socket连接
	if PCfg.SocketAddr != "" && !strings.HasPrefix(PCfg.SocketAddr, "!") {
		// socket服务
		AZap.Logger.Info("StartSocket: " + PCfg.SocketAddr)
		serv, err := net.Listen("tcp", PCfg.SocketAddr)
		Kt.Panic(err)
		defer serv.Close()
		go func() {
			for !APro.Stopped {
				conn, err := serv.Accept()
				if err != nil {
					if APro.Stopped {
						return
					}

					AZap.Logger.Warn("Serv Accept Err", zap.Error(err))
					continue
				}

				go gateway.Server.ConnLoop(ANet.NewConnSocket(conn.(*net.TCPConn), false))
			}
		}()
	}

	// 启动完成
	AZap.Logger.Info("Proxy all AXJ started")
	// 日志配置
	AZapIst.InitCfg(true)
	// 等待关闭
	APro.Signal()
}
