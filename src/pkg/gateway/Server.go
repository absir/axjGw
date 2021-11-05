package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axjGW/gen/gw"
	"axjGW/pkg/ext"
	"context"
	"github.com/apache/thrift/lib/go/thrift"
	"go.uber.org/zap"
	"io"
	"sync"
	"time"
)

type server struct {
	Manager  *ANet.Manager
	Context  context.Context
	prodsMap *sync.Map
	prodsGw  *Prods
	gatewayI gw.GatewayI
}

var Server = new(server)

func (that *server) Init(workId int32, cfg map[interface{}]interface{}, gatewayI gw.GatewayI) {
	// 配置初始化
	initConfig(workId)
	// 初始化服务
	that.initProds(cfg)
	// Handler初始化
	initHandler()
	// 路由字典初始化
	initUriDict()
	// 消息管理初始化
	initMsgMng()
	// 群组管理器
	initTeamMng()
	// 聊天管理
	initChatMng()
	that.Manager = ANet.NewManager(Handler, workId, time.Duration(Config.IdleDrt)*time.Millisecond, time.Duration(Config.CheckDrt)*time.Millisecond)
	that.Context = context.Background()
	that.gatewayI = gatewayI
}

func (that *server) initProds(cfg map[interface{}]interface{}) {
	that.prodsMap = new(sync.Map)
	prods := cfg["prods"]
	if prods != nil {
		if mp, _ := prods.(*Kt.LinkedMap); mp != nil {
			for el := mp.Front(); el != nil; el = el.Next() {
				if key, ok := el.Value.(string); ok {
					if pMp, _ := mp.Get(key).(*Kt.LinkedMap); pMp != nil {
						pProds := new(Prods)
						for pEl := pMp.Front(); pEl != nil; pEl = pEl.Next() {
							pProds.Add(KtCvt.ToType(pEl.Value, KtCvt.Int32).(int32), KtCvt.ToType(pMp.Get(pEl.Value), KtCvt.String).(string))
						}

						that.prodsMap.Store(key, pProds)
					}
				}
			}
		}
	}
}

func (that *server) StartGw() {
	go that.Manager.CheckLoop()
	go MsgMng.CheckLoop()
	go ChatMng.CheckLoop()
}

func (that *server) ConnLoop(conn ANet.Conn) {
	client := that.connOpen(conn)
	if client != nil {
		client.Get().ReqLoop()

	} else {
		// 连接失败关闭
		conn.Close()
	}
}

