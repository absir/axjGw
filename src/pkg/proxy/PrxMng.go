package proxy

import (
	"axj/Thrd/AZap"
	"axj/Thrd/cmap"
	"errors"
	"go.uber.org/zap"
	"io"
	"net"
	"sync"
	"time"
)

type prxMng struct {
	locker    sync.Locker
	connMap   *cmap.CMap
	connId    int32
	loopTime  int64
	checkTime int64
	checkBuff []interface{}
}

var PrxMng *prxMng

var ERR_TIMEOUT = errors.New("TIMEOUT")

func initPrxMng() {
	that := new(prxMng)
	that.locker = new(sync.Mutex)
	that.connMap = cmap.NewCMapInit()
	PrxMng = that
}

func (that *prxMng) closeConn(conn *net.TCPConn, err error) {
	if conn == nil {
		return
	}

	conn.SetLinger(0)
	conn.Close()
	if err != nil && err != io.EOF {
		AZap.Logger.Warn("PrxAdap close", zap.Error(err))
	}
}

func (that *prxMng) CheckStop() {
	that.loopTime = 0
}

func (that *prxMng) CheckLoop() {
	loopTime := time.Now().UnixNano()
	that.loopTime = loopTime
	for loopTime == that.loopTime {
		time.Sleep(Config.AdapCheckDrt)
		that.checkTime = time.Now().UnixNano()
		that.connMap.RangeBuff(that.checkRange, &that.checkBuff, Config.AdapCheckBuff)
	}
}

func (that *prxMng) checkRange(key, val interface{}) bool {
	adap, _ := val.(*PrxAdap)
	if adap == nil {
		that.connMap.Delete(key)
		return true
	}

	if adap.passTime < that.checkTime {
		that.connMap.Delete(key)
		adap.Close(ERR_TIMEOUT)
	}

	return true
}

func (that *prxMng) adapOpen(proto PrxProto, outConn *net.TCPConn, outBuff []byte) (int32, *PrxAdap) {
	adap := new(PrxAdap)
	adap.locker = new(sync.Mutex)
	adap.proto = proto
	adap.outConn = outConn
	adap.outBuff = outBuff
	var num int32 = 0
	that.locker.Lock()
	id := that.connId
	for {
		if id >= Config.AdapMaxId {
			id = 0

		} else {
			id++
		}

		if _, ok := that.connMap.Load(id); !ok {
			// 保证实时性
			adap.passTime = time.Now().UnixNano() + Config.AdapTimeout
			that.connMap.Store(id, adap)
			that.connId = id
			break
		}

		num++
		if num >= Config.AdapMaxId {
			num = 0
			time.Sleep(time.Millisecond)
		}
	}

	that.locker.Unlock()
	return id, adap
}

func (that *prxMng) adapConn(id int32, inConn *net.TCPConn) {
	val, ok := that.connMap.LoadAndDelete(id)
	adap, _ := val.(*PrxAdap)
	if adap == nil {
		// adap不存在
		if ok {
			that.connMap.Delete(id)
		}

		// 关闭连接
		inConn.SetLinger(0)
		inConn.Close()
		return
	}

	adap.doInConn(inConn)
}
