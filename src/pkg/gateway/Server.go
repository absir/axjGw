package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axjGW/gen/gw"
	"context"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type server struct {
	Locker     sync.Locker
	Manager    *ANet.Manager
	Context    context.Context
	prodsMap   *sync.Map
	prodsGw    *Prods
	gatewayISC *gatewayISC
	cron       *cron.Cron
	started    bool
}

var Server = new(server)

func (that *server) Id32(rep *gw.Id32Rep) int32 {
	if rep == nil {
		return 0
	}

	return rep.Id
}

func (that *server) Id64(rep *gw.Id64Rep) int64 {
	if rep == nil {
		return 0
	}

	return rep.Id
}

func (that *server) CidGen(compress bool) int64 {
	flg := 0
	if compress {
		flg = 1
	}

	return that.Manager.IdWorker().GenerateM(2, flg)
}

func (that *server) CidCompress(cid int64) bool {
	return (cid & 0X01) == 1
}

func (that *server) Cron() *cron.Cron {
	if that.cron == nil {
		that.Locker.Lock()
		defer that.Locker.Unlock()
		if that.cron == nil {
			that.cron = cron.New(cron.WithParser(cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)))
			if that.started {
				that.cron.Start()
			}
		}
	}

	return that.cron
}

func (that *server) Init(workId int32, cfg map[interface{}]interface{}, gatewayI gw.GatewayIServer) {
	// 全局锁
	that.Locker = new(sync.Mutex)
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
	that.gatewayISC = &gatewayISC{Server: gatewayI}
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
	that.Locker.Lock()
	defer that.Locker.Unlock()
	that.started = true
	if that.cron != nil {
		that.cron.Start()
	}
}

func (that *server) ConnLoop(conn ANet.Conn) {
	client := that.connOpen(&conn)
	if client != nil {
		client.Get().ReqLoop()

	} else if conn != nil {
		// 连接失败关闭
		conn.Close()
	}
}

func (that *server) connOpen(pConn *ANet.Conn) ANet.Client {
	conn := *pConn
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
	cid := that.CidGen(compress)
	aclClient := that.GetProds(Config.AclProd).GetProdHash(Config.WorkHash).GetAclClient()
	login, err := aclClient.Login(that.Context, &gw.LoginReq{
		Cid:  cid,
		Data: loginData,
		Addr: conn.RemoteAddr(),
	})

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

	if login.KickData != nil {
		// 登录失败被踢
		processor.Rep(nil, true, conn, nil, compress, ANet.REQ_KICK, "", 0, login.KickData, false, 0)
		*pConn = nil
		go ANet.CloseDelay(conn, Config.KickDrt)
		return nil
	}

	// 客户端注册
	client := manager.Open(conn, encryptKey, compress, cid)
	clientG := Handler.ClientG(client)
	// clientG.Kick()
	// 用户状态设置
	clientG.SetId(login.Uid, login.Sid, login.Unique, login.DiscBack)
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

	// 消息处理 单用户登录
	if clientG.Gid() != "" && clientG.unique == "" {
		// 消息队列清理开启
		rep, err := Server.GetProdClient(clientG).GetGWIClient().GQueue(Server.Context, &gw.IGQueueReq{
			Gid:    clientG.gid,
			Cid:    clientG.Id(),
			Unique: clientG.Unique(),
			Clear:  login.Clear,
		})

		if Server.Id32(rep) < R_SUCC_MIN {
			clientG.Close(err, nil)
			return nil
		}
	}

	// 注册成功回调
	if login.Back {
		rep, err := aclClient.LoginBack(that.Context, &gw.LoginBack{
			Cid:    clientG.Id(),
			Unique: clientG.unique,
			Uid:    clientG.uid,
			Sid:    clientG.sid,
		})

		if Server.Id32(rep) < R_SUCC_MIN {
			clientG.Close(err, nil)
			return nil
		}
	}

	// 登录成功
	client.Get().Rep(true, ANet.REQ_LOOP, strconv.FormatInt(cid, 10)+"/"+clientG.unique+"/"+clientG.gid, 0, login.Data, false, false, 0)
	if client.Get().IsClosed() {
		return nil
	}

	// 并发限制
	if login.Limit > 0 {
		clientG.SetLimiter(int(login.Limit))
	}

	return client
}

func (that *server) StartGrpc(addr string, ips []string, gateway gw.GatewayServer) {
	AZap.Logger.Info("StartGrpc: " + addr)
	lis, err := net.Listen("tcp", addr)
	Kt.Panic(err)
	recoverFun := func() {
		if err := recover(); err != nil {
			AZap.Logger.Warn("grpc stream err", zap.Reflect("err", err))
		}
	}
	serv := grpc.NewServer(
		grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
			defer recoverFun()
			return handler(ctx, req)
		}),
	)
	gw.RegisterGatewayIServer(serv, that.gatewayISC.Server)
	gw.RegisterGatewayServer(serv, gateway)
	matchers := KtStr.ForMatchers(ips, false, true)
	lisIps := ANet.NewListenerIps(lis, func(ip string) bool {
		return KtStr.Matchers(matchers, ip, true)
	})
	go func() {
		if err := serv.Serve(lisIps); err != nil {
			AZap.Logger.Error("grpc server err "+addr, zap.Error(err))
		}
	}()
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

	prods, _ := val.(*Prods)
	return prods
}

func (that *server) PutProds(name string, prods *Prods) {
	if prods == nil {
		that.prodsMap.Delete(name)

	} else {
		that.prodsMap.Store(name, prods)
	}
}

func (that *server) GetProdId(id int32) *Prod {
	return that.GetProds(Config.GwProd).GetProd(id)
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