func (that *server) connOpen(conn ANet.Conn) ANet.Client {
	manager := that.Manager
	// 交换密钥
	var encryptKey []byte = nil
	processor := manager.Processor()
	_, _, _, flag, _ := processor.Req(conn, encryptKey)
	compress := (flag & ANet.FLG_COMPRESS) != 0
	if (flag&ANet.FLG_ENCRYPT) != 0 && processor.Encrypt != nil {
		sKey, cKey := processor.Encrypt.NewKeys()
		if sKey != nil && cKey != nil {
			encryptKey = sKey
			// 连接秘钥
			err := processor.Rep(nil, true, conn, nil, compress, ANet.REQ_KEY, "", 0, encryptKey, false, 0)
			if err != nil {
				return nil
			}
		}
	}

	// Acl准备
	err := processor.Rep(nil, true, conn, nil, compress, ANet.REQ_ACL, "", 0, encryptKey, false, 0)
	if err != nil {
		return nil
	}

	// 登录请求
	err, _, uriHash, uriRoute, loginData := processor.Req(conn, encryptKey)
	if err != nil {
		if err == io.EOF {
			AZap.Logger.Debug("serv acl Req EOF")

		} else {
			AZap.Logger.Warn("serv acl Req ERR", zap.Error(err))
		}

		return nil
	}

	// 登录Acl处理
	id := manager.IdWorker().Generate()
	aclClient := that.GetProds(Config.AclProd).GetProdHash(Config.WorkHash).GetAclClient()
	login, err := aclClient.Login(that.Context, id, loginData, conn.RemoteAddr())
	if err != nil || login == nil {
		if err == io.EOF {
			AZap.Logger.Debug("serv acl Login EOF")

		} else {
			if err == nil {
				AZap.Logger.Debug("serv acl Login Fail nil")

			} else {
				AZap.Logger.Debug("serv acl Login Fail " + err.Error())
			}
		}

		return nil
	}

	// 客户端注册
	client := manager.Open(conn, encryptKey, compress, id)
	clientG := Handler.ClientG(client)
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
			processor.Rep(nil, true, conn, nil, compress, ANet.REQ_ROUTE, UriDict.UriMapHash, 0, UriDict.UriMapJsonData, false, 0)
		}
	}

	// 消息处理
	if clientG.Gid() != "" {
		// 清理消息队列配置
		result, err := Server.GetProdClient(clientG).GetGWIClient().GQueue(Server.Context, clientG.gid, clientG.Id(), clientG.Unique(), login.Clear)
		if result != gw.Result__Succ {
			clientG.Close(err, nil)
			return nil
		}
	}

	// 注册成功回调
	if login.Back {
		aclClient.LoginBack(that.Context, clientG.Id(), clientG.uid, clientG.sid)
	}

	// 登录成功
	client.Get().Rep(true, ANet.REQ_LOOP, "", 0, login.Data, false, false, 0)
	if !client.Get().IsClosed() {
		return nil
	}

	// 并发限制
	if login.Limit > 0 {
		clientG.SetLimiter(int(login.Limit))
	}

	return client
}

func (that *server) StartThrift(addr string, ips []string, gateway gw.Gateway) {
	AZap.Logger.Info("StartThrift: " + addr)
	socket, err := thrift.NewTServerSocket(addr)
	Kt.Panic(err)
	processor := thrift.NewTMultiplexedProcessor()
	processor.RegisterProcessor("i", gw.NewGatewayIProcessor(that.gatewayI))
	processor.RegisterProcessor("", gw.NewGatewayProcessor(gateway))
	matchers := KtStr.ForMatchers(ips, false, true)
	go thrift.NewTSimpleServer4(processor, ext.NewTServerSocketIps(socket, func(ip string) bool {
		return KtStr.Matchers(matchers, ip, true)
	}), thrift.NewTTransportFactory(), thrift.NewTCompactProtocolFactoryConf(Config.TConfig)).Serve()
}

func (that *server) IsProdCid(cid int64) bool {
	workId := that.Manager.IdWorker().GetWorkerId(cid)
	if workId == Config.WorkId {
		return true
	}

	return false
}

func (that *server) IsProdHash(hash int) bool {
	prod := that.GetProds(Config.GwProd).GetProdHash(hash)
	return prod != nil && prod.id == Config.WorkId
}

func (that *server) IsProdHashS(hash string) bool {
	prod := that.GetProds(Config.GwProd).GetProdHashS(hash)
	return prod != nil && prod.id == Config.WorkId
}

func (that *server) GetProds(name string) *Prods {
	val, _ := that.prodsMap.Load(name)
	if val == nil {
		if name == Config.GwProd {
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

func (that *server) PutProds(name string, prods *Prods) {
	if prods == nil {
		that.prodsMap.Delete(name)

	} else {
		that.prodsMap.Store(name, prods)
	}
}

func (that *server) GetProdCid(cid int64) *Prod {
	return that.GetProds(Config.GwProd).GetProd(that.Manager.IdWorker().GetWorkerId(cid))
}

func (that *server) GetProdGid(gid string) *Prod {
	return that.GetProds(Config.GwProd).GetProdHashS(gid)
}

func (that *server) GetProdClient(clientG *ClientG) *Prod {
	return that.GetProds(Config.GwProd).GetProdHash(clientG.Hash())
}
