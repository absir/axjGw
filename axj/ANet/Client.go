package ANet

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"errors"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

// 错误
var ERR_CLOSED = errors.New("CLOSED")
var ERR_CRASH = errors.New("CRASH")
var ERR_NOWAY = errors.New("NOWAY")

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

const (
	FLG_ENCRYPT  int32 = 1      // 加密
	FLG_COMPRESS int32 = 1 << 2 // 压缩
)

type Client interface {
	Get() *ClientCnn
}

type ClientCnn struct {
	client   Client
	conn     Conn
	handler  Handler
	encryKey []byte
	compress bool
	locker   sync.Locker
	closed   int8
	limiter  Util.Limiter
}

func (that *ClientCnn) Client() Client {
	return that.client
}

func (that *ClientCnn) Locker() sync.Locker {
	return that.locker
}

func (that *ClientCnn) IsClosed() bool {
	return that.closed != 0
}

func (that *ClientCnn) Open(client Client, conn Conn, handler Handler, encryKey []byte, compress bool) {
	if client == nil {
		client = that
	}

	that.client = client
	that.conn = conn
	that.handler = handler
	that.encryKey = encryKey
	that.compress = compress
	that.locker = new(sync.Mutex)
}

func (that *ClientCnn) SetLimiter(limit int) {
	if limit == 0 {
		that.limiter = Util.LimiterOne

	} else if limit > 0 {
		that.limiter = Util.NewLimiterLocker(limit, nil)

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
		err = nil
		reason = nil
	}

	// 关闭连接
	conn := that.conn
	if conn != nil {
		that.conn = nil
		conn.Close()
	}

	// 解除reqLoop阻塞
	limiter := that.limiter
	if limiter != nil {
		that.limiter = nil
		if limiter != Util.LimiterOne && !limiter.StrictAs(1) {
			limiter.Done()
		}
	}

	// 关闭处理
	handler := that.handler
	if handler != nil {
		that.handler = nil
		handler.OnClose(that.client, err, reason)
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

	if err == nil {
		AZap.Logger.Debug("Conn close", zap.Reflect("reason", reason))

	} else if err == io.EOF {
		AZap.Logger.Debug("Conn close EOF", zap.Reflect("reason", reason))

	} else {
		if nErr, ok := err.(*net.OpError); ok {
			AZap.Logger.Debug("Conn close", zap.String("ERR", nErr.Error()), zap.Reflect("reason", reason))

		} else {
			AZap.Logger.Warn("Conn close ERR", zap.Error(err), zap.Reflect("reason", reason))
		}
	}
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
	handler.OnKeep(that.client, false)

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
	return that.RepCData(out, req, uri, uriI, data, 0, isolate, encry, id)
}

func (that *ClientCnn) RepCData(out bool, req int32, uri string, uriI int32, data []byte, cData int32, isolate bool, encry bool, id int64) error {
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
	handler.OnKeep(that.client, false)

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

	var err error
	if cData > 0 {
		if cData == 1 {
			// 推送已压缩数据
			if !that.compress {
				// 客户端不支持压缩数据
				return ERR_NOWAY
			}

			err = handler.Processor().RepCData(nil, out, that.conn, encryKey, req, uri, uriI, data, isolate, id)

		} else {
			// 无法压缩
			err = handler.Processor().Rep(nil, out, that.conn, encryKey, false, req, uri, uriI, data, isolate, id)
		}

	} else {
		// 自决压缩
		err = handler.Processor().Rep(nil, out, that.conn, encryKey, that.compress, req, uri, uriI, data, isolate, id)
	}

	if err != nil {
		that.Close(err, nil)
	}

	return err
}

func CloseDelay(conn Conn, drt time.Duration) {
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
		if drt <= 0 {
			drt = that.handler.KickDrt()
		}

		go CloseDelay(conn, drt)
		that.conn = nil
		that.handler.Processor().Rep(nil, true, conn, that.encryKey, that.compress, REQ_KICK, "", 0, data, isolate, 0)
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
			limiter := that.limiter
			if limiter == nil || limiter == Util.LimiterOne || (limiter != nil && limiter.StrictAs(1)) {
				that.poolReqIO(nil, req, uri, uriI, data)

			} else {
				go that.poolReqIO(limiter, req, uri, uriI, data)
				if limiter != nil {
					limiter.Add()
				}
			}
		}
	}
}

func (that *ClientCnn) poolReqIO(limiter Util.Limiter, req int32, uri string, uriI int32, data []byte) {
	if limiter != nil {
		defer limiter.Done()
	}

	that.handler.OnReqIO(that.client, req, uri, uriI, data)
}
