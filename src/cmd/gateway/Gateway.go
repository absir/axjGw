package main

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axjGW/pkg/gateway"
	"go.uber.org/zap"
	"golang.org/x/net/websocket"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"
)

type Config struct {
	pool       bool
	idleTime   int64
	checkTime  int64
	httpAddr   string
	httpWs     bool
	socketAddr string
	socketSize int
	socketOut  bool
	addrPub    string
}

var GCfg = Config{
	pool:       true,
	idleTime:   30000,
	checkTime:  3000,
	httpAddr:   ":8082",
	httpWs:     true,
	socketAddr: ":8083",
	addrPub:    "127.0.0.1:8082",
}
var GWorkHash int

var GHandler *gateway.HandlerG
var GConnMng *ANet.ConnMng

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
	APro.Load(nil, "config.yaml")

	// 默认配置
	{
		KtCvt.BindInterface(GCfg, APro.Cfg)
		GWorkHash = int(APro.WorkId())
	}

	// ANet服务
	GHandler = new(gateway.HandlerG)
	GConnMng = ANet.NewConnMng(GHandler, APro.WorkId(), time.Duration(GCfg.idleTime)*time.Millisecond, time.Duration(GCfg.checkTime)*time.Millisecond, GCfg.pool)
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
		if GCfg.pool {
			// Socket客户端对象池开启
			ANet.SetClientSocketPool(true)
		}

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
	conn := GConnMng.OpenConn(client)
	if !connectDo(conn) {
		return
	}

	go conn.Get().ReqLoop()
}

func connectDo(conn ANet.Conn) bool {
	connM := GConnMng.ConnM(conn)
	// 交换密钥
	encrypt := GConnMng.Processor().Encrypt
	if encrypt != nil {
		sKey, cKey := encrypt.NewKeys()
		if sKey != nil && cKey != nil {
			connM.SetEncryKey(sKey)
			// 路由缓存
			connM.Get().Rep(ANet.REQ_KEY, "", 0, cKey, false, false, nil)
		}
	}

	// 服务准备
	connM.Get().Rep(ANet.REQ_ACL, "", 0, nil, false, false, nil)
	// Arl请求
	err, _, uri, uriI, data := connM.Req()
	if err != nil || data == nil {
		AZap.Logger.Warn("serv acl Req err", zap.Error(err))
		return false
	}

	login, err := gateway.GetProds(gateway.Config.AclProd).GetProdHash(GWorkHash).GetAclClient().Login(gateway.MsgMng.Context, connM.Id(), data)
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
	connG := GHandler.ConnG(conn)
	connG.SetId(login.UID, login.Sid)
	// 路由服务规则
	connG.PutRId("", login.Rid)
	connG.PutRIds(login.Rids)
	// 连接注册
	GConnMng.RegConn(connG, int(login.PoolG))
	// 开启服务
	connM.Get().Rep(ANet.REQ_LOOP, "", 0, nil, false, false, nil)
	return true
}
