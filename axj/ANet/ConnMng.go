package ANet

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Thrd/SnowFlake"
	"sync"
	"time"
)

const (
	// 特殊请求
	REQ_PUSH  int32 = 0 // 推送
	REQ_BEAT  int32 = 1 // 心跳
	REQ_ROUTE int32 = 2 // 路由字典
	REQ_URI   int32 = 3 // 路由交换
)

type Mng struct {
	id       int64
	initTime int64
	idleTime int64
	poolG    APro.PoolG
	userData interface{}
}

type ConnMng struct {
	HandlerW
	idWorker  *SnowFlake.IdWorker
	idleTime  int64
	checkTime time.Duration
	ConnMap   sync.Map
	beatBs    []byte
}

func (c *ConnMng) Data(conn *Conn) interface{} {
	conn.poolG = APro.PoolOne
	mng := new(Mng)
	mng.id = c.idWorker.Generate()
	mng.initTime = time.Now().UnixNano()
	mng.idleTime = mng.initTime
	mng.poolG = nil
	mng.userData = c.handler.Data(conn)
	return mng
}

func (c *ConnMng) Last(conn *Conn, req bool) {
	// 心跳延长
	conn.handlerData.(*Mng).idleTime = time.Now().UnixNano() + c.idleTime
}

func (c *ConnMng) OnClose(conn *Conn, err error, reason interface{}) {
	mng := conn.handlerData.(*Mng)
	c.ConnMap.Delete(mng.id)
	c.handler.OnClose(conn, err, reason)
}

func NewConnMng(handler Handler, workerId int32, idleTime time.Duration, checkTime time.Duration) *ConnMng {
	c := new(ConnMng)
	c.handler = handler
	var err error
	c.idWorker, err = SnowFlake.NewIdWorker(workerId)
	Kt.Panic(err)
	c.idleTime = int64(idleTime)
	c.checkTime = checkTime
	c.ConnMap = sync.Map{}
	c.beatBs = handler.Processor().Protocol.Rep(REQ_BEAT, "", 0, nil, false, 0)
	return c
}

func (c *ConnMng) InitConn(conn *Conn, client Client) {
	InitConn(conn, client, c)
}

// 空闲检测
func (c *ConnMng) IdleCheck() {
	time.Sleep(c.checkTime)
	time := time.Now().UnixNano()
	c.ConnMap.Range(func(key, value interface{}) bool {
		conn := value.(*Conn)
		// 已关闭链接
		if conn.closed {
			c.ConnMap.Delete(key)
			return true
		}

		if conn.handlerData.(*Mng).idleTime <= time {
			// 直接心跳
			c.Last(conn, false)
			go conn.Rep(-1, "", c.beatBs, false, false, nil)
		}

		return true
	})
}

func (c *ConnMng) RegConn(conn *Conn, poolG int) {
	c.Last(conn, true)
	c.ConnMap.Store(conn.handlerData.(*Mng).id, conn)
	if poolG > 1 {
		conn.poolG = APro.NewPoolLimit(poolG)
	}
}

func (c *ConnMng) UnRegConn(conn *Conn, close bool) {
	c.ConnMap.Delete(conn.handlerData.(*Mng).id)
	if close {
		conn.Close(ERR_CLOSED, nil)
	}
}
