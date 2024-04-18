package gateway

import (
	"axj/ANet"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axjGW/gen/gw"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	ERR_PROD_NO   = int32(gw.Result_ProdNo)  // 服务不存在
	ERR_PORD_ERR  = int32(gw.Result_ProdErr) // 服务错误
	ERR_PORD_SUCC = int32(gw.Result_Succ)    // 服务成功
)

var Processor ANet.Processor = nil
var Handler *handler = nil

func initHandler() {
	if Processor == nil {
		processor := &ANet.ProcessorV{
			Protocol:    &ANet.ProtocolV{},
			CompressMin: Config.CompressMin,
			DataMax:     Config.DataMax,
		}

		// CompressMin < 0 不压缩
		if Config.CompressMin >= 0 {
			processor.Compress = &ANet.CompressZip{}
		}

		if Config.Encrypt {
			processor.Encrypt = &ANet.EncryptSr{}
		}

		Processor = processor
	}

	if Handler == nil {
		Handler = new(handler)
	}
}

type handler struct {
	checkBuff []interface{}
}

func (that *handler) ClientG(client ANet.Client) *ClientG {
	return client.(*ClientG)
}

func (that *handler) OnOpen(client ANet.Client) {
}

func (that *handler) OnClose(client ANet.Client, err error, reason interface{}) {
	clientG := that.ClientG(client)
	if clientG.discBack {
		Server.GetProds(Config.AclProd).GetProdHash(clientG.Hash()).GetAclClient().DiscBack(Server.Context, &gw.LoginBack{
			Cid:    clientG.Id(),
			Unique: clientG.unique,
			Uid:    clientG.uid,
			Sid:    clientG.sid,
		})
	}
}

func (that *handler) OnKeep(client ANet.Client, req bool) {
}

func (that *handler) OnReq(client ANet.Client, req int32, uri string, uriI int32, data []byte) bool {
	if req >= ANet.REQ_ONEWAY || req == ANet.REQ_READ {
		return false
	}

	return true
}

func (that *handler) OnReqIO(client ANet.Client, req int32, uri string, uriI int32, data []byte) {
	if req == ANet.REQ_READ {
		clientG := that.ClientG(client)
		if clientG.gid != "" {
			Server.GetProdGid(clientG.gid).GetGWIClient().Read(Server.Context, &gw.ReadReq{
				Gid:    clientG.gid,
				Tid:    uri,
				LastId: KtBytes.GetInt64(data, 0, nil),
			})
		}

		return
	}

	reped := false
	pReped := &reped
	defer that.reqRcvr(client, req, pReped)
	name := Config.PassProd
	if uri != "" && uri[0] == '@' {
		i := strings.IndexByte(uri, '/')
		if i > 0 {
			name = uri[0:i]
			uri = uri[i+1:]
		}
	}

	clientG := that.ClientG(client)
	prod, prods := clientG.GetProd(name, false)
	if prod == nil {
		if req > ANet.REQ_ONEWAY {
			// 服务不存在
			*pReped = true
			clientG.Get().Rep(true, req, "", ERR_PROD_NO, nil, false, false, 0)
		}

		return
	}

	if req > ANet.REQ_ONEWAY {
		if len(uri) > 0 {
			code := uri[0]
			if code == '<' || code == '>' {
				// 消息列表
				idx0 := strings.IndexByte(uri, '/')
				idx1 := strings.IndexByte(uri[idx0+1:], '/')
				if idx0 > 0 && idx1 > 0 {
					limit := KtCvt.ToInt32(uri[1:idx0])
					idx0++
					idx1 += idx0
					id := KtCvt.ToInt64(uri[idx0:idx1])
					gid := uri[idx1+1:]
					if clientG.Gid() != gid {
						result, err := Server.GetProdGid(gid).GetGWIClient().GidHasCid(prods.TimeoutCtx(), &gw.GidHasCidReq{
							Gid: gid,
							Cid: clientG.Id(),
						})

						if err != nil {
							*pReped = true
							clientG.Get().Rep(true, req, "", ERR_PORD_ERR, nil, false, false, 0)
							return
						}

						if result == nil || !result.GetVal() {
							*pReped = true
							clientG.Get().Rep(true, req, "", 0, nil, false, false, 0)
							return
						}
					}

					rep := MsgMng().MsgListRep(gid, id, int(limit), code == '>')
					*pReped = true
					clientG.Get().Rep(true, req, "", 0, rep, false, false, 0)
					return
				}
			}
		}

		// 请求返回
		result, err := prod.GetPassClient().Req(prods.TimeoutCtx(), &gw.PassReq{
			Cid:  clientG.Id(),
			Uid:  clientG.uid,
			Sid:  clientG.sid,
			Uri:  uri,
			Data: data,
		})
		if err != nil || result == nil {
			if err == nil {
				AZap.Logger.Warn("Pass Err " + uri + " nil")

			} else {
				AZap.Logger.Warn("Pass Err " + uri + " " + err.Error())
			}

			*pReped = true
			if err == nil {
				clientG.Get().Rep(true, req, "", ERR_PROD_NO, nil, false, false, 0)

			} else {
				clientG.Get().Rep(true, req, "", ERR_PORD_ERR, nil, false, false, 0)
			}

		} else {
			*pReped = true
			clientG.Get().Rep(true, req, "", result.Err, result.Data, false, false, 0)
		}

	} else {
		// 单向发送
		prod.GetPassClient().Send(prods.TimeoutCtx(), &gw.PassReq{
			Cid:  clientG.Id(),
			Uid:  clientG.uid,
			Sid:  clientG.sid,
			Uri:  uri,
			Data: data,
		})
	}
}

func (that *handler) reqRcvr(client ANet.Client, req int32, reped *bool) {
	if err := recover(); err != nil {
		AZap.LoggerS.Warn("Rep Err", zap.Reflect("err", err))
	}

	if !*reped && req > ANet.REQ_ONEWAY {
		client.Get().Rep(true, req, "", ERR_PORD_ERR, nil, false, false, 0)
	}
}

func (that *handler) Processor() ANet.Processor {
	return Processor
}

func (that *handler) UriDict() ANet.UriDict {
	return UriDict
}

func (that *handler) KickDrt() time.Duration {
	return Config.KickDrt
}

func (that *handler) New(conn ANet.Conn) ANet.ClientM {
	clientG := new(ClientG)
	clientG.hash = -1
	return clientG
}

func (that *handler) Check(time int64, client ANet.Client) {
	clientG := that.ClientG(client)
	clientG.GidConn(time, clientG.gid, clientG.gidConn)
	if clientG.gidMap != nil {
		clientG.gidMap.RangeBuff(clientG.GidConnRange, &that.checkBuff, 1024)
	}
}

func (that *handler) CheckDone(time int64) {
}

func CompressorCData(data []byte, cDataP *[]byte, cDidP *bool) {
	if *cDataP != nil || data == nil {
		return
	}

	dLen := len(data)
	processorV := Processor.Get()
	if processorV.Compress == nil || dLen <= 0 || dLen < processorV.CompressMin {
		*cDataP = data
		return
	}

	cData, err := processorV.Compress.Compress(data)
	if cData == nil || err != nil || len(cData) >= dLen {
		if err != nil {
			// 压缩错误
			AZap.Logger.Warn("Msg CData Err", zap.Error(err))
		}

		*cDataP = data
		return
	}

	*cDataP = cData
	*cDidP = true
}
