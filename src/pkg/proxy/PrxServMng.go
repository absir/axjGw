package proxy

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"axjGW/pkg/agent"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"io"
	"net"
	"strconv"
	"sync"
	"time"
)

type prxServMng struct {
	locker   sync.Locker
	Manager  *ANet.Manager
	idWorker *Util.IdWorker
	servMap  *cmap.CMap
}

var PrxServMng = new(prxServMng)

var Processor *ANet.Processor
var Handler = &handler{}
var AclClient gw.AclClient

func (that *prxServMng) Init(wordId int32, Cfg KtCfg.Cfg) {
	initConfig()
	KtCvt.BindInterface(Config, Cfg)
	that.locker = new(sync.Mutex)
	that.idWorker = Util.NewIdWorkerPanic(wordId)
	that.servMap = cmap.NewCMapInit()
	Processor = &ANet.Processor{
		Protocol:    &ANet.ProtocolV{},
		CompressMin: Config.CompressMin,
		DataMax:     Config.DataMax,
	}

	// CompressMin < 0 不压缩
	if Config.CompressMin >= 0 {
		Processor.Compress = &ANet.CompressZip{}
	}

	if Config.Encrypt {
		Processor.Encrypt = &ANet.EncryptSr{}
	}
	that.Manager = ANet.NewManager(Handler, wordId, Config.IdleDrt*int64(time.Millisecond), Config.CheckDrt*time.Millisecond)
	initProtos()
	initPrxMng()
	// Acl服务客户端
	if Config.Acl != "" {
		go func() {
			for {
				client, err := grpc.Dial(Config.Acl, grpc.WithInsecure())
				if err != nil {
					AZap.Logger.Warn("Acl grpc.Dial Err "+Config.Acl, zap.Error(err))
					time.Sleep(Config.AclTry * time.Millisecond)
					continue
				}

				AclClient = gw.NewAclClient(client)
				break
			}
		}()
	}
}

func (that *prxServMng) Start() {
	if Config.Servs != nil {
		for name, serv := range Config.Servs {
			s := StartServ(name, serv.Addr, FindProto(serv.Proto, true), serv.Cfg)
			if s != nil {
				serv.Addr = s.Addr
			}
		}
	}

	go that.Manager.CheckLoop()
	go PrxMng.CheckLoop()
}

func (that *prxServMng) StartGrpc() {
	AZap.Logger.Info("StartGrpc: " + Config.GrpcAddr)
	lis, err := net.Listen("tcp", Config.GrpcAddr)
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
	gw.RegisterGatewayServer(serv, &GatewayS{})
	matchers := KtStr.ForMatchers(Config.GrpcIps, false, true)
	lisIps := ANet.NewListenerIps(lis, func(ip string) bool {
		return KtStr.Matchers(matchers, ip, true)
	})
	go func() {
		if err := serv.Serve(lisIps); err != nil {
			AZap.Logger.Error("Grpc Serve Err "+Config.GrpcAddr, zap.Error(err))
		}
	}()
}

func (that *prxServMng) Accept(tConn *net.TCPConn) {
	var conn ANet.Conn = ANet.NewConnSocket(tConn)
	Util.GoSubmit(func() {
		// 内存池
		var buffer *KtBuffer.Buffer
		client := that.open(&buffer, tConn, &conn)
		// 内存池释放
		Util.PutBuffer(buffer)
		if client != nil {
			client.Get().ReqLoop()

		} else if conn != nil {
			// 连接失败关闭
			conn.Close(true)
		}
	})
}

