package proxy

import (
	"axj/ANet"
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
	outBuffer *bytes.Buffer
	outCtx    interface{}
	buffer    *bytes.Buffer
	passTime  int64        // 超时时间
	inConn    *net.TCPConn // 内部处理连接
}

func (that *PrxAdap) Close(err error) {
	if that.closed {
		return
	}

	inConnNil := false
	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		return
	}

	that.closed = true
	inConnNil = that.inConn == nil
	that.locker.Unlock()
	if inConnNil {
		Util.PutBuffer(that.outBuffer)
		Util.PutBuffer(that.buffer)
	}

	PrxMng.connMap.Delete(that.id)
	PrxMng.closeConn(that.outConn, inConnNil, nil)
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
		var inBuffer *bytes.Buffer
		inBuff := Util.GetBufferBytes(that.serv.Proto.ReadBufferSize(that.serv.Cfg), &inBuffer)
		Util.GoSubmit(func() {
			// 数据转发
			for {
				size, err := that.inConn.Read(inBuff)
				if err != nil {
					that.Close(err)
					break
				}

				that.OnKeep()
				_, err = that.outConn.Write(inBuff[:size])
				if err != nil {
					that.Close(err)
					break
				}
			}

			Util.PutBuffer(inBuffer)
			if Config.CloseDelay > 0 {
				ANet.CloseDelayTcp(that.outConn, Config.CloseDelay)
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
					break
				}

				that.OnKeep()
				data := that.outBuff[:size]
				if that.outCtx != nil {
					data, err = that.serv.Proto.ProcServerData(that.serv.Cfg, that.outCtx, that.buffer, data, that.outConn)
				}

				if err != nil {
					that.Close(err)
					break
				}

				if data == nil {
					break
				}

				_, err = that.inConn.Write(data)
				if err != nil {
					that.Close(err)
					break
				}
			}

			Util.PutBuffer(that.outBuffer)
			Util.PutBuffer(that.buffer)
			if Config.CloseDelay > 0 {
				ANet.CloseDelayTcp(that.inConn, Config.CloseDelay)
			}
		})
	}
}
