package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtStr"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axjGW/pkg/ext"
	"axjGW/pkg/gws"
	"context"
	"github.com/apache/thrift/lib/go/thrift"
	"go.uber.org/zap"
	"gw"
	"sync"
	"time"
)

type server struct {
	Handler  *handler
	Manager  *ANet.Manager
	Context  context.Context
	prodsMap sync.Map
	prodsGw  *Prods
	gatewayI gw.GatewayI
}

var Server = new(server)

func (that server) Init(workId int32) {
	// 配置初始化
	initConfig(workId)
	// 路由字典初始化
	initUriDict()
	// 消息管理初始化
	initMsgMng()
	that.Handler = new(handler)
	that.Manager = ANet.NewManager(that.Handler, workId, time.Duration(Config.IdleDrt)*time.Millisecond, time.Duration(Config.CheckDrt)*time.Millisecond)
	that.Context = context.Background()
	that.gatewayI = new(gws.GatewayIs)
}

func (that server) StartGw() {
	go that.Manager.CheckLoop()
}

func (that server) ConnLoop(conn ANet.Conn) {
	client := that.connOpen(conn)
	if client != nil {
		client.Get().ReqLoop()
	}
}

func (that server) connOpen(conn ANet.Conn) ANet.Client {
	manager := that.Manager
	// 交换密钥
	var encryptKey []byte = nil
	processor := manager.Processor()
	if processor.Encrypt != nil {
		sKey, cKey := processor.Encrypt.NewKeys()
		if sKey != nil && cKey != nil {
			encryptKey = sKey
			// 连接秘钥
			err := processor.Rep(nil, true, conn, nil, ANet.REQ_KEY, "", 0, encryptKey, false)
			if err != nil {
				return nil
			}
		}
	}

	// Acl准备
	err := processor.Rep(nil, true, conn, nil, ANet.REQ_ACL, "", 0, encryptKey, false)
	if err != nil {
		return nil
	}

	// 登录请求
	err, _, uriHash, uriRoute, loginData := processor.Req(conn, encryptKey)
	if err != nil || loginData == nil {
		AZap.Logger.Warn("serv acl Req err", zap.Error(err))
		return nil
	}

	// 登录Acl处理
	id := manager.IdWorker().Generate()
	aclClient := that.GetProds(Config.AclProd).GetProdHash(Config.WorkHash).GetAclClient()
	login, err := aclClient.Login(that.Context, id, loginData)
	if err != nil || login == nil {
		AZap.Logger.Warn("serv acl Login err", zap.Error(err))
		return nil
	}

	// 客户端注册
	client := manager.Open(conn, encryptKey, id)
	clientG := that.Handler.ClientG(client)
	// 用户状态设置
	clientG.SetId(login.UID, login.Sid, login.Unique)
	if clientG.Gid() != "" {
		// 用户连接保持
		clientG.ConnKeep()
		clientG.ConnCheck()
		if clientG.IsClosed() {
			return nil
		}
	}

	// 路由服务规则
	clientG.PutRId("", login.Rid)
	clientG.PutRIds(login.Rids)
	// 路由hash校验
	if uriRoute > 0 {
		if uriHash != UriDict.UriMapHash {
			// 路由缓存
			processor.Rep(nil, true, conn, nil, ANet.REQ_ROUTE, UriDict.UriMapHash, 0, UriDict.UriMapJsonData, false)
		}
	}

	// 注册成功回调
	if login.Back {
		aclClient.LoginBack(that.Context, clientG.Id(), clientG.uid, clientG.sid)
	}

	// 登录成功
	client.Get().Rep(true, ANet.REQ_LOOP, "", 0, login.Data, false, false)
	if !client.Get().IsClosed() {
		return nil
	}

	// 并发限制
	if login.Limit > 0 {
		clientG.SetLimiter(int(login.Limit))
	}

	return client
}

func (that server) StartThrift(addr string, ips []string) {
	socket, err := thrift.NewTServerSocket(addr)
	Kt.Panic(err)
	processor := gw.NewGatewayIProcessor(that.gatewayI)
	matchers := KtStr.ForMatchers(ips, false, true)
	go thrift.NewTSimpleServer4(processor, ext.NewTServerSocketIps(socket, func(ip string) bool {
		return KtStr.Matchers(matchers, ip, true)
	}), thrift.NewTTransportFactory(), thrift.NewTCompactProtocolFactoryConf(Config.TConfig)).Serve()
}

func (that server) IsProdCid(cid int64) bool {
	workId := that.Manager.IdWorker().GetWorkerId(cid)
	if workId == Config.WorkId {
		return true
	}

	return false
}

func (that server) IsProdHash(hash int) bool {
	prod := that.GetProds(Config.GwProd).GetProdHash(hash)
	return prod != nil && prod.id == Config.WorkId
}

func (that server) GetProds(name string) *Prods {
	val, _ := that.prodsMap.Load(name)
	if val == nil {
		if val == Config.GwProd {
			if that.prodsGw == nil {
				prods := new(Prods)
				prods.Add(Config.WorkId, "")
				that.prodsGw = prods
			}

			return that.prodsGw
		}

		return nil
	}

	return val.(*Prods)
}

func (that server) PutProds(name string, prods *Prods) {
	if prods == nil {
		that.prodsMap.Delete(name)

	} else {
		that.prodsMap.Store(name, prods)
	}
}

func (that server) GetProdCid(cid int64) *Prod {
	return that.GetProds(Config.GwProd).GetProd(that.Manager.IdWorker().GetWorkerId(cid))
}

func (that server) GetProdMid(mid string) *Prod {
	return that.GetProds(Config.GwProd).GetProdHash(Kt.HashCode(KtUnsafe.StringToBytes(mid)))
}

func (that server) GetProdClient(clientG *ClientG) *Prod {
	return that.GetProds(Config.GwProd).GetProdHash(clientG.Hash())
}
