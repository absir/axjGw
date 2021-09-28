package gateway

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axjGW/pkg/gateway"
	"context"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	cacheHttp  string
	idleTime   int64
	checkTime  int64
	httpAddr   string
	httpWs     bool
	socketAddr string
	socketSize int
	socketOut  bool
	addrPub    string
}

var GCfg = Config{}
var GWorkHash int
var GContext context.Context

var GHandler *gateway.Handler
var GConnMng *ANet.ConnMng

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
	APro.Load(nil, "config.yaml")

	// 默认配置
	{
		GCfg.idleTime = 30000
		GCfg.checkTime = 3000
		GCfg.httpAddr = ":8082"
		GCfg.httpWs = true
		GCfg.socketAddr = ":8083"
		GCfg.addrPub = "127.0.0.1:8082"
		KtCvt.BindInterface(GCfg, APro.Cfg)
		GWorkHash = int(APro.WorkId())
		GContext = context.Background()
	}

	// ANet服务
	GHandler = new(gateway.Handler)
	GConnMng = ANet.NewConnMng(GHandler, APro.WorkId(), time.Duration(GCfg.idleTime)*time.Millisecond, time.Duration(GCfg.checkTime)*time.Millisecond)
	// 空闲检测
	go GConnMng.IdleLoop()

	if GCfg.httpAddr != "" && !strings.HasPrefix(GCfg.socketAddr, "!") {
		// http服务
		if GCfg.httpWs {
			// websocket连接
			http.Handle("ws", websocket.Handler(func(conn *websocket.Conn) {
				connect(ANet.NewClientWebsocket(conn))
			}))
		}

		err := http.ListenAndServe(GCfg.httpAddr, nil)
		Kt.Panic(err)
	}

	if GCfg.socketAddr != "" && !strings.HasPrefix(GCfg.socketAddr, "!") {
		// socket服务
		serv, err := net.Listen("tcp", GCfg.socketAddr)
		Kt.Panic(err)
		defer serv.Close()
		go func() {
			conn, err := serv.Accept()
			if err != nil {
				AZap.Logger.Warn("serv Accept err", zap.Error(err))
			}

			go connect(ANet.NewClientSocket(conn.(*net.TCPConn), GCfg.socketSize, GCfg.socketOut))
		}()
	}

	// 等待关闭
	APro.Signal()
}

func connect(client ANet.Client) {
	connM := GConnMng.OpenConnM(client)
	if !connectDo(connM) {
		return
	}

	go connM.ReqLoop()
}

func connectDo(connM *ANet.ConnM) bool {
	err, _, uri, uriI, data := connM.Req()
	if err != nil || data == nil {
		AZap.Logger.Warn("serv acl Req err", zap.Error(err))
		return false
	}

	login, err := gateway.GetProds(gateway.Config.AclProd).GetProdHash(GWorkHash).GetAclClient().Login(GContext, connM.Id(), data)
	if err != nil || login == nil {
		AZap.Logger.Warn("serv acl Login err", zap.Error(err))
		return false
	}

	// 路由hash校验
	if uri != "" && uriI > 0 {
		if uri != gateway.UriDict.UriMapHash {
			// 路由缓存
			connM.Get().Rep(ANet.REQ_ROUTE, gateway.UriDict.UriMapHash, 0, gateway.UriDict.UriMapJsonData, false, false, nil)
		}
	}

	// 用户注册

	// 用户状态设置
	connH := GHandler.ConnH(connM)
	connH.SetId(login.UID, login.Sid)
	// 路由规则
	connH.PutId("", login.Aid)
	// 连接注册
	GConnMng.RegConn(connH, int(login.PoolG))
	// 开启服务
	connM.Get().Rep(ANet.REQ_READY, "", 0, nil, false, false, nil)
	return true
}
