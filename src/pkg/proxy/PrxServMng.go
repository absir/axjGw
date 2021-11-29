package proxy

import (
	"axj/ANet"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"axjGW/pkg/agent"
	"axjGW/pkg/gateway"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"time"
)

type prxServMng struct {
	Manager  *ANet.Manager
	idWorker *Util.IdWorker
	servMap  *cmap.CMap
}

var PrxServMng = new(prxServMng)

var Processor *ANet.Processor
var Handler = &handler{}

func (that *prxServMng) Init(wordId int32, Cfg KtCfg.Cfg) {
	initConfig()
	KtCvt.BindInterface(Config, Cfg)
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
}

func (that *prxServMng) Start() {
	go that.Manager.CheckLoop()
	go PrxMng.CheckLoop()
}

func (that *prxServMng) Accept(tConn *net.TCPConn) {
	var conn ANet.Conn = ANet.NewConnSocket(tConn, Config.SocketOut)
	Util.GoSubmit(func() {
		client := that.open(tConn, &conn)
		if client != nil {
			client.Get().ReqLoop()

		} else if conn != nil {
			// 连接失败关闭
			conn.Close()
		}
	})
}

func (that *prxServMng) open(tConn *net.TCPConn, pConn *ANet.Conn) ANet.Client {
	conn := *pConn
	var encryptKey []byte
	err, req, _, uriI, data := Processor.Req(conn, encryptKey)
	if err != nil {
		conn.Close()
		return nil
	}

	if req == agent.REQ_CONN {
		Util.GoSubmit(func() {
			PrxMng.adapConn(uriI, tConn)
		})
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
			err = Processor.Rep(nil, true, conn, nil, compress, ANet.REQ_KEY, "", 0, encryptKey, false, 0)
			if err != nil {
				return nil
			}
		}
	}

	// Acl准备
	err = Processor.Rep(nil, true, conn, nil, compress, ANet.REQ_ACL, "", 0, encryptKey, false, 0)
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
		that.Manager.Processor().Rep(nil, true, conn, nil, compress, ANet.REQ_KICK, "", 0, login.KickData, false, 0)
		*pConn = nil
		ANet.CloseDelay(conn, Config.KickDrt)
		return nil
	}

	// 客户端注册
	manager := that.Manager
	client := manager.Open(conn, encryptKey, compress, (flag&ANet.FLG_OUT) != 0, cid)
	clientG := Handler.ClientG(client)
	// 用户状态设置
	clientG.SetLogin(login)
	// 请求并发限制
	if login.Limit > 0 {
		clientG.SetLimiter(int(login.Limit))
	}

	// 注册成功回调
	if login.Back {
		ok, err := that.loginBack(cid, login.Unique, login.Uid, login.Sid)
		if !ok {
			clientG.Close(err, nil)
			return nil
		}
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

	return client
}

func (that *prxServMng) login(cid int64, data []byte, conn ANet.Conn) (*gw.LoginRep, error) {
	// 客户端授权
	return nil, nil
}

func (that *prxServMng) loginBack(cid int64, unique string, uid int64, sid string) (bool, error) {
	// 客户端鉴权
	return true, nil
}

type ClientG struct {
	ANet.ClientMng
	gid    string
	unique string
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
}

type handler struct {
}

func (h handler) ClientG(client ANet.Client) *ClientG {
	return client.(*ClientG)
}

func (h handler) OnOpen(client ANet.Client) {
}

func (h handler) OnClose(client ANet.Client, err error, reason interface{}) {
}

func (h handler) OnKeep(client ANet.Client, req bool) {
}

func (h handler) OnReq(client ANet.Client, req int32, uri string, uriI int32, data []byte) bool {
	if req > ANet.REQ_ONEWAY && req != agent.REQ_RULES {
		return false
	}

	return true
}

func (h handler) OnReqIO(client ANet.Client, req int32, uri string, uriI int32, data []byte) {
	clientG := Handler.ClientG(client)
	if req == agent.REQ_RULES && clientG.IsRules() {
		// 接受本地映射配置
		var rules map[string]*agent.RULE
		json.Unmarshal(data, &rules)
		if rules != nil {
			for name, rule := range rules {
				proto := FindProto(rule.Proto, true)
				if proto == nil {
					continue
				}

				serv := StartServ(name, rule.Addr, proto)
				serv.Update(proto, clientG.Id(), rule.Addr)
			}
		}

		return
	}

	clientG.Get().Rep(true, req, "", gateway.ERR_PROD_NO, nil, false, false, 0)
}

func (h handler) Processor() *ANet.Processor {
	return Processor
}

func (h handler) UriDict() ANet.UriDict {
	return nil
}

func (h handler) KickDrt() time.Duration {
	return 0
}

func (h handler) New(conn ANet.Conn) ANet.ClientM {
	clientG := new(ClientG)
	return clientG
}

func (h handler) Check(time int64, client ANet.Client) {
}

func (h handler) CheckDone(time int64) {
}
