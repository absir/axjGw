package ANet

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Thrd/Util"
	"sync"
	"time"
)

const (
	// 特殊请求
	REQ_PUSH   int32 = 0  // 推送
	REQ_LEAV   int32 = 1  // 软关闭
	REQ_LAST   int32 = 2  // 消息推送检查+
	REQ_KEY    int32 = 3  // 秘钥
	REQ_ACL    int32 = 4  // 请求开启
	REQ_BEAT   int32 = 5  // 心跳
	REQ_ROUTE  int32 = 6  // 路由字典
	REQ_LOOP   int32 = 15 // 连接接受
	REQ_ONEWAY int32 = 16 // 路由处理
)

type ConnM struct {
	ConnC
	id       int64
	initTime int64
	idleTime int64
}

func (that ConnM) PInit() {
	that.ConnC.PInit()
	that.id = 0
	//that.initTime = 0
	//that.idleTime = 0
}

//func (that ConnM) PRelease() bool {
//	if that.ConnC.PRelease() {
//		that.id = 0
//		return true
//	}
//
//	return false
//}

func (that ConnM) Id() int64 {
	return that.id
}

func (that ConnM) InitTime() int64 {
	return that.initTime
}

func (that ConnM) IdleTime() int64 {
	return that.idleTime
}

type HandlerM interface {
	Handler
	ConnM(conn Conn) ConnM
}

type ConnMng struct {
	HandlerW
	handlerM  HandlerM
	idWorker  *Util.IdWorker
	idleTime  int64
	checkTime time.Duration
	connPool  *Util.AllocPool
	loopTime  int64
	ConnMap   sync.Map
	beatBs    []byte
}

func NewConnMng(handler HandlerM, workerId int32, idleTime time.Duration, checkTime time.Duration, connPool bool) *ConnMng {
	c := new(ConnMng)
	c.handler = handler
	c.handlerM = handler
	var err error
	c.idWorker, err = Util.NewIdWorker(workerId)
	Kt.Panic(err)
	c.idleTime = int64(idleTime)
	c.checkTime = checkTime
	c.connPool = Util.NewAllocPool(connPool, func() Util.Pool {
		return c.New()
	})

	c.loopTime = 0
	c.ConnMap = sync.Map{}
	c.beatBs = c.Processor().Protocol.Rep(REQ_BEAT, "", 0, nil, false, 0)
	return c
}

func (that ConnMng) ConnM(conn Conn) ConnM {
	return that.handlerM.ConnM(conn)
}

func (that ConnMng) OpenConn(client Client) Conn {
	conn := that.connPool.Get().(Conn)
	that.Open(conn, client)
	return conn
}

func (that ConnMng) Open(conn Conn, client Client) {
	conn.Get().Open(client, that)
	connM := that.ConnM(conn)
	connM.id = that.idWorker.Generate()
	connM.initTime = time.Now().UnixNano()
	connM.idleTime = connM.initTime
	// connM.Get().poolG = APro.PoolOne
	that.handler.Open(conn, client)
}

func (that ConnMng) OnClose(conn Conn, err error, reason interface{}) {
	connM := that.ConnM(conn)
	if connM.id != 0 {
		that.ConnMap.Delete(connM.id)
		connM.id = 0
		that.connPool.Put(conn, true)
	}

	that.handler.OnClose(conn, err, reason)
}

func (that ConnMng) Last(conn Conn, req bool) {
	// 心跳延长
	connM := that.ConnM(conn)
	connM.idleTime = time.Now().UnixNano() + that.idleTime
	that.handler.Last(conn, req)
}

// 空闲检测
func (that ConnMng) IdleStop() {
	that.loopTime = -1
}

func (that ConnMng) IdleLoop() {
	loopTime := time.Now().UnixNano()
	that.loopTime = loopTime
	for loopTime == that.loopTime {
		time.Sleep(that.checkTime)
		time := time.Now().UnixNano()
		that.ConnMap.Range(func(key, value interface{}) bool {
			conn := value.(Conn)
			connM := that.ConnM(conn)
			connC := connM.ConnC
			// 已关闭链接
			if connC.Closed() {
				that.ConnMap.Delete(key)
				return true
			}

			if connM.idleTime <= time {
				// 直接心跳
				that.Last(conn, false)
				go connC.Rep(-1, "", 0, that.beatBs, false, false, nil)
			}

			return true
		})
	}
}

func (that ConnMng) RegConn(conn Conn, poolG int) {
	that.Last(conn, true)
	connM := that.ConnM(conn)
	that.ConnMap.Store(connM.id, conn)
	if poolG > 1 {
		pg := connM.ConnC.poolG
		if pg == nil || !pg.StrictAs(poolG) {
			connM.ConnC.poolG = APro.NewPoolLimit(poolG)
		}
	}
}

func (that ConnMng) UnRegConn(conn Conn, close bool) {
	connM := that.ConnM(conn)
	that.ConnMap.Delete(connM.id)
	if close {
		connM.ConnC.Close(ERR_CLOSED, nil)
	}
}
