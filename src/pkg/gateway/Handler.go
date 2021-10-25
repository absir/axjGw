package gateway

import (
	"axj/ANets"
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"context"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
)

const (
	ERR_PROD_NO  = 1 // 服务不存在
	ERR_PORD_ERR = 2 // 服务错误
)

var Processor = ANets.Processor{
	Protocol:    ANets.ProtocolV{},
	Compress:    ANets.CompressZip{},
	CompressMin: Config.CompressMin,
	Encrypt:     ANets.EncryptSr{},
	DataMax:     Config.DataMax,
}

type ConnG struct {
	ANets.ConnM
	uid    int64     // 用户编号int64
	sid    string    // 用户编号string
	hash   int       // hash值
	rid    int32     // 请求编号
	ridMap *sync.Map // 请求字典
}

func (that ConnG) PInit() {
	that.ConnM.PInit()
	that.uid = 0
	that.sid = ""
	that.hash = -1
	that.rid = 0
	that.ridMap = nil
}

func (that ConnG) PRelease() bool {
	if that.ConnM.PRelease() {
		//that.uid = 0
		that.sid = ""
		//that.hash = -1
		//that.rid = 0
		//that.ridMap = nil
		return true
	}

	return false
}

func (that ConnG) SetId(uid int64, sid string) {
	that.uid = uid
	if uid > 0 {
		sid = strconv.FormatInt(uid, 10)
	}
	
	that.sid = sid
	that.hash = -1
}

func (that ConnG) Hash() int {
	if that.hash < 0 {
		conn := that.ConnC
		conn.Locker().Lock()
		defer conn.Locker().Unlock()
		if that.hash < 0 {
			var hash int
			if that.uid > 0 {
				hash = int(that.uid)

			} else if that.sid != "" {
				hash = Kt.HashCode(KtUnsafe.StringToBytes(that.sid))

			} else {
				hash = int(that.Id())
			}

			if hash < 0 {
				hash = -hash
			}

			that.hash = hash
		}
	}

	return that.hash
}

func (that ConnG) GetId(name string) int32 {
	if name == "" || name == Config.AclProd {
		return that.rid
	}

	if that.ridMap == nil {
		return 0
	}

	id, _ := that.ridMap.Load(name)
	if id == nil {
		return 0
	}

	return id.(int32)
}

func (that ConnG) initRidMap() {
	if that.ridMap == nil {
		conn := that.ConnC
		conn.Locker().Lock()
		defer conn.Locker().Unlock()
		if that.ridMap == nil {
			that.ridMap = new(sync.Map)
		}
	}
}

func (that ConnG) PutRId(name string, id int32) {
	if name == "" || name == Config.AclProd {
		that.rid = id
		return
	}

	if that.ridMap == nil {
		if id <= 0 {
			return
		}
	}

	that.initRidMap()
	if id <= 0 {
		that.ridMap.Delete(name)

	} else {
		that.ridMap.Store(name, id)
	}
}

func (that ConnG) PutRIds(ids map[string]int32) {
	if ids == nil {
		return
	}

	for name, id := range ids {
		that.PutRId(name, id)
	}
}

func (that ConnG) GetProd(name string, rand bool) *Prod {
	prods := GetProds(name)
	if prods == nil {
		return nil
	}

	id := that.GetId(name)
	if id > 0 {
		return prods.GetProd(id)
	}

	if rand {
		return prods.GetProdRand()
	}

	return prods.GetProdHash(that.Hash())
}

type HandlerG struct {
}

func (that HandlerG) ConnG(conn ANets.Conn) *ConnG {
	return conn.(*ConnG)
}

func (that HandlerG) ConnM(conn ANets.Conn) ANets.ConnM {
	return conn.(*ConnG).ConnM
}

func (that HandlerG) New() ANets.Conn {
	return new(ConnG)
}

func (that HandlerG) Open(conn ANets.Conn, client ANets.Client) {
	connG := that.ConnG(conn)
	connG.uid = 0
	connG.sid = ""
	connG.hash = -1
	connG.rid = 0
	connG.ridMap = nil
}

func (that HandlerG) OnClose(conn ANets.Conn, err error, reason interface{}) {
}

func (that HandlerG) Last(conn ANets.Conn, req bool) {
}

func (that HandlerG) OnReq(conn ANets.Conn, req int32, uri string, uriI int32, data []byte) bool {
	if req >= ANets.REQ_ONEWAY {
		return false
	}

	return true
}

func (that HandlerG) OnReqIO(conn ANets.Conn, req int32, uri string, uriI int32, data []byte) {
	reped := false
	defer that.OnReqErr(conn, req, reped)
	name := Config.AclProd
	if uri[0] == '@' {
		i := strings.IndexByte(uri, '/')
		if i > 0 {
			name = uri[0:i]
			uri = uri[i+1:]
		}
	}

	connG := that.ConnG(conn)
	prod := connG.GetProd(name, false)
	if prod == nil {
		if req > ANets.REQ_ONEWAY {
			// 服务不存在
			reped = true
			conn.Get().Rep(req, "", ERR_PROD_NO, nil, false, false, nil)
		}

		return
	}

	if req > ANets.REQ_ONEWAY {
		// 请求返回
		bs, err := prod.GetPassClient().Req(context.Background(), connG.Id(), connG.uid, connG.sid, uri, data)
		if err != nil {
			panic(err)

		} else {
			reped = true
			conn.Get().Rep(req, "", ERR_PROD_NO, bs, false, false, nil)
		}

	} else {
		// 单向发送
		prod.GetPassClient().Send(context.Background(), connG.Id(), connG.uid, connG.sid, uri, data)
	}
}

func (that HandlerG) OnReqErr(conn ANets.Conn, req int32, reped bool) {
	if err := recover(); err != nil {
		AZap.Logger.Warn("rep err", zap.Reflect("err", err))
	}

	if !reped && req > ANets.REQ_ONEWAY {
		conn.Get().Rep(req, "", ERR_PORD_ERR, nil, false, false, nil)
	}
}

func (that HandlerG) Processor() ANets.Processor {
	return Processor
}

func (that HandlerG) UriDict() ANets.UriDict {
	return UriDict
}
