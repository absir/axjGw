package gateway

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"axjGW/gen/gw"
	"go.uber.org/zap"
	"strings"
	"time"
)

const (
	ERR_PROD_NO  = int32(gw.Result_ProdNo)  // 服务不存在
	ERR_PORD_ERR = int32(gw.Result_ProdErr) // 服务错误
)

var Processor *ANet.Processor
var Handler *handler

func initHandler() {
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

	Handler = new(handler)
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
	if clientG.gid != "" {
		// 断开连接通知
		Server.GetProdClient(clientG).GetGWIClient().Disc(Server.Context, &gw.GDiscReq{
			Cid:    clientG.Id(),
			Gid:    clientG.gid,
			Unique: clientG.unique,
		})
	}

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
	if req >= ANet.REQ_ONEWAY {
		return false
	}

	return true
}

func (that *handler) OnReqIO(client ANet.Client, req int32, uri string, uriI int32, data []byte) {
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

func (that *handler) Processor() *ANet.Processor {
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
