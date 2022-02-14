package proxy

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"axjGW/pkg/agent"
	"axjGW/pkg/gateway"
	"encoding/json"
	"time"
)

type handler struct {
}

func (h handler) ClientG(client ANet.Client) *ClientG {
	return client.(*ClientG)
}

func (h handler) OnOpen(client ANet.Client) {
}

func (h handler) OnClose(client ANet.Client, err error, reason interface{}) {
	clientG := Handler.ClientG(client)
	if clientG.gid != "" {
		val, _ := PrxMng.gidMap.LoadAndDelete(clientG.gid)
		oId, _ := val.(int64)
		//AZap.Debug("OnClose %s = %d : %d", clientG.gid, clientG.Id(), oId)
		if oId != 0 && oId != clientG.Id() {
			PrxServMng.locker.Lock()
			_, ok := PrxMng.gidMap.Load(clientG.gid)
			//AZap.Debug("OnClose LoadCurr %s = %v", ok)
			if !ok {
				clientG.discBack = false
				PrxMng.gidMap.Store(clientG.gid, oId)
			}

			PrxServMng.locker.Unlock()
		}
	}

	// 断开回调
	if clientG.discBack && AclClient != nil {
		Util.GoSubmit(func() {
			AclClient.DiscBack(Config.AclCtx(), &gw.LoginBack{
				Cid:    clientG.Id(),
				Unique: clientG.unique,
				Sid:    clientG.gid,
			})
		})
	}
}

func (h handler) OnKeep(client ANet.Client, req bool) {
}

func (h handler) OnReq(client ANet.Client, req int32, uri string, uriI int32, data []byte) bool {
	if req > ANet.REQ_ONEWAY || req == agent.REQ_RULES {
		return false
	}

	if req == agent.REQ_CLOSED {
		PrxMng.adapClose(uriI)

	} else if req == agent.REQ_DIAL {
		PrxMng.DialRep(uriI, true)

	} else if req == agent.REQ_DIAL_ERR {
		PrxMng.DialRep(uriI, false)
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
			clientG.ruleServs = map[string]*RuleServ{}
			for name, rule := range rules {
				if rule.Addr == "" {
					continue
				}

				var serv *PrxServ = nil
				var proto PrxProto = nil
				servName := ""
				if rule.Serv != "" {
					servName = rule.Serv

				} else if rule.Proto != "" && rule.Port != "" {
					proto = FindProto(rule.Proto, true)
					if proto == nil {
						continue
					}

					serv = StartServ(name, rule.Port, proto, rule.Cfg)

				} else {
					servName = name
				}

				if servName != "" {
					s := Config.Servs[servName]
					if s != nil {
						val, _ := PrxServMng.servMap.Load(s.Addr)
						serv, _ = val.(*PrxServ)
					}

					if serv == nil {
						AZap.Logger.Warn("Serv No Name " + servName)
					}
				}

				if serv == nil {
					continue
				}

				// 添加协议路由规则
				clientG.ruleServs[name] = &RuleServ{
					rule: rule,
					serv: serv,
				}

				serv.Rule(proto, clientG, name, rule)
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
