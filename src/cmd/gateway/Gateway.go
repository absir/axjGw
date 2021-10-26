package main

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axjGW/pkg/gateway"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"runtime"
	"strings"
)

type Config struct {
	IdleTime   int64
	CheckTime  int64
	HttpAddr   string
	HttpWs     bool
	SocketAddr string
	SocketSize int
	SocketOut  bool
	ThriftAddr string
	ThriftIps  []string
}

var GCfg = Config{
	HttpAddr:   ":8082",
	HttpWs:     true,
	SocketAddr: ":8083",
	ThriftAddr: "127.0.0.1:8082",
	ThriftIps:  KtStr.SplitByte("*", ',', true, 0, 0),
}

var GwWorkHash int

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../resources")
	APro.Load(nil, "config.yaml")

	// 默认配置
	{
		KtCvt.BindInterface(GCfg, APro.Cfg)
		GwWorkHash = int(APro.WorkId())
	}

	// Gw服务初始化
	gateway.Server.Init(APro.WorkId())
	// Gw服务开启
	gateway.Server.StartGw()
	// thrift服务开启
	gateway.Server.StartThrift(GCfg.ThriftAddr, GCfg.ThriftIps)

	// socket连接
	if GCfg.SocketAddr != "" && !strings.HasPrefix(GCfg.SocketAddr, "!") {
		// socket服务
		serv, err := net.Listen("tcp", GCfg.SocketAddr)
		Kt.Panic(err)
		defer serv.Close()
		go func() {
			conn, err := serv.Accept()
			if err != nil {
				AZap.Logger.Warn("serv Accept err", zap.Error(err))
			}

			go gateway.Server.ConnLoop(ANet.NewConnSocket(conn.(*net.TCPConn)))
		}()
	}

	// websocket连接
	if GCfg.HttpAddr != "" && !strings.HasPrefix(GCfg.SocketAddr, "!") {
		// http服务
		if GCfg.HttpWs {
			// websocket连接
			http.Handle("gw", websocket.Handler(func(conn *websocket.Conn) {
				go gateway.Server.ConnLoop(ANet.NewConnWebsocket(conn))
			}))
		}

		err := http.ListenAndServe(GCfg.HttpAddr, nil)
		Kt.Panic(err)
	}

	// 等待关闭
	APro.Signal()
}
