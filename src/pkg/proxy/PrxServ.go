package proxy

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"axjGW/pkg/agent"
	"axjGW/pkg/proxy/PProto"
	"bytes"
	"go.uber.org/zap"
	"net"
	"strconv"
	"strings"
)

type PrxServ struct {
	Name  string
	Addr  string
	Proto PrxProto
	Cfg   interface{}
	cid   int64
	rule  *agent.RULE
}

func StartServ(name string, addr string, proto PrxProto, cfg map[string]string) *PrxServ {
	if addr == "" || proto == nil {
		return nil
	}

	if addr[0] >= '0' && addr[0] <= '9' && KtCvt.ToInt32(addr) > 0 {
		addr = ":" + addr
	}

	val, _ := PrxServMng.servMap.Load(addr)
	if val != nil {
		that, _ := val.(*PrxServ)
		return that
	}

	that := new(PrxServ)
	that.Name = name
	that.Addr = addr
	that.Proto = proto
	// 协议配置
	that.Cfg = that.Proto.NewCfg()
	if cfg != nil {
		KtCvt.BindInterface(that.Cfg, cfg)
	}

	serv, err := net.Listen("tcp", addr)
	if err != nil {
		AZap.Logger.Error("PrxServ Start err", zap.Error(err))
		return nil
	}

	PrxServMng.servMap.Store(addr, that)
	AZap.Logger.Info("PrxServ Start " + proto.Name() + "://" + addr)
	Util.GoSubmit(func() {
		for !APro.Stopped {
			conn, err := serv.Accept()
			if err != nil {
				if APro.Stopped {
					return
				}

				AZap.Logger.Warn("PrxServ Accept Err", zap.Error(err))
				continue
			}

			Util.GoSubmit(func() {
				that.accept(conn.(*net.TCPConn))
			})
		}
	})

	return that
}

func (that *PrxServ) Rule(proto PrxProto, clientG *ClientG, name string, rule *agent.RULE) {
	if proto != nil && that.Proto != proto {
		AZap.Logger.Warn("PrxServ Rule err " + that.Proto.Name() + " => " + proto.Name())
		return
	}

	sName := clientG.gid
	if sName == "" {
		sName = strconv.FormatInt(clientG.Id(), 10)
	}

	// 服务名一致，不需要别名
	if name != that.Name {
		sName = sName + "_" + name
	}

	// 提示映射规则
	desc := that.Proto.ServAddr(that.Cfg, sName)
	if desc == "" {
		desc = sName + that.Addr
	}

	desc = that.Proto.Name() + "://" + desc + " => " + rule.Addr
	AZap.Logger.Info("PrxServ Rule " + desc)
	// 发送给客户端
	clientG.Rep(true, agent.REQ_ON_RULE, desc, 0, nil, false, false, 0)
	if _, ok := proto.(*PProto.Tcp); ok {
		that.cid = clientG.Id()
		that.rule = rule
	}
}

func (that *PrxServ) accept(conn *net.TCPConn) {
	buff := make([]byte, that.Proto.ReadBufferSize(that.Cfg))
	buffer := &bytes.Buffer{}
	name := ""
	var size = 0
	var err error
	ctx := that.Proto.ReadServerCtx(that.Cfg, conn)
	for {
		ok := false
		ok, err = that.Proto.ReadServerName(that.Cfg, ctx, buffer, buff[:size], &name, conn)
		if err != nil {
			PrxMng.closeConn(conn, true, err)
			return
		}

		if ok {
			break
		}

		// 读取缓冲数据过大
		if buffer.Len() > that.Proto.ReadBufferMax(that.Cfg) {
			PrxMng.closeConn(conn, true, err)
			return
		}

		// 二次读取
		size, err = conn.Read(buff)
		if err != nil || size <= 0 {
			PrxMng.closeConn(conn, true, err)
			return
		}
	}

	// 解析代理连接成功
	client, pAddr := that.clientPAddr(name, that.Proto)
	if client == nil || pAddr == "" {
		PrxMng.closeConn(conn, true, Kt.NewErrReason("NO CLIENT RULE "+that.Name+" - "+name))
		return
	}

	data := buffer.Bytes()
	buffer.Reset()
	id, adap := PrxMng.adapOpen(that, conn, buff, ctx, buffer)
	err = client.Get().Rep(true, agent.REQ_CONN, pAddr, id, data, false, false, 0)
	if err != nil {
		adap.Close(err)
		return
	}
}

// 获取代理客户端，代理地址
func (that *PrxServ) clientPAddr(name string, proto PrxProto) (ANet.Client, string) {
	idx := strings.IndexByte(name, '.')
	str := name
	if idx >= 0 {
		str = name[0:idx]
	}

	strs := strings.SplitN(str, "_", 2)
	gid := strs[0]
	sub := ""
	if len(strs) > 1 {
		sub = strs[1]
	}

	var client ANet.Client = nil
	if gid == "" && sub == "" {
		if that.cid > 0 && that.rule != nil {
			client = PrxServMng.Manager.Client(that.cid)
			if client == nil {
				return nil, ""
			}

			return client, that.rule.Addr
		}

	} else if gid != "" {
		val, _ := PrxMng.gidMap.Load(gid)
		cid, _ := val.(int64)
		if cid <= 0 {
			cid = KtCvt.ToInt64(gid)
			if cid <= 0 {
				return nil, ""
			}
		}

		client = PrxServMng.Manager.Client(cid)
		if client == nil {
			return nil, ""
		}

		clientG := Handler.ClientG(client)
		if clientG.ruleServs != nil {
			// 服务名一致，不需要别名
			sName := sub
			if sName == that.Name {
				return client, ""

			} else if sName == "" {
				sName = that.Name
			}

			ruleServ := clientG.ruleServs[sName]
			if ruleServ != nil {
				if ruleServ.serv != that {
					return client, ""
				}

				// 返回客户端，规则地址
				return client, ruleServ.rule.Addr
			}
		}
	}

	if AclClient != nil {
		rep, err := AclClient.Addr(Config.AclCtx(), &gw.AddrReq{
			Gid:   gid,
			Sub:   sub,
			Name:  name,
			Proto: that.Proto.Name(),
		})

		if err != nil {
			AZap.Warn("AclClient.Addr Err", zap.Error(err))
		}

		if rep != nil && rep.Addr != "" {
			return client, rep.Addr
		}
	}

	return client, ""
}
