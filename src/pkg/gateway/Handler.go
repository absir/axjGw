package gateway

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"context"
	"go.uber.org/zap"
	"strings"
)

const (
	ERR_PROD_NO  = 1 // 服务不存在
	ERR_PORD_ERR = 2 // 服务错误
)

var processor *ANet.Processor
var Handler *handler

func initHandler() {
	processor = &ANet.Processor{
		Protocol:    &ANet.ProtocolV{},
		Compress:    &ANet.CompressZip{},
		CompressMin: Config.CompressMin,
		DataMax:     Config.DataMax,
	}

	if Config.Encrypt {
		processor.Encrypt = &ANet.EncryptSr{}
	}

	Handler = new(handler)
}

type handler struct {
}

func (that *handler) ClientG(client ANet.Client) *ClientG {
	return client.(*ClientG)
}

func (that *handler) OnOpen(client ANet.Client) {
	clientG := new(ClientG)
	clientG.ConnKeep()
}

func (that *handler) OnClose(client ANet.Client, err error, reason interface{}) {
	clientG := new(ClientG)
	if clientG.gid != "" {
		// 断开连接通知
		Server.GetProdClient(clientG).GetGWIClient().Disc(Server.Context, clientG.Id(), clientG.gid, clientG.unique, 0)
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
	defer that.reqRcvr(client, req, reped)
	name := Config.AclProd
	if uri[0] == '@' {
		i := strings.IndexByte(uri, '/')
		if i > 0 {
			name = uri[0:i]
			uri = uri[i+1:]
		}
	}

	clientG := that.ClientG(client)
	prod := clientG.GetProd(name, false)
	if prod == nil {
		if req > ANet.REQ_ONEWAY {
			// 服务不存在
			reped = true
			clientG.Get().Rep(true, req, "", ERR_PROD_NO, nil, false, false, 0)
		}

		return
	}

	if req > ANet.REQ_ONEWAY {
		// 请求返回
		bs, err := prod.GetPassClient().Req(context.Background(), clientG.Id(), clientG.uid, clientG.sid, uri, data)
		if err != nil {
			panic(err)

		} else {
			reped = true
			clientG.Get().Rep(true, req, "", ERR_PROD_NO, bs, false, false, 0)
		}

	} else {
		// 单向发送
		prod.GetPassClient().Send(context.Background(), clientG.Id(), clientG.uid, clientG.sid, uri, data)
	}
}

func (that *handler) reqRcvr(client ANet.Client, req int32, reped bool) {
	if err := recover(); err != nil {
		AZap.Logger.Warn("rep err", zap.Reflect("err", err))
	}

	if !reped && req > ANet.REQ_ONEWAY {
		client.Get().Rep(true, req, "", ERR_PORD_ERR, nil, false, false, 0)
	}
}

func (that *handler) Processor() *ANet.Processor {
	return processor
}

func (that *handler) UriDict() ANet.UriDict {
	return UriDict
}

func (that *handler) New(conn ANet.Conn) ANet.ClientM {
	clientG := new(ClientG)
	clientG.hash = -1
	return clientG
}

func (that *handler) Check(time int64, client ANet.Client) {
	clientG := new(ClientG)
	if clientG.connTime < time {
		clientG.ConnKeep()
		go clientG.ConnCheck()
	}
}
