package AGnet

import (
	"axj/Thrd/AZap"
	"github.com/panjf2000/gnet"
	"time"
)

type AHandler struct {
	out       bool
	openedFun func(aConn *AConn)
}

func NewAHandler(out bool, fun func(aConn *AConn)) *AHandler {
	that := new(AHandler)
	that.out = out
	that.openedFun = fun
	return that
}

func (that AHandler) OnInitComplete(server gnet.Server) (action gnet.Action) {
	AZap.Logger.Info("gnet.Server OnInitComplete" + server.Addr.String())
	return gnet.None
}

func (that AHandler) OnShutdown(server gnet.Server) {
	AZap.Logger.Info("gnet.Server OnShutdown" + server.Addr.String())
}

func (that AHandler) OnOpened(c gnet.Conn) (out []byte, action gnet.Action) {
	aConn := open(c, that.out)
	c.SetContext(aConn)
	if that.openedFun != nil {
		that.openedFun(aConn)
	}

	return nil, gnet.None
}

func (that AHandler) OnClosed(c gnet.Conn, err error) (action gnet.Action) {
	conn := connCtx(c)
	if conn != nil {
		conn.Close()
	}

	return gnet.Close
}

func (that AHandler) PreWrite() {
}

func (that AHandler) React(frame []byte, c gnet.Conn) (out []byte, action gnet.Action) {
	return nil, gnet.None
}

func (that AHandler) Tick() (delay time.Duration, action gnet.Action) {
	return 0, gnet.Shutdown
}
