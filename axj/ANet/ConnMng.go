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
	REQ_ENTRY  int32 = 1  // 秘钥
	REQ_BEAT   int32 = 2  // 心跳
	REQ_ROUTE  int32 = 3  // 路由字典
	REQ_URI    int32 = 4  // 路由交换
	REQ_READY  int32 = 31 // 准备完毕
	REQ_ONEWAY int32 = 32 // 路由处理
)

type ConnM struct {
	ConnC
	id       int64
	initTime int64
	idleTime int64
	poolG    APro.PoolG
}

func (c *ConnM) Id() int64 {
	return c.id
}

func (c *ConnM) InitTime() int64 {
	return c.initTime
}

func (c *ConnM) IdleTime() int64 {
	return c.idleTime
}

type HandlerH interface {
	Handler
	ConnM(conn Conn) ConnM
}

type ConnMng struct {
	HandlerW
	HandlerH  HandlerH
	idWorker  *Util.IdWorker
	idleTime  int64
	checkTime time.Duration
	loopTime  int64
	ConnMap   sync.Map
	beatBs    []byte
}

func NewConnMng(handler HandlerH, workerId int32, idleTime time.Duration, checkTime time.Duration) *ConnMng {
	c := new(ConnMng)
	c.handler = handler
	c.HandlerH = handler
	var err error
	c.idWorker, err = Util.NewIdWorker(workerId)
	Kt.Panic(err)
	c.idleTime = int64(idleTime)
	c.checkTime = checkTime
	c.loopTime = 0
	c.ConnMap = sync.Map{}
	c.beatBs = c.Processor().Protocol.Rep(REQ_BEAT, "", 0, nil, false, 0)
	return c
}

func (c *ConnMng) ConnM(conn Conn) ConnM {
	return c.HandlerH.ConnM(conn)
}

func (c *ConnMng) Open(client Client) Conn {
	connM := c.handler.Open(client).(*ConnM)
	connM.id = c.idWorker.Generate()
	connM.initTime = time.Now().UnixNano()
	connM.idleTime = connM.initTime
	connM.poolG = nil
	connM.Get().poolG = APro.PoolOne
	return connM
}

func (c *ConnMng) Last(conn Conn, req bool) {
	// 心跳延长
	connM := c.ConnM(conn)
	connM.idleTime = time.Now().UnixNano() + c.idleTime
	c.handler.Last(conn, req)
}

func (c *ConnMng) OnClose(conn Conn, err error, reason interface{}) {
	connM := c.ConnM(conn)
	c.ConnMap.Delete(connM.id)
	c.handler.OnClose(conn, err, reason)
}

func (c *ConnMng) OpenConnM(client Client) *ConnM {
	return OpenConn(client, c).(*ConnM)
}

// 空闲检测
func (c *ConnMng) IdleStop() {
	c.loopTime = -1
}

func (c *ConnMng) IdleLoop() {
	loopTime := time.Now().UnixNano()
	c.loopTime = loopTime
	for loopTime == c.loopTime {
		time.Sleep(c.checkTime)
		time := time.Now().UnixNano()
		c.ConnMap.Range(func(key, value interface{}) bool {
			conn := value.(Conn)
			connM := c.ConnM(conn)
			connC := connM.ConnC
			// 已关闭链接
			if connC.closed {
				c.ConnMap.Delete(key)
				return true
			}

			if connM.idleTime <= time {
				// 直接心跳
				c.Last(conn, false)
				go connC.Rep(-1, "", 0, c.beatBs, false, false, nil)
			}

			return true
		})
	}
}

func (c *ConnMng) RegConn(conn Conn, poolG int) {
	c.Last(conn, true)
	connM := c.ConnM(conn)
	c.ConnMap.Store(connM.id, conn)
	if poolG > 1 {
		connM.ConnC.poolG = APro.NewPoolLimit(poolG)
	}
}

func (c *ConnMng) UnRegConn(conn Conn, close bool) {
	connM := c.ConnM(conn)
	c.ConnMap.Delete(connM.id)
	if close {
		connM.ConnC.Close(ERR_CLOSED, nil)
	}
}
