package proxy

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtBytes"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/pkg/agent"
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
	gidMap    *cmap.CMap
	dialId    int32
	dialsMap  *cmap.CMap
}

var PrxMng *prxMng

var ERR_TIMEOUT = Kt.NewErrReason("TIMEOUT")

func initPrxMng() {
	that := new(prxMng)
	that.locker = new(sync.Mutex)
	that.connMap = cmap.NewCMapInit()
	that.gidMap = cmap.NewCMapInit()
	that.dialsMap = cmap.NewCMapInit()
	PrxMng = that
}

func (that *prxMng) Dial(cid int64, gid string, addr string, timeout time.Duration) bool {
	client := that.Client(cid, gid)
	if client == nil {
		return false
	}

	var dials chan bool = nil
	var num int32 = 0
	that.locker.Lock()
	id := that.dialId
	for {
		if id >= Config.AdapMaxId {
			id = 0

		} else {
			id++
		}

		if _, ok := that.connMap.Load(id); !ok {
			// 保证实时性
			dials = make(chan bool, 1)
			that.dialsMap.Store(id, dials)
			that.dialId = id
			break
		}

		num++
		if num >= Config.AdapMaxId {
			num = 0
			time.Sleep(time.Second)
		}
	}

	that.locker.Unlock()
	if dials == nil {
		return false
	}

	bs := make([]byte, 8)
	KtBytes.SetInt64(bs, 0, int64(timeout), nil)
	err := client.Get().Rep(false, agent.REQ_DIAL, addr, id, bs, false, false, 0)
	if err != nil {
		AZap.Logger.Warn("REQ_DIAL err", zap.Error(err))
		return false
	}

	if timeout <= 0 {
		timeout = Config.DialTimeout
	}

	select {
	case dial := <-dials:
		that.dialsMap.Delete(id)
		return dial
	case <-time.After(timeout):
		that.dialsMap.Delete(id)
		return false
	}

	close(dials)
	return false
}

func (that *prxMng) DialRep(id int32, ok bool) {
	that.locker.Lock()
	val, _ := that.dialsMap.LoadAndDelete(id)
	that.locker.Unlock()
	if val == nil {
		return
	}

	defer recover()
	dials, _ := val.(chan bool)
	if dials != nil {
		dials <- ok
	}
}

func (that *prxMng) closeConn(conn *net.TCPConn, immed bool, err error) {
	if conn == nil {
		return
	}

	if immed {
		conn.SetLinger(0)
	}

	conn.Close()
	if err != nil && err != io.EOF {
		if _, ok := err.(*Kt.ErrReason); ok {
			AZap.Debug("PrxAdap Close %v", err.Error())

		} else if _, ok = err.(*net.OpError); ok {
			AZap.Debug("PrxAdap Close %v", err)

		} else {
			AZap.Logger.Warn("PrxAdap Close", zap.Error(err))
		}
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

func (that *prxMng) adapOpen(serv *PrxServ, outConn *net.TCPConn, outBuff []byte, outBuffer *KtBuffer.Buffer, outCtx interface{}, buffer *KtBuffer.Buffer) (int32, *PrxAdap) {
	adap := new(PrxAdap)
	adap.locker = new(sync.Mutex)
	adap.serv = serv
	adap.outConn = outConn
	adap.outBuff = outBuff
	adap.outBuffer = outBuffer
	adap.outCtx = adap.serv.Proto.ProcServerCtx(adap.serv.Cfg, outCtx, buffer, outConn)
	if adap.outCtx == nil {
		Util.PutBuffer(buffer)

	} else {
		adap.buffer = buffer
	}

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
			adap.id = id
			adap.OnKeep()
			that.connMap.Store(id, adap)
			that.connId = id
			break
		}

		num++
		if num >= Config.AdapMaxId {
			num = 0
			time.Sleep(time.Second)
		}
	}

	that.locker.Unlock()
	return id, adap
}

func (that *prxMng) adapConn(id int32, inConn *net.TCPConn) {
	val, ok := that.connMap.Load(id)
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

func (that *prxMng) adapClose(id int32) {
	val, _ := that.connMap.LoadAndDelete(id)
	adap, _ := val.(*PrxAdap)
	if adap != nil {
		adap.Close(nil)
	}
}

func (that *prxMng) Client(cid int64, gid string) ANet.Client {
	for i := 0; i < 2; i++ {
		if cid > 0 {
			client := PrxServMng.Manager.Client(cid)
			if client != nil {
				return client
			}
		}

		if i == 0 && gid != "" {
			val, _ := that.gidMap.Load(gid)
			if val != nil {
				cid, _ = val.(int64)
				continue
			}
		}

		break
	}

	return nil
}
