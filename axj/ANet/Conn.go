package ANet

import (
	"axj/APro"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"errors"
	"go.uber.org/zap"
	"sync"
)

// 错误
var ERR_CLOSED = errors.New("CLOSED")
var ERR_CRASH = errors.New("CRASH")

type Conn interface {
	Util.Pool
	Get() *ConnC
}

type ConnC struct {
	locker    sync.Locker
	encryKey  []byte
	repBuffer *[]byte
	closed    int8
	client    Client
	manager   Manager
	poolG     APro.PoolG
}

func (that ConnC) PInit() {
	that.locker = new(sync.Mutex)
	that.encryKey = nil
	that.repBuffer = nil
	that.closed = 0
	that.client = nil
	that.manager = nil
	that.poolG = nil
}

func (that ConnC) PRelease() bool {
	if that.client != nil {
		that.Close(nil, nil)
		return true
	}

	return false
}

func (that ConnC) Open(client Client, manager Manager) {
	that.closed = 0
	that.client = client
	that.manager = manager
	// that.poolG = nil
}

func (that *ConnC) Get() *ConnC {
	return that
}

func (that ConnC) Client() Client {
	return that.client
}

func (that ConnC) Locker() sync.Locker {
	return that.locker
}

func (that ConnC) Closed() bool {
	return that.closed != 0
}

func (that ConnC) Manager() Manager {
	return that.manager
}

func (that ConnC) SetEncryKey(encryKey []byte) {
	that.encryKey = encryKey
}

type Handler interface {
	New() Conn
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

func (that ConnC) Close(err error, reason interface{}) {
	that.close(err, reason, false)
}

const CONN_CLOSED int8 = 127

func (that *ConnC) close(err error, reason interface{}, inner bool) {
	if !inner {
		if that.Closed() {
			return
		}

		that.locker.Lock()
		defer that.locker.Unlock()
		if that.Closed() {
			return
		}
	}

	// 关闭中
	that.closed++
	// 关闭恢复
	defer that.recover()
	// 关闭日志
	if that.closed <= 3 {
		// logger before
		that.logClose(err, reason)
	}

	// that.poolG = nil
	manager := that.manager
	if manager != nil {
		that.manager = nil
		manager.OnClose(that, err, reason)
	}

	client := that.client
	if client != nil {
		that.client = nil
		client.Close()
	}

	if that.closed > 3 && that.closed <= 6 {
		// logger last
		that.logClose(err, reason)
	}

	// 已关闭
	that.closed = CONN_CLOSED
}

func (that ConnC) logClose(err error, reason interface{}) {
	AZap.Logger.Info("Conn close", zap.Error(err), zap.Reflect("reason", reason))
}

func (that ConnC) recover() error {
	if reason := recover(); reason != nil {
		err, ok := reason.(error)
		if ok {
			reason = nil
			that.close(err, nil, true)

		} else {
			err = ERR_CRASH
			that.close(err, reason, true)
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

func (that *ConnC) Req() (err error, req int32, uri string, uriI int32, data []byte) {
	if that.Closed() {
		return ERR_CLOSED, 0, "", 0, nil
	}

	that.manager.Last(that, true)
	processor := that.manager.Processor()
	err, req, uri, uriI, data = Req(that.client, processor.Protocol, processor.Compress, processor.Encrypt, that.encryKey, processor.DataMax)
	if uri == "" && uriI > 0 {
		uriDict := that.manager.UriDict()
		if uriDict != nil {
			uri = uriDict.UriIMapUri()[uriI]
		}
	}

	return
}

func (that *ConnC) ReqLoop() {
	for {
		err, req, uri, uriI, data := that.Req()
		if err != nil {
			that.Close(err, nil)
			break
		}

		if !that.manager.OnReq(that, req, uri, uriI, data) {
			poolG := that.poolG
			if poolG == APro.PoolOne {
				that.handlerReqIo(nil, req, uri, uriI, data)

			} else {
				go that.handlerReqIo(poolG, req, uri, uriI, data)
				if poolG != nil {
					poolG.Add()
				}
			}
		}
	}
}

func (that *ConnC) handlerReqIo(poolG APro.PoolG, req int32, uri string, uriI int32, data []byte) {
	if poolG == nil {
		defer poolG.Done()
	}

	that.manager.OnReqIO(that, req, uri, uriI, data)
}

// 单进程阻塞 req < 0 直接 WriteData
func (that *ConnC) Rep(req int32, uri string, uriI int32, data []byte, isolate bool, encry bool, batch *RepBatch) error {
	if that.Closed() {
		return ERR_CLOSED
	}

	that.manager.Last(that, false)
	uriDict := that.manager.UriDict()
	if uriI <= 0 {
		if uri != "" && uriDict != nil {
			uriI = uriDict.UriMapUriI()[uri]
			if uriI > 0 {
				uri = ""
			}
		}
	}

	encryKey := that.encryKey
	if !encry {
		encryKey = nil
	}

	processor := that.manager.Processor()
	var err error = nil
	if batch == nil {
		err = Rep(that.client, that.repBuffer, processor.Protocol, processor.Compress, processor.CompressMin, processor.Encrypt, encryKey, req, uri, uriI, data, isolate)

	} else {
		err = batch.Rep(that.client, that.repBuffer, processor.Encrypt, encryKey)
	}

	if err != nil {
		that.Close(err, nil)
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

func (that HandlerW) New() Conn {
	return that.handler.New()
}

func (that HandlerW) Open(conn Conn, client Client) {
	that.handler.Open(conn, client)
}

func (that HandlerW) OnClose(conn Conn, err error, reason interface{}) {
	that.handler.OnClose(conn, err, reason)
}

func (that HandlerW) Last(conn Conn, req bool) {
	that.handler.Last(conn, req)
}

func (that HandlerW) OnReq(conn Conn, req int32, uri string, uriI int32, data []byte) bool {
	return that.handler.OnReq(conn, req, uri, uriI, data)
}

func (that HandlerW) OnReqIO(conn Conn, req int32, uri string, uriI int32, data []byte) {
	that.handler.OnReqIO(conn, req, uri, uriI, data)
}

func (that HandlerW) Processor() Processor {
	return that.handler.Processor()
}

func (that HandlerW) UriDict() UriDict {
	return that.handler.UriDict()
}
