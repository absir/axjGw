package gateway

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axj/Thrd/Dscv"
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
	Locker      sync.Locker
	Manager     *ANet.Manager
	Context     context.Context
	prodsMap    map[string]*Prods
	prodsGw     *Prods
	gatewayISC  *gatewayISC
	gateway     gw.GatewayServer
	cron        *cron.Cron
	started     bool
	connLimiter Util.Limiter
	liveLimiter Util.Limiter
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

func (that *server) Cron(locker bool) *cron.Cron {
	if that.cron == nil {
		if locker {
			that.Locker.Lock()
			defer that.Locker.Unlock()
		}

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
	that.Manager = ANet.NewManager(Handler, workId, Config.IdleDrt*int64(time.Millisecond), Config.CheckDrt*time.Millisecond)
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
					that.initProdsDscv(name, prods)
				}
			}

			return true
		})
	}

	// 无服务开发配置
	prodsLen := len(that.prodsMap)
	if prodsLen <= 0 || (prodsLen == 1 && that.prodsMap[Config.GwProd] != nil) {
		Config.zDevAcl = true
		prods := new(Prods)
		prods.Add(0, "")
		that.initProdsDscv(Config.AclProd, prods)
	}

	// Gw服务默认配置
	if that.prodsMap[Config.GwProd] == nil {
		prods := new(Prods)
		prods.Add(APro.WorkId(), "")
		that.initProdsDscv(Config.GwProd, prods)
	}

	// 服务发现启动发现线程
	Dscv.InstMngStart(false)
}

