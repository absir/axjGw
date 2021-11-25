package AgNet

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"container/list"
	"errors"
	"github.com/panjf2000/gnet"
	"github.com/panjf2000/gnet/pool/ringbuffer"
	prb "github.com/panjf2000/gnet/pool/ringbuffer"
	"io"
	"sync"
	"time"
)

var BUFF_MAX = 1024 * 1024 * 32
var ERR_BUFF_MAX = errors.New("ERR_BUFF_MAX")

func ConnCtx(c gnet.Conn) *AgNetConn {
	ctx := c.Context()
	if ctx == nil {
		return nil
	}

	conn, _ := ctx.(*AgNetConn)
	return conn
}

type AgNetConn struct {
	c      gnet.Conn
	closed bool
	locker sync.Locker
	cond   *sync.Cond
	buffs  *list.List
	buffer *ringbuffer.RingBuffer
	rBuff  []byte
	async  *Util.NotifierAsync
	client ANet.Client
}

func open(c gnet.Conn) *AgNetConn {
	that := new(AgNetConn)
	that.c = c
	that.locker = new(sync.Mutex)
	that.cond = sync.NewCond(that.locker)
	that.buffer = prb.Get()
	that.buffs = new(list.List)
	that.async = Util.NewNotifierAsync(nil, that.locker, that.reqCond)
	return that
}

func (that *AgNetConn) Close() {
	if that.closed {
		return
	}

	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		return
	}

	// 关闭状态
	that.closed = true
	// 通道关闭
	that.cond.Signal()
	that.locker.Unlock()
	// 连接关闭
	that.c.Close()
	// 回收that.buffer
	prb.Put(that.buffer)
}

func (that *AgNetConn) ReqStart(client ANet.Client) {
	that.client = client
	that.async.Start(that.reqOne)
}

func (that *AgNetConn) reqCond() bool {
	return that.buffer.Length() > 0 || that.buffs.Front() != nil
}

func (that *AgNetConn) reqOne() {
	if that.client == nil || !that.reqCond() {
		return
	}

	that.client.Get().ReqOne()
}

func (that *AgNetConn) rBufferLen() (int, error) {
	var aLen int
	for {
		aLen = that.buffer.Length()
		if aLen > 0 {
			// 最大缓冲区过大
			if aLen > BUFF_MAX {
				return aLen, ERR_BUFF_MAX
			}

			return aLen, nil
		}

		if that.closed {
			return 0, io.EOF
		}

		// 锁写入bs
		that.locker.Lock()
		if that.closed {
			return 0, io.EOF
		}

		el := that.buffs.Front()
		if el == nil {
			// buffs为空，阻塞
			that.cond.Wait()

		} else {
			nxt := el
			for nxt != nil {
				el = nxt
				nxt = el.Next()
				that.buffs.Remove(el)
				that.buffer.Write(el.Value.([]byte))
			}
		}

		that.locker.Unlock()
	}
}

func (that *AgNetConn) Read(p []byte) (int, error) {
	_, err := that.rBufferLen()
	if err != nil {
		return 0, err
	}

	return that.buffer.Read(p)
}

func (that *AgNetConn) ReadByte() (byte, error) {
	_, err := that.rBufferLen()
	if err != nil {
		return 0, err
	}

	return that.buffer.ReadByte()
}

func (that *AgNetConn) ReadA() (error, []byte, ANet.Reader) {
	return nil, nil, that
}

func (that *AgNetConn) Sticky() bool {
	return true
}

func (that *AgNetConn) Out() *[]byte {
	return nil
}

func (that *AgNetConn) Write(bs []byte) error {
	if that.closed {
		return io.EOF
	}

	return that.c.AsyncWrite(bs)
}

func (that *AgNetConn) IsWriteAsync() bool {
	return true
}

func (that *AgNetConn) RemoteAddr() string {
	return that.c.RemoteAddr().String()
}

type AgCode struct {
}

func (that AgCode) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (that AgCode) Decode(c gnet.Conn) ([]byte, error) {
	conn := ConnCtx(c)
	if conn != nil {
		conn.locker.Lock()
		if conn.closed {
			conn.locker.Unlock()
			return nil, io.EOF
		}

		conn.buffs.PushBack(c.Read())
		conn.cond.Signal()
		// 触发请求
		conn.async.StartLock(nil, false)
		conn.locker.Unlock()
		c.ResetBuffer()
	}

	return nil, nil
}

type AgHandler struct {
}

func (that AgHandler) OnInitComplete(server gnet.Server) (action gnet.Action) {
	AZap.Logger.Info("gnet.Server OnInitComplete" + server.Addr.String())
	return gnet.None
}

func (that AgHandler) OnShutdown(server gnet.Server) {
	AZap.Logger.Info("gnet.Server OnShutdown" + server.Addr.String())
}

func (that AgHandler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	c.SetContext(open(c))
	return nil, gnet.None
}

func (that AgHandler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	conn := ConnCtx(c)
	if conn != nil {
		conn.Close()
	}

	return gnet.Close
}

func (that AgHandler) PreWrite() {
}

func (that AgHandler) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (that AgHandler) Tick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.Shutdown
}
