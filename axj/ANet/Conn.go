package ANet

import (
	"axj/APro"
	"axj/Thrd/AZap"
	"errors"
	"go.uber.org/zap"
	"sync"
)

// 错误
var ERR_CLOSED = errors.New("CLOSED")
var ERR_CRASH = errors.New("CRASH")

type Conn struct {
	client      Client
	locker      sync.Locker
	closed      bool
	encryKey    []byte
	repBuffer   *[]byte
	poolG       APro.PoolG
	handler     Handler
	handlerData interface{}
}

type Handler interface {
	Processor() Processor
	UriDict() UriDict
	Data(conn *Conn) interface{}
	Last(conn *Conn, req bool)
	OnReq(conn *Conn, req int32, uri string, data []byte) bool
	OnReqIO(conn *Conn, req int32, uri string, data []byte)
	OnClose(conn *Conn, err error, reason interface{})
}

func InitConn(conn *Conn, client Client, handler Handler) *Conn {
	if conn == nil {
		conn = new(Conn)
	}

	conn.client = client
	conn.locker = new(sync.Mutex)
	conn.closed = false
	conn.encryKey = nil
	conn.repBuffer = nil
	conn.poolG = nil
	conn.handler = handler
	conn.handlerData = handler.Data(conn)
	return conn
}

func (c *Conn) Close(err error, reason interface{}) {
	if c.closed {
		return
	}

	c.locker.Lock()
	defer c.locker.Unlock()
	if c.closed {
		return
	}

	// 关闭执行
	defer c.Recover()
	defer c.client.Close()
	c.closed = true
	// 关闭日志
	AZap.Logger.Info("Recover crash", zap.Error(err), zap.Reflect("reason", reason))
	if c.handler != nil {
		c.handler.OnClose(c, err, reason)
	}
}

func (c *Conn) Recover() error {
	if reason := recover(); reason != nil {
		err, ok := reason.(error)
		if ok {
			reason = nil
			c.Close(err, nil)

		} else {
			err = ERR_CRASH
			c.Close(err, reason)
		}

		if reason == nil {
			AZap.Logger.Warn("Recover crash", zap.Error(err))

		} else {
			AZap.Logger.Warn("Recover crash", zap.Error(err), zap.Reflect("reason", reason))
		}

		return err
	}

	return nil
}

func (c *Conn) Req() (err error, req int32, uri string, data []byte) {
	if c.closed {
		return ERR_CLOSED, 0, "", nil
	}

	c.handler.Last(c, true)
	processor := c.handler.Processor()
	var uriI int32 = 0
	err, req, uri, uriI, data = Req(c.client, processor.Protocol, processor.Compress, processor.Encrypt, c.encryKey)
	if uri == "" && uriI > 0 {
		uriDict := c.handler.UriDict()
		if uriDict != nil {
			uri = uriDict.UriIMapUri()[uriI]
		}
	}

	return nil, req, uri, data
}

func (c *Conn) ReqLoop() {
	for {
		err, req, uri, data := c.Req()
		if err != nil {
			c.Close(err, nil)
			break
		}

		if !c.handler.OnReq(c, req, uri, data) {
			poolG := c.poolG
			if poolG == APro.PoolOne {
				c.handlerReqIo(nil, req, uri, data)

			} else {
				go c.handlerReqIo(poolG, req, uri, data)
				if poolG != nil {
					poolG.Add()
				}
			}
		}
	}
}

func (c *Conn) handlerReqIo(poolG APro.PoolG, req int32, uri string, data []byte) {
	if poolG == nil {
		defer poolG.Done()
	}

	c.handler.OnReqIO(c, req, uri, data)
}

// 单进程阻塞 req < 0 直接 WriteData
func (c *Conn) Rep(req int32, uri string, data []byte, isolate bool, encry bool, batch *RepBatch) error {
	if c.closed {
		return ERR_CLOSED
	}

	c.handler.Last(c, false)
	uriDict := c.handler.UriDict()
	var uriI int32 = 0
	if uri != "" && uriDict != nil {
		uriI = uriDict.UriMapUriI()[uri]
		if uriI > 0 {
			uri = ""
		}
	}

	encryKey := c.encryKey
	if !encry {
		encryKey = nil
	}

	processor := c.handler.Processor()
	var err error = nil
	if batch == nil {
		err = Rep(c.client, c.repBuffer, processor.Protocol, processor.Compress, processor.CompressMin, processor.Encrypt, encryKey, req, uri, uriI, data, isolate)

	} else {
		err = batch.Rep(c.client, c.repBuffer, processor.Encrypt, encryKey)
	}

	if err != nil {
		c.Close(err, nil)
		return err
	}

	return err
}

func RepBatchHandler(batch *RepBatch, handler Handler, req int32, uri string, data []byte) *RepBatch {
	processor := handler.Processor()
	uriDict := handler.UriDict()
	var uriI int32 = 0
	if uri != "" && uriDict != nil {
		uriI = uriDict.UriMapUriI()[uri]
		if uriI > 0 {
			uri = ""
		}
	}

	batch.Init(processor.Protocol, processor.Compress, processor.CompressMin, req, uri, uriI, data)
	return batch
}

type HandlerW struct {
	handler Handler
}

func (h HandlerW) Processor() Processor {
	return h.handler.Processor()
}

func (h HandlerW) UriDict() UriDict {
	return h.handler.UriDict()
}

func (h HandlerW) Data(conn *Conn) interface{} {
	return h.handler.Data(conn)
}

func (h HandlerW) Last(conn *Conn, req bool) {
	h.handler.Last(conn, req)
}

func (h HandlerW) OnReq(conn *Conn, req int32, uri string, data []byte) bool {
	return h.handler.OnReq(conn, req, uri, data)
}

func (h HandlerW) OnReqIO(conn *Conn, req int32, uri string, data []byte) {
	h.handler.OnReqIO(conn, req, uri, data)
}

func (h HandlerW) OnClose(conn *Conn, err error, reason interface{}) {
	h.handler.OnClose(conn, err, reason)
}
