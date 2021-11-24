package main

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axj/Thrd/AZap/AZapIst"
	"axjGW/pkg/gateway"
	"axjGW/pkg/gws"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"runtime"
	"strings"
)

type config struct {
	HttpAddr   string   // http服务地址
	HttpWs     bool     // 启用ws网关
	HttpWsPath string   // ws连接地址
	SocketAddr string   // socket服务地址
	GrpcAddr   string   // grpc服务地址
	GrpcIps    []string // grpc调用Ip白名单，支持*通配
	LastUrl    string   // 消息持久化，数据库连接
}

var Config = &config{
	HttpAddr:   ":8682",
	HttpWs:     true,
	HttpWsPath: "/gw",
	SocketAddr: ":8683",
	GrpcAddr:   "127.0.0.1:8082",
	GrpcIps:    KtStr.SplitByte("*", ',', true, 0, 0),
}

var WorkHash int

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../resources")
	APro.Load(nil, "config.yml")

	// 默认配置
	{
		KtCvt.BindInterface(Config, APro.Cfg)
		WorkHash = int(APro.WorkId())
	}

	// Gw服务初始化
	gateway.Server.Init(APro.WorkId(), APro.Cfg, new(gws.GatewayIs))
	if APro.Cfg.Get("msg") != nil {
		// 初始化MsgMng
		gateway.MsgMng()
	}
	// Gw服务开启
	gateway.Server.StartGw()
	// Grpc服务开启
	gateway.Server.StartGrpc(Config.GrpcAddr, Config.GrpcIps, new(gws.GatewayS))

	// socket连接
	if Config.SocketAddr != "" && !strings.HasPrefix(Config.SocketAddr, "!") {
		// socket服务
		AZap.Logger.Info("StartSocket: " + Config.SocketAddr)
		serv, err := net.Listen("tcp", Config.SocketAddr)
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

				go gateway.Server.ConnLoop(ANet.NewConnSocket(conn.(*net.TCPConn)))
			}
		}()
	}

	// websocket连接
	if Config.HttpAddr != "" && !strings.HasPrefix(Config.SocketAddr, "!") {
		// http服务
		AZap.Logger.Info("StartHttp: " + Config.HttpAddr)
		if Config.HttpWs {
			AZap.Logger.Info("StartHttpWs: " + Config.HttpWsPath)
			// websocket连接
			http.Handle(Config.HttpWsPath, websocket.Handler(func(conn *websocket.Conn) {
				gateway.Server.ConnLoop(ANet.NewConnWebsocket(conn))
			}))
		}

		go func() {
			err := http.ListenAndServe(Config.HttpAddr, nil)
			Kt.Panic(err)
		}()
	}

	// 启动完成
	AZap.Logger.Info("Gateway all AXJ started")
	// 日志配置
	AZapIst.InitCfg(true)
	// 等待关闭
	APro.Signal()
}
