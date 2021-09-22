package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"context"
	"go.uber.org/zap"
	"strings"
	"sync"
)

const (
	ERR_PROD_NO  = 1 // 服务不存在
	ERR_PORD_ERR = 2 // 服务错误
)

var Processor = ANet.Processor{
	Protocol:    ANet.ProtocolV{},
	Compress:    ANet.CompressZip{},
	CompressMin: Config.CompressMin,
	Encrypt:     ANet.EncryptSr{},
	DataMax:     Config.DataMax,
}

type Handler struct {
}

type Data struct {
	uid    int64     // 用户编号int64
	sid    string    // 用户编号string
	hash   int       // hash值
	rid    int32     // 请求编号
	ridMap *sync.Map // 请求字典
}

func GetMData(conn *ANet.Conn) *ANet.Mng {
	return conn.MData().(*ANet.Mng)
}

func GetHData(conn *ANet.Conn) *Data {
	return GetMData(conn).HData().(*Data)
}

func (h Handler) Data(conn *ANet.Conn) interface{} {
	data := new(Data)
	data.uid = 0
	data.sid = ""
	data.hash = -1
	data.rid = 0
	data.ridMap = nil
	return data
}

func (d *Data) SetId(uid int64, sid string) {
	d.uid = uid
	d.sid = sid
	d.hash = -1
}

func (d *Data) Hash(conn *ANet.Conn) int {
	if d.hash < 0 {
		if conn != nil {
			conn.Locker().Lock()
			defer conn.Locker().Unlock()
		}

		if d.hash < 0 {
			var hash int
			if d.uid > 0 {
				hash = int(d.uid)

			} else if d.sid != "" {
				hash = Kt.HashCode(KtUnsafe.StringToBytes(d.sid))

			} else {
				hash = KtUnsafe.PointerHash(d)
			}

			if hash < 0 {
				hash = -hash
			}

			d.hash = hash
		}
	}

	return d.hash
}

func (d *Data) GetId(name string) int32 {
	if name == "" || name == Config.accProd {
		return d.rid
	}

	if d.ridMap == nil {
		return 0
	}

	id, _ := d.ridMap.Load(name)
	if id == nil {
		return 0
	}

	return id.(int32)
}

func (d *Data) initRidMap(conn *ANet.Conn) {
	if d.ridMap == nil {
		conn.Locker().Lock()
		defer conn.Locker().Unlock()
		if d.ridMap == nil {
			d.ridMap = new(sync.Map)
		}
	}
}

func (d *Data) PutId(conn *ANet.Conn, name string, id int32) {
	if name == "" || name == Config.accProd {
		d.rid = id
		return
	}

	if d.ridMap == nil {
		if id <= 0 {
			return
		}
	}

	d.initRidMap(conn)
	if id <= 0 {
		d.ridMap.Delete(name)

	} else {
		d.ridMap.Store(name, id)
	}
}

func (d *Data) GetProd(conn *ANet.Conn, name string, rand bool) *Prod {
	prods := GetProds(name)
	if prods == nil {
		return nil
	}

	id := d.GetId(name)
	if id > 0 {
		return prods.GetProd(id)
	}

	if rand {
		return prods.GetProdRand()
	}

	return prods.GetProdHash(d.Hash(conn))
}

func (h Handler) Last(conn *ANet.Conn, req bool) {
}

func (h Handler) OnReq(conn *ANet.Conn, req int32, uri string, uriI int32, data []byte) bool {
	if req >= ANet.REQ_ONEWAY {
		return false
	}

	return true
}

func (h Handler) OnReqIO(conn *ANet.Conn, req int32, uri string, uriI int32, data []byte) {
	reped := false
	defer h.OnReqErr(conn, req, reped)
	name := Config.accProd
	if uri[0] == '@' {
		i := strings.IndexByte(uri, '/')
		if i > 0 {
			name = uri[0:i]
			uri = uri[i+1:]
		}
	}

	hData := GetHData(conn)
	prod := hData.GetProd(conn, name, false)
	if prod == nil {
		if req > ANet.REQ_ONEWAY {
			// 服务不存在
			reped = true
			conn.Rep(req, "", ERR_PROD_NO, nil, false, false, nil)
		}

		return
	}

	if req > ANet.REQ_ONEWAY {
		// 请求返回
		bs, err := prod.GetPassClient().Req(context.Background(), GetMData(conn).Id(), hData.uid, hData.sid, uri, data)
		if err != nil {
			panic(err)

		} else {
			reped = true
			conn.Rep(req, "", ERR_PROD_NO, bs, false, false, nil)
		}

	} else {
		// 单向发送
		prod.GetPassClient().Send(context.Background(), GetMData(conn).Id(), hData.uid, hData.sid, uri, data)
	}
}

func (h Handler) OnReqErr(conn *ANet.Conn, req int32, reped bool) {
	if err := recover(); err != nil {
		AZap.Logger.Warn("rep err", zap.Reflect("err", err))
	}

	if !reped && req > ANet.REQ_ONEWAY {
		conn.Rep(req, "", ERR_PORD_ERR, nil, false, false, nil)
	}
}

func (h Handler) OnClose(conn *ANet.Conn, err error, reason interface{}) {
}

func (h Handler) Processor() ANet.Processor {
	return Processor
}

func (h Handler) UriDict() ANet.UriDict {
	return nil
}
