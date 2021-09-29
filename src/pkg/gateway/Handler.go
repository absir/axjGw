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

type ConnG struct {
	ANet.ConnM
	uid    int64     // 用户编号int64
	sid    string    // 用户编号string
	hash   int       // hash值
	rid    int32     // 请求编号
	ridMap *sync.Map // 请求字典
}

func (c *ConnG) SetId(uid int64, sid string) {
	c.uid = uid
	c.sid = sid
	c.hash = -1
}

func (c *ConnG) Hash() int {
	if c.hash < 0 {
		conn := c.ConnC
		conn.Locker().Lock()
		defer conn.Locker().Unlock()

		if c.hash < 0 {
			var hash int
			if c.uid > 0 {
				hash = int(c.uid)

			} else if c.sid != "" {
				hash = Kt.HashCode(KtUnsafe.StringToBytes(c.sid))

			} else {
				hash = int(c.Id())
			}

			if hash < 0 {
				hash = -hash
			}

			c.hash = hash
		}
	}

	return c.hash
}

func (c *ConnG) GetId(name string) int32 {
	if name == "" || name == Config.AclProd {
		return c.rid
	}

	if c.ridMap == nil {
		return 0
	}

	id, _ := c.ridMap.Load(name)
	if id == nil {
		return 0
	}

	return id.(int32)
}

func (c *ConnG) initRidMap() {
	if c.ridMap == nil {
		conn := c.ConnC
		conn.Locker().Lock()
		defer conn.Locker().Unlock()
		if c.ridMap == nil {
			c.ridMap = new(sync.Map)
		}
	}
}

func (c *ConnG) PutRId(name string, id int32) {
	if name == "" || name == Config.AclProd {
		c.rid = id
		return
	}

	if c.ridMap == nil {
		if id <= 0 {
			return
		}
	}

	c.initRidMap()
	if id <= 0 {
		c.ridMap.Delete(name)

	} else {
		c.ridMap.Store(name, id)
	}
}

func (c *ConnG) PutRIds(ids map[string]int32) {
	if ids == nil {
		return
	}

	for name, id := range ids {
		c.PutRId(name, id)
	}
}

func (c *ConnG) GetProd(name string, rand bool) *Prod {
	prods := GetProds(name)
	if prods == nil {
		return nil
	}

	id := c.GetId(name)
	if id > 0 {
		return prods.GetProd(id)
	}

	if rand {
		return prods.GetProdRand()
	}

	return prods.GetProdHash(c.Hash())
}

type HandlerG struct {
}

func (h *HandlerG) ConnG(conn ANet.Conn) *ConnG {
	return conn.(*ConnG)
}

func (h *HandlerG) ConnM(conn ANet.Conn) ANet.ConnM {
	return conn.(*ConnG).ConnM
}

func (h *HandlerG) New() ANet.Conn {
	return new(ConnG)
}

func (h *HandlerG) Init(conn ANet.Conn) {
	connG := h.ConnG(conn)
	connG.ridMap = nil
}

func (h *HandlerG) Open(conn ANet.Conn, client ANet.Client) {
	connG := h.ConnG(conn)
	connG.uid = 0
	connG.sid = ""
	connG.hash = -1
	connG.rid = 0
	connG.ridMap = nil
}

func (h *HandlerG) OnClose(conn ANet.Conn, err error, reason interface{}) {
	connG := h.ConnG(conn)
	connG.ridMap = nil
}

func (h *HandlerG) Last(conn ANet.Conn, req bool) {
}

func (h *HandlerG) OnReq(conn ANet.Conn, req int32, uri string, uriI int32, data []byte) bool {
	if req >= ANet.REQ_ONEWAY {
		return false
	}

	return true
}

func (h *HandlerG) OnReqIO(conn ANet.Conn, req int32, uri string, uriI int32, data []byte) {
	reped := false
	defer h.OnReqErr(conn, req, reped)
	name := Config.AclProd
	if uri[0] == '@' {
		i := strings.IndexByte(uri, '/')
		if i > 0 {
			name = uri[0:i]
			uri = uri[i+1:]
		}
	}

	connG := h.ConnG(conn)
	prod := connG.GetProd(name, false)
	if prod == nil {
		if req > ANet.REQ_ONEWAY {
			// 服务不存在
			reped = true
			conn.Get().Rep(req, "", ERR_PROD_NO, nil, false, false, nil)
		}

		return
	}

	if req > ANet.REQ_ONEWAY {
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

func (h *HandlerG) OnReqErr(conn ANet.Conn, req int32, reped bool) {
	if err := recover(); err != nil {
		AZap.Logger.Warn("rep err", zap.Reflect("err", err))
	}

	if !reped && req > ANet.REQ_ONEWAY {
		conn.Get().Rep(req, "", ERR_PORD_ERR, nil, false, false, nil)
	}
}

func (h *HandlerG) Processor() ANet.Processor {
	return Processor
}

func (h HandlerG) UriDict() ANet.UriDict {
	return UriDict
}