func (that *server) initProdsDscv(name string, prods *Prods) {
	if prods == nil {
		delete(that.prodsMap, name)

	} else {
		if prods.Timeout > 0 {
			// 超时时间单位为秒
			prods.Timeout *= time.Second

		} else {
			prods.Timeout = Config.ProdTimeout
		}

		if prods.Dscv != "" {
			// 服务发现
			dscvName := prods.DscvName
			if dscvName == "" {
				dscvName = name
			}

			prods.discSucc = Dscv.InstMng().SetDiscoveryS(prods.Dscv, dscvName, prods.Set, prods.DscvIdle, true) != nil
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
		conn.Close(true)
	}
}

func (that *server) connOpen(pConn *ANet.Conn) ANet.Client {
	conn := *pConn
	var encryptKey []byte = nil
	fun := that.connOpenFun(pConn, &encryptKey)
	var buffer *KtBuffer.Buffer = nil
	for {
		done, client := fun(that.Manager.Processor().Req(&buffer, conn, encryptKey))
		Util.PutBuffer(buffer)
		buffer = nil
		if done {
			return client
		}
	}
}

func (that *server) ConnPoll(conn ANet.Conn) {
	connPoll := conn.ConnPoll()
	if connPoll == nil {
		Util.GoSubmit(func() {
			Server.ConnLoop(conn)
		})
		return
	}

	var done bool
	var client ANet.Client
	var encryptKey []byte = nil
	fun := that.connOpenFun(&conn, &encryptKey)
	connPoll.FrameStart(func(el interface{}) {
		frame, _ := el.(*ANet.ReqFrame)
		if frame == nil {
			conn.Close(true)
			return
		}

		if !done {
			done, client = fun(Processor.ReqFrame(frame, encryptKey))
			// 内存池回收
			Util.PutBuffer(frame.Buffer)
			return
		}

		if done {
			if client == nil {
				// 内存池回收
				Util.PutBuffer(frame.Buffer)
				conn.Close(true)
				return
			}

			// client.OnReq有内存池回收
			connPoll.FrameReq(client, frame)
			return
		}
	})
}

func (that *server) connOpenFun(pConn *ANet.Conn, pEncryptKey *[]byte) func(err error, req int32, uri string, uriI int32, data []byte) (bool, ANet.Client) {
	conn := *pConn
	var flag int32
	var compress bool
	var encryptKey []byte
	_i := 0
	return func(err error, req int32, uri string, uriI int32, data []byte) (_done bool, _client ANet.Client) {
		_done = true
		if err != nil {
			return
		}

		switch _i {
		case 0:
			// flag
			flag = uriI
			// 连接压缩
			compress = (flag & ANet.FLG_COMPRESS) != 0
			processor := that.Manager.Processor()
			// 连接密钥
			if (flag&ANet.FLG_ENCRYPT) != 0 && processor.Encrypt != nil {
				sKey, cKey := processor.Encrypt.NewKeys()
				if sKey != nil && cKey != nil {
					encryptKey = sKey
					if pEncryptKey != nil {
						*pEncryptKey = encryptKey
					}

					// 连接秘钥
					err = processor.Rep(true, conn, nil, compress, ANet.REQ_KEY, "", 0, encryptKey, false, 0)
					if err != nil {
						return
					}
				}
			}
			// Acl准备
			err = processor.Rep(true, conn, nil, compress, ANet.REQ_ACL, "", 0, encryptKey, false, 0)
			if err != nil {
				return
			}
			break
		case 1:
			// 登录Acl处理
			cid := that.CidGen(compress)
			aclProds := that.GetProds(Config.AclProd)
			aclClient := aclProds.GetProdHash(Config.WorkHash).GetAclClient()
			var login *gw.LoginRep
			login, err = aclClient.Login(aclProds.TimeoutCtx(), &gw.LoginReq{
				Cid:  cid,
				Data: data,
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

				return
			}

			// 登录踢出
			if login.KickData != nil {
				// 登录失败被踢
				that.Manager.Processor().Rep(true, conn, nil, compress, ANet.REQ_KICK, "", 0, login.KickData, false, 0)
				*pConn = nil
				ANet.CloseDelay(conn, Config.KickDrt)
				return
			}

			// 登录失败
			if !login.Succ {
				AZap.Logger.Debug("Serv Login Acl Fail")
				return
			}

			// 客户端注册
			manager := that.Manager
			client := manager.Open(conn, encryptKey, compress, cid)
			clientG := Handler.ClientG(client)
			// 用户状态设置
			clientG.SetId(login.Uid, login.Sid, login.Unique, login.DiscBack)
			if clientG.Gid() != "" {
				// 用户连接保持
				clientG.connKeep()
				clientG.connCheck(nil)
				if clientG.IsClosed() {
					return
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
			uriRoute := uriI
			uriHash := uri
			if uriRoute > 0 {
				if uriHash != UriDict.UriMapHash {
					// 路由缓存
					manager.Processor().Rep(true, conn, nil, compress, ANet.REQ_ROUTE, UriDict.UriMapHash, 0, UriDict.UriMapJsonData, false, 0)
				}
			}

			// 消息处理 单用户登录
			var rep *gw.Id32Rep = nil
			if clientG.Gid() != "" && clientG.unique == "" {
				// 消息队列清理开启
				rep, err = Server.GetProdClient(clientG).GetGWIClient().GQueue(Server.Context, &gw.IGQueueReq{
					Gid:    clientG.gid,
					Cid:    clientG.Id(),
					Unique: clientG.Unique(),
					Clear:  login.Clear,
				})

				if Server.Id32(rep) < R_SUCC_MIN {
					clientG.Close(err, nil)
					return true, nil
				}
			}

			// GLasts管道
			if login.LastsReq != nil {
				rep, err = that.gateway.GLasts(that.Context, login.LastsReq)
				if Server.Id32(rep) < R_SUCC_MIN {
					clientG.Close(err, nil)
					return true, nil
				}
			}

			if login.LastsReqs != nil {
				for _, lastsReqs := range login.LastsReqs {
					rep, err = that.gateway.GLasts(that.Context, lastsReqs)
					if Server.Id32(rep) < R_SUCC_MIN {
						clientG.Close(err, nil)
						return true, nil
					}
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
					return true, nil
				}
			}

			// 登录成功
			client.Get().Rep(true, ANet.REQ_LOOP, strconv.FormatInt(cid, 10)+"/"+clientG.unique+"/"+clientG.gid, 0, login.Data, false, false, 0)
			if client.Get().IsClosed() {
				return
			}

			// 并发限制
			if login.Limit > 0 {
				clientG.SetLimiter(int(login.Limit))
			}

			_client = client
			return
			break
		default:
			return
			break
		}

		if _client == nil {
			_done = false
			_i++
		}
		return
	}
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
	that.gateway = gateway
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