func (that *prxServMng) open(pBuffer **KtBuffer.Buffer, tConn *net.TCPConn, pConn *ANet.Conn) ANet.Client {
	conn := *pConn
	var encryptKey []byte
	err, req, _, uriI, data := Processor.Req(pBuffer, conn, encryptKey)
	if err != nil {
		conn.Close(true)
		return nil
	}

	if req == agent.REQ_CONN {
		Util.GoSubmit(func() {
			PrxMng.adapConn(uriI, tConn)
		})
		*pConn = nil
		return nil
	}

	flag := uriI
	compress := flag&ANet.FLG_COMPRESS != 0
	// 连接密钥
	if (flag&ANet.FLG_ENCRYPT) != 0 && Processor.Encrypt != nil {
		sKey, cKey := Processor.Encrypt.NewKeys()
		if sKey != nil && cKey != nil {
			encryptKey = sKey
			// 连接秘钥
			err = Processor.Rep(true, conn, nil, compress, ANet.REQ_KEY, "", 0, encryptKey, false, 0)
			if err != nil {
				return nil
			}
		}
	}

	// Acl准备
	err = Processor.Rep(true, conn, nil, compress, ANet.REQ_ACL, "", 0, encryptKey, false, 0)
	// 内存池释放
	Util.PutBuffer(*pBuffer)
	*pBuffer = nil
	err, req, _, uriI, data = Processor.Req(pBuffer, conn, encryptKey)
	cid := that.idWorker.Generate()
	login, err := that.login(cid, data, conn)
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

	// 登录踢出
	if login.KickData != nil {
		// 登录失败被踢
		that.Manager.Processor().Rep(true, conn, nil, compress, ANet.REQ_KICK, "", 0, login.KickData, false, 0)
		*pConn = nil
		ANet.CloseDelay(conn, Config.KickDrt)
		return nil
	}

	if !login.Succ {
		AZap.Logger.Debug("Serv Login Acl Fail")
		return nil
	}

	// 客户端注册
	manager := that.Manager
	client := manager.Open(conn, encryptKey, compress, cid)
	clientG := Handler.ClientG(client)
	// 用户状态设置
	clientG.SetLogin(login)
	// 请求并发限制
	if login.Limit > 0 {
		clientG.SetLimiter(int(login.Limit))
	}

	if clientG.IsRules() {
		// 接受客户端本地映射配置
		client.Get().Rep(true, agent.REQ_RULES, "", 0, login.Data, false, false, 0)
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

	//AZap.Debug("Login gid %s = %d", clientG.gid, cid)
	if clientG.gid != "" {
		PrxServMng.locker.Lock()
		PrxMng.gidMap.Store(clientG.gid, clientG.Id())
		PrxServMng.locker.Unlock()
	}

	return client
}

func (that *prxServMng) login(cid int64, data []byte, conn ANet.Conn) (*gw.LoginRep, error) {
	// 客户端授权
	if Config.ClientKeys != nil && len(Config.ClientKeys) > 0 {
		var strs []string
		json.Unmarshal(data, &strs)
		if len(strs) >= 4 {
			flag := Config.ClientKeys[strs[0]]
			fLen := len(flag)
			if fLen > 0 && flag[0] > '0' {
				rep := &gw.LoginRep{}
				rep.Succ = true
				if flag[1] > '1' {
					gid := strs[2]
					if gid == "" {
						gid = strs[3]
					}

					rep.Sid = gid
				}

				if fLen > 1 && flag[1] > '1' {
					rep.Unique = "*"
				}

				return rep, nil
			}
		}
	}

	if AclClient != nil {
		// Acl服务登录
		return AclClient.Login(Config.AclCtx(), &gw.LoginReq{
			Cid:  cid,
			Data: data,
			Addr: conn.RemoteAddr(),
		})
	}

	return nil, nil
}

type ClientG struct {
	ANet.ClientMng
	gid       string
	unique    string
	discBack  bool
	ruleServs map[string]*RuleServ
}

type RuleServ struct {
	rule *agent.RULE
	serv *PrxServ
}

func (that *ClientG) IsRules() bool {
	return that.unique == "*"
}

func (that *ClientG) SetLogin(login *gw.LoginRep) {
	gid := login.Sid
	if gid == "" {
		gid = strconv.FormatInt(login.Uid, 10)
	}

	that.gid = gid
	that.unique = login.Unique
	that.discBack = login.DiscBack
}
