package proxy

import (
	"axj/Kt/KtBuffer"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type PrxAdap struct {
	id         int32
	locker     sync.Locker
	closed     bool
	serv       *PrxServ
	outConn    *net.TCPConn // 外部代理连接
	outBuff    []byte
	outBuffer  *KtBuffer.Buffer
	outCtx     interface{}
	buffer     *KtBuffer.Buffer
	passTime   int64          // 超时时间
	inConn     *net.TCPConn   // 内部处理连接
	trafficReq *gw.TrafficReq // 流量监控
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
	that.passTime = time.Now().Unix() + Config.AdapTimeout
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
	trafficReq := that.trafficReq
	if trafficReq != nil {
		trafficDrt := that.serv.TrafficDrt
		trafficReq0 := &gw.TrafficReq{Cid: trafficReq.Cid, Gid: trafficReq.Gid, Sub: trafficReq.Sub}
		Util.GoSubmit(func() {
			for !that.closed {
				time.Sleep(trafficDrt)
				trafficReq0.Start = trafficReq.Start
				trafficReq0.In = trafficReq.In
				trafficReq0.Out = trafficReq.Out
				rep, err := AclClient.Traffic(Config.AclCtx(), trafficReq0)
				if rep != nil && rep.Val {
					trafficReq.Start = time.Now().Unix()
					trafficReq.In -= trafficReq0.In
					trafficReq.Out -= trafficReq0.Out
				}

				if err != nil {
					AZap.Logger.Warn("Acl TrafficReq Err "+Config.Acl, zap.Error(err))
				}
			}
		})
	}

	{
		var inBuffer *KtBuffer.Buffer
		inBuff := Util.GetBufferBytes(that.serv.Proto.ReadBufferSize(that.serv.Cfg), &inBuffer)
		Util.GoSubmit(func() {
			// 数据转发
			for {
				size, err := that.inConn.Read(inBuff)
				if err != nil {
					that.Close(err)
					break
				}

				if trafficReq != nil {
					trafficReq.In += int64(size)
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
				that.outConn.SetLinger(Config.CloseDelay)
				that.outConn.Close()
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

				if trafficReq != nil {
					trafficReq.Out += int64(size)
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
				that.inConn.SetLinger(Config.CloseDelay)
				that.inConn.Close()
			}
		})
	}
}
