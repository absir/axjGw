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

type Conn interface {
	Get() *ConnC
}

type ConnC struct {
	client    Client
	locker    sync.Locker
	encryKey  []byte
	repBuffer *[]byte
	poolG     APro.PoolG
	manager   Manager
}

func (c *ConnC) Get() *ConnC {
	return c
}

func (c *ConnC) Client() Client {
	return c.client
}

func (c *ConnC) Locker() sync.Locker {
	return c.locker
}

func (c *ConnC) Closed() bool {
	return c.client == nil
}

func (c *ConnC) Manager() Manager {
	return c.manager
}

func (c *ConnC) SetEncryKey(encryKey []byte) {
	c.encryKey = encryKey
}

type Handler interface {
	New() Conn
	Init(conn Conn)
	Open(conn Conn, client Client)
	OnClose(conn Conn, err error, reason interface{})
	Last(conn Conn, req bool)
	OnReq(conn Conn, req int32, uri string, uriI int32, data []byte) bool
	OnReqIO(conn Conn, req int32, uri string, uriI int32, data []byte)
	Processor() Processor
	UriDict() UriDict
}

type Manager interface {
	Handler
}

func (c *ConnC) Init() {
	c.client = nil
	c.locker = new(sync.Mutex)
	c.encryKey = nil
	c.repBuffer = nil
	c.poolG = nil
	c.manager = nil
}

func (c *ConnC) Open(client Client, manager Manager) {
	c.client = client
	// c.poolG = nil
	c.manager = manager
}

func (c *ConnC) Close(err error, reason interface{}) {
	if c.Closed() {
		return
	}

	c.locker.Lock()
	defer c.locker.Unlock()
	if c.Closed() {
		return
	}

	client := c.client
	// 关闭执行
	defer c.Recover()
	defer client.Close()
	c.client = nil
	// c.poolG = nil
	c.manager = nil
	// 关闭日志
	AZap.Logger.Info("Conn close", zap.Error(err), zap.Reflect("reason", reason))
	if c.manager != nil {
		c.manager.OnClose(c, err, reason)
	}
}

func (c *ConnC) Recover() error {
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
			AZap.Logger.Warn("Conn crash", zap.Error(err))

		} else {
			AZap.Logger.Warn("Conn crash", zap.Error(err), zap.Reflect("reason", reason))
		}

		return err
	}

	return nil
}

func (c *ConnC) Req() (err error, req int32, uri string, uriI int32, data []byte) {
	if c.Closed() {
		return ERR_CLOSED, 0, "", 0, nil
	}

	c.manager.Last(c, true)
	processor := c.manager.Processor()
	err, req, uri, uriI, data = Req(c.client, processor.Protocol, processor.Compress, processor.Encrypt, c.encryKey, processor.DataMax)
	if uri == "" && uriI > 0 {
		uriDict := c.manager.UriDict()
		if uriDict != nil {
			uri = uriDict.UriIMapUri()[uriI]
		}
	}

	return
}

func (c *ConnC) ReqLoop() {
	for {
		err, req, uri, uriI, data := c.Req()
		if err != nil {
			c.Close(err, nil)
			break
		}

		if !c.manager.OnReq(c, req, uri, uriI, data) {
			poolG := c.poolG
			if poolG == APro.PoolOne {
				c.handlerReqIo(nil, req, uri, uriI, data)

			} else {
				go c.handlerReqIo(poolG, req, uri, uriI, data)
				if poolG != nil {
					poolG.Add()
				}
			}
		}
	}
}

func (c *ConnC) handlerReqIo(poolG APro.PoolG, req int32, uri string, uriI int32, data []byte) {
	if poolG == nil {
		defer poolG.Done()
	}

	c.manager.OnReqIO(c, req, uri, uriI, data)
}

// 单进程阻塞 req < 0 直接 WriteData
func (c *ConnC) Rep(req int32, uri string, uriI int32, data []byte, isolate bool, encry bool, batch *RepBatch) error {
	if c.Closed() {
		return ERR_CLOSED
	}

	c.manager.Last(c, false)
	uriDict := c.manager.UriDict()
	if uriI <= 0 {
		if uri != "" && uriDict != nil {
			uriI = uriDict.UriMapUriI()[uri]
			if uriI > 0 {
				uri = ""
			}
		}
	}

	encryKey := c.encryKey
	if !encry {
		encryKey = nil
	}

	processor := c.manager.Processor()
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

func RepBatchHandler(batch *RepBatch, manager Manager, req int32, uri string, data []byte) *RepBatch {
	processor := manager.Processor()
	uriDict := manager.UriDict()
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

func (h HandlerW) New() Conn {
	return h.handler.New()
}

func (h HandlerW) Init(conn Conn) {
	h.handler.Init(conn)
}

func (h HandlerW) Open(conn Conn, client Client) {
	h.handler.Open(conn, client)
}

func (h HandlerW) OnClose(conn Conn, err error, reason interface{}) {
	h.handler.OnClose(conn, err, reason)
}

func (h HandlerW) Last(conn Conn, req bool) {
	h.handler.Last(conn, req)
}

func (h HandlerW) OnReq(conn Conn, req int32, uri string, uriI int32, data []byte) bool {
	return h.handler.OnReq(conn, req, uri, uriI, data)
}

func (h HandlerW) OnReqIO(conn Conn, req int32, uri string, uriI int32, data []byte) {
	h.handler.OnReqIO(conn, req, uri, uriI, data)
}

func (h HandlerW) Processor() Processor {
	return h.handler.Processor()
}

func (h HandlerW) UriDict() UriDict {
	return h.handler.UriDict()
}
