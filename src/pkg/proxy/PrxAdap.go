package proxy

import (
	"axj/Thrd/Util"
	"net"
	"sync"
)

type PrxAdap struct {
	locker   sync.Locker
	closed   bool
	proto    PrxProto
	outConn  *net.TCPConn // 外部代理连接
	outBuff  []byte
	passTime int64        // 超时时间
	inConn   *net.TCPConn // 内部处理连接
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
	PrxMng.closeConn(that.outConn, nil)
	PrxMng.closeConn(that.inConn, err)
}

func (that *PrxAdap) doInConn(inConn *net.TCPConn) {
	// 接入连接
	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		PrxMng.closeConn(inConn, nil)
		return
	}

	that.inConn = inConn
	that.locker.Unlock()
	{
		inBuff := make([]byte, that.proto.ReadBufferSize())
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

				_, err = that.inConn.Write(that.outBuff[:size])
				if err != nil {
					that.Close(err)
					return
				}
			}
		})
	}
}
