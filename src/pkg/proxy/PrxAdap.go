package proxy

import (
	"axj/Thrd/Util"
	"bytes"
	"net"
	"sync"
	"time"
)

type PrxAdap struct {
	id        int32
	locker    sync.Locker
	closed    bool
	serv      *PrxServ
	outConn   *net.TCPConn // 外部代理连接
	outBuff   []byte
	outCtx    interface{}
	outBuffer *bytes.Buffer
	passTime  int64        // 超时时间
	inConn    *net.TCPConn // 内部处理连接
}

func (that *PrxAdap) Close(err error) {
	if that.closed {
		return
	}

	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		return
	}

	that.closed = true
	that.locker.Unlock()
	PrxMng.connMap.Delete(that.id)
	PrxMng.closeConn(that.outConn, false, nil)
	PrxMng.closeConn(that.inConn, false, err)
}

func (that *PrxAdap) OnKeep() {
	that.passTime = time.Now().UnixNano() + Config.AdapTimeout
}

func (that *PrxAdap) doInConn(inConn *net.TCPConn) {
	// 接入连接
	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		PrxMng.closeConn(inConn, true, nil)
		return
	}

	that.inConn = inConn
	that.locker.Unlock()
	{
		inBuff := make([]byte, that.serv.Proto.ReadBufferSize(that.serv.Cfg))
		Util.GoSubmit(func() {
			// 数据转发
			for {
				size, err := that.inConn.Read(inBuff)
				if err != nil {
					that.Close(err)
					return
				}

				_, err = that.outConn.Write(inBuff[:size])
				if err != nil {
					that.Close(err)
					return
				}
			}
		})
	}

	{
		Util.GoSubmit(func() {
			// 数据转发
			for {
				size, err := that.outConn.Read(that.outBuff)
				if err != nil {
					that.Close(err)
					return
				}

				data := that.outBuff[:size]
				if that.outCtx != nil {
					data, err = that.serv.Proto.ProcServerData(that.serv.Cfg, that.outCtx, that.outBuffer, data, that.outConn)
				}

				if err != nil {
					that.Close(err)
					return
				}

				if data == nil {
					continue
				}

				_, err = that.inConn.Write(data)
				if err != nil {
					that.Close(err)
					return
				}
			}
		})
	}
}
