package ANet

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"errors"
	"go.uber.org/zap"
	"sync"
	"time"
)

// 错误
var ERR_CLOSED = errors.New("CLOSED")
var ERR_CRASH = errors.New("CRASH")

// 最终关闭状态
const CONN_CLOSED int8 = 127

// 请求定义
const (
	// 特殊请求
	REQ_PUSH   int32 = 0  // 推送
	REQ_PUSHI  int32 = 1  // 推送+id
	REQ_KICK   int32 = 2  // 软关闭
	REQ_LAST   int32 = 3  // 消息推送检查+
	REQ_LASTC  int32 = 4  // 消息推送检查+连续
	REQ_KEY    int32 = 5  // 秘钥
	REQ_ACL    int32 = 6  // 请求开启
	REQ_BEAT   int32 = 7  // 心跳
	REQ_ROUTE  int32 = 8  // 路由字典
	REQ_LOOP   int32 = 15 // 连接接受
	REQ_ONEWAY int32 = 16 // 路由处理
)

type Client interface {
	Get() *ClientCnn
}

type ClientCnn struct {
	conn     Conn
	handler  Handler
	encryKey []byte
	locker   sync.Locker
	closed   int8
	limiter  Util.Limiter
}

func (that *ClientCnn) Locker() sync.Locker {
	return that.locker
}

func (that *ClientCnn) IsClosed() bool {
	return that.closed != 0
}

func (that *ClientCnn) Open(conn Conn, handler Handler, encryKey []byte) {
	that.conn = conn
	that.handler = handler
	that.encryKey = encryKey
	that.locker = new(sync.Mutex)
}

func (that *ClientCnn) SetLimiter(limit int) {
	if limit == 0 {
		that.limiter = Util.LimiterOne

	} else if limit > 0 {
		that.limiter = Util.NewLimiterLocker(limit)

	} else {
		that.limiter = nil
	}
}

func (that *ClientCnn) Close(err error, reason interface{}) {
	that.close(err, reason, false)
}

func (that *ClientCnn) close(err error, reason interface{}, inner bool) {
	if !inner {
		if that.IsClosed() {
			return
		}

		that.locker.Lock()
		defer that.locker.Unlock()
		if that.IsClosed() {
			return
		}
	}

	// 关闭中
	that.closed++
	// 关闭恢复
	defer that.closeRcvr()
	// 关闭日志
	if that.closed <= 3 {
		// logger before
		that.closeLog(err, reason)
	}

	// 关闭连接
	conn := that.conn
	if conn != nil {
		that.conn = nil
		conn.Close()
	}

	// 关闭处理
	handler := that.handler
	if handler != nil {
		that.handler = nil
		handler.OnClose(that, err, reason)
	}

	if err != nil || reason != nil {
		// logger last
		that.closeLog(err, reason)
	}

	// 已关闭
	that.closed = CONN_CLOSED
}

func (that *ClientCnn) closeLog(err error, reason interface{}) {
	if err == nil && reason == nil {
		return
	}
	AZap.Logger.Info("Conn close", zap.Error(err), zap.Reflect("reason", reason))
}

func (that *ClientCnn) closeRcvr() error {
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

func (that *ClientCnn) Get() *ClientCnn {
	return that
}

func (that *ClientCnn) Req() (error, int32, string, int32, []byte) {
	if that.IsClosed() {
		return ERR_CLOSED, 0, "", 0, nil
	}

	handler := that.handler
	// 保持连接
	handler.OnKeep(that, true)
	// 获取请求
	err, req, uri, uriI, data := handler.Processor().Req(that.conn, that.encryKey)
	if uri == "" && uriI > 0 {
		// 路由解压
		uriDict := handler.UriDict()
		if uriDict != nil {
			uri = uriDict.UriIMapUri()[uriI]
		}
	}

	return err, req, uri, uriI, data
}

func (that *ClientCnn) Rep(out bool, req int32, uri string, uriI int32, data []byte, isolate bool, encry bool, id int64) error {
	handler := that.handler
	uriDict := handler.UriDict()
	if uriI <= 0 {
		// 路由压缩
		if uri != "" && uriDict != nil {
			uriI = uriDict.UriMapUriI()[uri]
			if uriI > 0 {
				uri = ""
			}
		}
	}

	// 保持连接
	handler.OnKeep(that, false)

	// 加密key
	var encryKey []byte = nil
	if encry {
		encryKey = that.encryKey
	}

	// 写入锁
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.IsClosed() {
		return ERR_CLOSED
	}

	err := handler.Processor().Rep(nil, out, that.conn, encryKey, req, uri, uriI, data, isolate, id)
	if err != nil {
		that.Close(err, nil)
	}

	return err
}

func closeDelay(conn Conn, drt time.Duration) {
	if drt < 1 {
		drt = 1
	}

	time.Sleep(drt)
	conn.Close()
}

func (that *ClientCnn) Kick(data []byte, isolate bool, drt time.Duration) {
	if that.IsClosed() {
		return
	}

	conn := that.conn
	if conn != nil {
		go closeDelay(conn, drt)
		that.conn = nil
		that.handler.Processor().Rep(nil, true, conn, that.encryKey, REQ_KICK, "", 0, data, isolate, 0)
	}

	that.Close(nil, nil)
}

func (that *ClientCnn) ReqLoop() {
	conn := that.conn
	for conn == that.conn {
		err, req, uri, uriI, data := that.Req()
		if err != nil {
			that.Close(err, nil)
			break
		}

		if !that.handler.OnReq(that, req, uri, uriI, data) {
			poolG := that.limiter
			if poolG == Util.LimiterOne || (poolG != nil && poolG.StrictAs(1)) {
				that.poolReqIO(nil, req, uri, uriI, data)

			} else {
				go that.poolReqIO(poolG, req, uri, uriI, data)
				if poolG != nil {
					poolG.Add()
				}
			}
		}
	}
}

func (that *ClientCnn) poolReqIO(poolG Util.Limiter, req int32, uri string, uriI int32, data []byte) {
	if poolG == nil {
		defer poolG.Done()
	}

	that.handler.OnReqIO(that, req, uri, uriI, data)
}
