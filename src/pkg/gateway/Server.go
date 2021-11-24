package gateway

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axj/Thrd/Disc"
	"axj/Thrd/Util"
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
	Locker       sync.Locker
	Manager      *ANet.Manager
	Context      context.Context
	prodsMap     map[string]*Prods
	prodsDiscMng *Disc.DiscoveryMng
	prodsGw      *Prods
	gatewayISC   *gatewayISC
	cron         *cron.Cron
	started      bool
	connLimiter  Util.Limiter
	liveLimiter  Util.Limiter
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

func (that *server) getLimiter(pLimiter *Util.Limiter, aclSize int) Util.Limiter {
	if aclSize < 0 {
		return nil
	}

	if aclSize == 0 {
		aclSize = that.GetProds(Config.AclProd).ids.Size()
	}

	if aclSize <= 1 {
		// 单节点，单协程
		return nil
	}

	limiter := *pLimiter
	if limiter == nil || limiter.Limit() != aclSize {
		that.Locker.Lock()
		limiter = *pLimiter
		if limiter == nil || limiter.Limit() != aclSize {
			limiter = Util.NewLimiterLocker(aclSize, nil)
			*pLimiter = limiter
		}

		that.Locker.Unlock()
	}

	return limiter
}

func (that *server) getConnLimiter() Util.Limiter {
	return that.getLimiter(&that.connLimiter, Config.ConnLimit)
}

func (that *server) getLiveLimiter() Util.Limiter {
	return that.getLimiter(&that.liveLimiter, Config.LiveLimit)
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
	that.Manager = ANet.NewManager(Handler, workId, time.Duration(Config.IdleDrt)*time.Millisecond, time.Duration(Config.CheckDrt)*time.Millisecond)
	that.Context = context.Background()
	that.gatewayISC = &gatewayISC{Server: gatewayI}
}

func (that *server) initProds(cfg map[interface{}]interface{}) {
	that.prodsMap = map[string]*Prods{}
	if mp, _ := cfg["prods"].(*Kt.LinkedMap); mp != nil {
		mp.Range(func(key interface{}, val interface{}) bool {
			name, _ := key.(string)
			if name == "" {
				return true
			}

			if pMp, _ := val.(*Kt.LinkedMap); pMp != nil {
				prods := new(Prods)
				KtCvt.BindInterface(prods, pMp)
				pMp.Range(func(key interface{}, val interface{}) bool {
					id := KtCvt.ToString(key)
					if id == "" {
						return true
					}

					if id[0] >= '0' && id[0] <= '9' {
						prods.Add(KtCvt.ToInt32(id), KtCvt.ToString(val))
					}

					return true
				})

				if prods != nil {
					that.initProdsReg(name, prods)
				}
			}

			return true
		})

		// 服务发现启动发现线程
		if that.prodsDiscMng != nil && !that.prodsDiscMng.CheckEmpty() {
			go that.prodsDiscMng.CheckLoop(Config.ProdCheckDrt)
		}
	}

	// 无服务配置
	if len(that.prodsMap) <= 0 {
		Config.zDevAcl = true
		prods := new(Prods)
		prods.Add(0, "")
		that.initProdsReg(Config.AclProd, prods)
	}

	// Gw服务默认配置
	if that.prodsMap[Config.GwProd] == nil {
		prods := new(Prods)
		prods.Add(APro.WorkId(), "")
		that.initProdsReg(Config.GwProd, prods)
	}
}

func (that *server) initProdsDiscMng() {
	that.prodsDiscMng = new(Disc.DiscoveryMng).Init()
	that.prodsDiscMng.RegDefs()
}

func (that *server) initProdsReg(name string, prods *Prods) {
	if prods == nil {
		delete(that.prodsMap, name)

	} else {
		if prods.Timeout > 0 {
			// 超时时间单位为秒
			prods.Timeout *= time.Second

		} else {
			prods.Timeout = Config.ProdTimeout
		}

		if prods.Disc != "" {
			if that.prodsDiscMng == nil {
				that.initProdsDiscMng()
			}

			// 服务发现
			prods.discS = that.prodsDiscMng.SetDiscoveryS(prods.Disc, name, prods.Set, prods.DiscIdle, true) != nil
		}

		that.prodsMap[name] = prods
	}
}

func (that *server) StartGw() {
	go that.Manager.CheckLoop()
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
			AZap.Logger.Debug("Serv Login Req EOF")

		} else {
			AZap.Logger.Warn("Serv Login Req ERR", zap.Error(err))
		}

		return nil
	}

	// 登录Acl处理
	cid := that.CidGen(compress)
	aclProds := that.GetProds(Config.AclProd)
	aclClient := aclProds.GetProdHash(Config.WorkHash).GetAclClient()
	login, err := aclClient.Login(aclProds.TimeoutCtx(), &gw.LoginReq{
		Cid:  cid,
		Data: loginData,
		Addr: conn.RemoteAddr(),
	})

	if err != nil || login == nil {
		if err == io.EOF {
			AZap.Logger.Debug("Serv Login Acl EOF")

		} else {
			if err == nil {
				AZap.Logger.Debug("Serv Login Acl Fail nil")

			} else {
				AZap.Debug("Serv Login Acl Fail %s", err.Error())
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
	client := manager.Open(conn, encryptKey, compress, (flag&ANet.FLG_OUT) != 0, cid)
	clientG := Handler.ClientG(client)
	// clientG.Kick()
	// 用户状态设置
	clientG.SetId(login.Uid, login.Sid, login.Unique, login.DiscBack)
	if clientG.Gid() != "" {
		// 用户连接保持
		clientG.connKeep()
		clientG.connCheck(nil)
		if clientG.IsClosed() {
			return nil
		}
	}

	// 请求并发限制
	if login.Limit > 0 {
		clientG.SetLimiter(int(login.Limit))
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
		rep, err := aclClient.LoginBack(aclProds.TimeoutCtx(), &gw.LoginBack{
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
			AZap.LoggerS.Warn("Grpc Recover Err", zap.Reflect("err", err))
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
			AZap.Logger.Error("Grpc Serve Err "+addr, zap.Error(err))
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
	return that.prodsMap[name]
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
