package AGnet

import (
	"axj/ANet"
	"axj/Thrd/Util"
	"github.com/panjf2000/gnet"
	"io"
	"sync"
)

func connCtx(c gnet.Conn) *AConn {
	ctx := c.Context()
	if ctx == nil {
		return nil
	}

	conn, _ := ctx.(*AConn)
	return conn
}

type AConn struct {
	c           gnet.Conn
	out         bool
	wBuff       []byte
	closed      bool
	locker      sync.Locker
	listAsync   *Util.ListAsync
	frameReader *ANet.FrameReader
	client      ANet.Client
}

func open(c gnet.Conn, out bool) *AConn {
	that := new(AConn)
	that.c = c
	that.out = out
	that.locker = new(sync.Mutex)
	that.listAsync = Util.NewListAsync(nil, that.locker)
	that.frameReader = &ANet.FrameReader{}
	return that
}

func (that *AConn) StartFun(fun func(interface{})) {
	that.listAsync.SetRun(fun)
	that.listAsync.Start()
}

func (that *AConn) StartReq(client ANet.Client, frame *ANet.ReqFrame) {
	that.client = client
	that.listAsync.SetRun(that.OnReq)
	if frame != nil {
		that.OnReq(frame)
	}
}

func (that *AConn) OnReq(el interface{}) {
	// 请求客户端
	client := that.client
	if client == nil {
		that.Close()
		return
	}

	frame, _ := el.(*ANet.ReqFrame)
	if frame == nil {
		// EOF close
		client.Get().Close(io.EOF, nil)
		return
	}

	// 请求处理
	client.Get().OnReq(client.Get().ReqFrame(frame))
}

func (that *AConn) Close() {
	client := that.client
	if client != nil {
		// client先关闭
		that.client = nil
		client.Get().Close(nil, nil)
	}

	if that.closed {
		return
	}

	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		return
	}

	that.listAsync.Close(false)
	that.locker.Unlock()
	// 关闭连接
	that.c.Close()
}

func (that *AConn) ReadA() (error, []byte, ANet.Reader) {
	return nil, nil, nil
}

func (that *AConn) Sticky() bool {
	return true
}

func (that *AConn) Out() *[]byte {
	if that.out {
		return &that.wBuff
	}

	return nil
}

func (that *AConn) Write(bs []byte) error {
	if that.closed {
		return io.EOF
	}

	return that.c.AsyncWrite(bs)
}

func (that *AConn) IsWriteAsync() bool {
	return true
}

func (that *AConn) RemoteAddr() string {
	return that.c.RemoteAddr().String()
}
