package AGnet

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"github.com/panjf2000/gnet"
	"sync"
	"time"
)

func ctxConn(c gnet.Conn) *Conn {
	ctx := c.Context()
	if ctx == nil {
		return nil
	}

	conn, _ := ctx.(*Conn)
	return conn
}

type Conn struct {
	c      gnet.Conn
	locker sync.Locker
	cond   *sync.Cond
	async  *Util.NotifierAsync
	client ANet.Client
}

func NewConn(c gnet.Conn) *Conn {
	that := new(Conn)
	that.locker = new(sync.Mutex)
	that.cond = sync.NewCond(that.locker)
	that.async = Util.NewNotifierAsync(nil, that.locker)
	return that
}

func (that Conn) reqOne() {
	if that.c.BufferLength() <= 0 || that.client == nil {
		return
	}

	that.client.Get().ReqOne(that)
}

func (that Conn) aBufferLen() int {
	var aLen int
	c := that.c
	for {
		aLen = c.BufferLength()
		if aLen > 0 {
			return aLen
		}

		that.cond.Wait()
	}
}

func (that Conn) Read(p []byte) (n int, err error) {
	aLen := that.aBufferLen()
	bLen := len(p)
	if bLen > aLen {
		bLen = aLen
	}

	var bs []byte
	aLen, bs = that.c.ReadN(aLen)
	copy(p, bs)
	that.c.ShiftN(aLen)
	return aLen, nil
}

func (that Conn) ReadByte() (byte, error) {
	that.aBufferLen()
	_, bs := that.c.ReadN(1)
	that.c.ShiftN(1)
	return bs[0], nil
}

func (that Conn) ReadA() (error, []byte, ANet.Reader) {
	return nil, nil, that
}

func (that Conn) Sticky() bool {
	return true
}

func (that Conn) Out() *[]byte {
	return nil
}

func (that Conn) Write(bs []byte) error {
	return that.c.AsyncWrite(bs)
}

func (that Conn) IsWriteAsync() bool {
	return true
}

func (that Conn) Close() {
	that.c.Close()
}

func (that Conn) RemoteAddr() string {
	return that.c.RemoteAddr().String()
}

type Code struct {
}

func (that Code) Encode(c gnet.Conn, buf []byte) ([]byte, error) {
	return buf, nil
}

func (that Code) Decode(c gnet.Conn) ([]byte, error) {
	conn := ctxConn(c)
	if conn != nil {
		// 触发阻塞和请求线程
		conn.cond.Signal()
		conn.async.Start(conn.reqOne)
	}

	return nil, nil
}

type Handler struct {
}

func (that Handler) OnInitComplete(server gnet.Server) (action gnet.Action) {
	AZap.Logger.Info("gnet.Server OnInitComplete" + server.Addr.String())
	return gnet.None
}

func (that Handler) OnShutdown(server gnet.Server) {
	AZap.Logger.Info("gnet.Server OnShutdown" + server.Addr.String())
}

func (that Handler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (that Handler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	conn := ctxConn(c)
	if conn != nil {
		conn.Close()
	}

	return gnet.Close
}

func (that Handler) PreWrite() {
}

func (that Handler) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (that Handler) Tick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.Shutdown
}
