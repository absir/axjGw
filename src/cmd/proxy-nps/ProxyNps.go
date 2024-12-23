package main

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/AZap/AZapIst"
	"axj/Thrd/Util"
	"axjGW/cmd/proxy-nps/nps"
	"axjGW/pkg/proxy"
	"go.uber.org/zap"
	"net"
	"runtime"
	"strings"
)

var WorkHash int

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../resources")
	APro.Load(nil, "proxy.yml")

	// 默认配置
	{
		WorkHash = int(APro.WorkId())
	}

	// 内存池
	Util.SetBufferPoolsS(APro.GetCfg("bPools", KtCvt.String, "256,512,1024,5120,10240,20480").(string))
	// 代理服务初始化
	proxy.PrxServMng.Init(APro.WorkId(), APro.Cfg, &nps.NpsAcl{})
	APro.SubCfgBind("nps", nps.NpsConfig)
	// 代理服务开启
	proxy.PrxServMng.Start()
	nps.LoadAll()
	if len(nps.NpsConfig.HttpAddr) > 1 {
		proxy.StartServ("http", nps.NpsConfig.HttpAddr, 0, proxy.FindProto("http", true), nil)
	}

	if len(nps.NpsConfig.RtspAddr) > 1 {
		proxy.StartServ("rtsp", nps.NpsConfig.RtspAddr, 0, proxy.FindProto("rtsp", true), nil)
	}

	// 开启面板
	nps.NpsApiInit()

	Config := proxy.Config
	// socket连接
	if Config.SocketAddr != "" && !strings.HasPrefix(Config.SocketAddr, "!") {
		// socket服务
		AZap.Logger.Info("StartProxy: " + Config.SocketAddr)
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

					AZap.Logger.Warn("Proxy Accept Err", zap.Error(err))
					continue
				}

				proxy.PrxServMng.Accept(conn.(*net.TCPConn))
			}
		}()
	}

	// 启动完成
	AZap.Logger.Info("ProxyNps all AXJ started")
	// 日志配置
	AZapIst.InitCfg(true)
	// 等待关闭
	APro.Signal()
}
