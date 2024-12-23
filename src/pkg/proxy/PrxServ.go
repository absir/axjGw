package proxy

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"axjGW/pkg/agent"
	"axjGW/pkg/proxy/PProto"
	"go.uber.org/zap"
	"net"
	"strconv"
	"strings"
	"time"
)

type PrxServ struct {
	Name       string
	Addr       string
	TrafficDrt time.Duration
	Proto      PrxProto
	Cfg        interface{}
	cid        int64
	rule       *agent.RULE
	serv       net.Listener
}

func StartServ(name string, addr string, trafficDrt time.Duration, proto PrxProto, cfg map[string]string) *PrxServ {
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
	// 流量上报间隔
	that.TrafficDrt = trafficDrt * time.Second
	// 协议配置
	that.Cfg = that.Proto.NewCfg()
	if cfg != nil {
		KtCvt.BindInterface(that.Cfg, cfg)
	}

	that.Proto.InitCfg(that.Cfg)
	serv, err := net.Listen("tcp", addr)
	if err != nil {
		AZap.Logger.Error("PrxServ Start err", zap.Error(err))
		return nil
	}

	that.serv = serv
	PrxServMng.servMap.Store(addr, that)
	AZap.Logger.Info("PrxServ Start T." + strconv.Itoa((int)(trafficDrt)) + " " + proto.Name() + "://" + addr)
	Util.GoSubmit(func() {
		for !APro.Stopped && that.serv != nil {
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

// 关闭服务
func (that *PrxServ) Close() {
	if that.serv == nil {
		return
	}

	PrxServMng.servMap.Delete(that.Addr)
	that.serv.Close()
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
	var outBuffer *KtBuffer.Buffer
	outBuff := Util.GetBufferBytes(that.Proto.ReadBufferSize(that.Cfg), &outBuffer)
	buffer := Util.GetBuffer(that.Proto.ReadBufferSize(that.Cfg)+that.Proto.ReadBufferMax(that.Cfg), true)
	name := ""
	var size = 0
	var err error
	ctx := that.Proto.ReadServerCtx(that.Cfg, conn)
	for {
		ok := false
		ok, err = that.Proto.ReadServerName(that.Cfg, ctx, buffer, outBuff[:size], &name, conn)
		if err != nil {
			Util.PutBuffer(outBuffer)
			Util.PutBuffer(buffer)
			PrxMng.closeConn(conn, true, err)
			return
		}

		if ok {
			break
		}

		// 读取缓冲数据过大
		if buffer.Len() > that.Proto.ReadBufferMax(that.Cfg) {
			Util.PutBuffer(outBuffer)
			Util.PutBuffer(buffer)
			PrxMng.closeConn(conn, true, err)
			return
		}

		// 二次读取
		size, err = conn.Read(outBuff)
		if err != nil || size <= 0 {
			Util.PutBuffer(outBuffer)
			Util.PutBuffer(buffer)
			PrxMng.closeConn(conn, true, err)
			return
		}
	}

	// 解析代理连接成功
	client, pAddr, gid, sub := that.clientPAddr(name, that.Proto)
	if client == nil || pAddr == "" {
		Util.PutBuffer(outBuffer)
		Util.PutBuffer(buffer)
		PrxMng.closeConn(conn, true, Kt.NewErrReason("NO CLIENT RULE "+that.Name+" - "+name))
		return
	}

	data := buffer.Bytes()
	buffer.Reset()
	id, adap := PrxMng.adapOpen(that, conn, outBuff, outBuffer, ctx, buffer)
	if that.TrafficDrt > 0 && AclClient != nil {
		trafficReq := &gw.TrafficReq{}
		adap.trafficReq = trafficReq
		cid := KtCvt.ToString(client.CId())
		if cid != "" {
			trafficReq.Cid = &cid
		}

		if gid != "" {
			trafficReq.Gid = &gid
		}

		if sub != "" {
			trafficReq.Sub = &sub
		}
	}

	err = client.Get().Rep(true, agent.REQ_CONN, strconv.Itoa(that.Proto.ReadBufferSize(that.Cfg))+"/"+pAddr, id, data, false, false, 0)
	if err != nil {
		adap.Close(err)
		return
	}
}

// 获取代理客户端，代理地址
func (that *PrxServ) clientPAddr(name string, proto PrxProto) (ANet.Client, string, string, string) {
	gid := ""
	sub := ""
	var client ANet.Client = nil
	if !Config.AclMain {
		if len(name) > 0 {
			idx := strings.IndexByte(name, '.')
			str := name
			if idx >= 0 {
				str = name[0:idx]
			}

			strs := strings.SplitN(str, "_", 2)
			gid = strs[0]
			if len(strs) > 1 {
				sub = strs[1]
			}
		}

		if gid == "" && sub == "" {
			var proxy = sProxy
			if proxy != nil && proxy.Rules != nil {
				var rule = proxy.Rules[that.Name]
				if rule != nil && rule.Addr != "" {
					client = PrxMng.Client(rule.Cid, rule.Gid)
					if client != nil {
						return client, rule.Addr, gid, sub
					}
				}
			}

			if that.cid != 0 && that.rule != nil {
				client = PrxServMng.Manager.Client(that.cid)
				if client == nil {
					return nil, "", gid, sub
				}

				return client, that.rule.Addr, gid, sub
			}

		} else if gid != "" {
			client = PrxMng.Client(0, gid)
			if client == nil {
				return nil, "", gid, sub
			}

			clientG := Handler.ClientG(client)
			//AZap.Debug("clientPAddr ruleServs %s = %v", gid, clientG.ruleServs == nil)
			if clientG.ruleServs != nil {
				// 服务名一致，不需要别名
				sName := sub
				if sName == that.Name {
					return client, "", gid, sub

				} else if sName == "" {
					sName = that.Name
				}

				ruleServ := clientG.ruleServs[sName]
				if ruleServ != nil {
					if ruleServ.serv != that {
						return client, "", gid, sub
					}

					// 返回客户端，规则地址
					return client, ruleServ.rule.Addr, gid, sub
				}
			}
		}
	}

	if AclClient != nil {
		rep, err := AclClient.Addr(Config.AclCtx(), &gw.AddrReq{
			Gid:   gid,
			Sub:   sub,
			Name:  name,
			Proto: that.Proto.Name(),
			SName: that.Name,
		})

		if err != nil {
			AZap.Warn("AclClient.Addr Err", zap.Error(err))
		}

		if rep != nil && rep.Addr != "" {
			if client == nil {
				client = PrxMng.Client(rep.Cid, rep.Gid)
			}

			return client, rep.Addr, gid, sub
		}
	}

	return client, "", gid, sub
}
