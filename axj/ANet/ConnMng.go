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
	REQ_LAST   int32 = 1  // 消息推送检查+
	REQ_KEY    int32 = 2  // 秘钥
	REQ_ACL    int32 = 3  // 请求开启
	REQ_BEAT   int32 = 4  // 心跳
	REQ_ROUTE  int32 = 5  // 路由字典
	REQ_LOOP   int32 = 15 // 连接接受
	REQ_ONEWAY int32 = 16 // 路由处理
)

type ConnM struct {
	ConnC
	id       int64
	initTime int64
	idleTime int64
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

type HandlerM interface {
	Handler
	ConnM(conn Conn) ConnM
}

type ConnMng struct {
	HandlerW
	hanlderM  HandlerM
	idWorker  *Util.IdWorker
	idleTime  int64
	checkTime time.Duration
	connPool  *sync.Pool
	loopTime  int64
	ConnMap   sync.Map
	beatBs    []byte
}

func NewConnMng(handler HandlerM, workerId int32, idleTime time.Duration, checkTime time.Duration, connPool bool) *ConnMng {
	c := new(ConnMng)
	c.handler = handler
	c.hanlderM = handler
	var err error
	c.idWorker, err = Util.NewIdWorker(workerId)
	Kt.Panic(err)
	c.idleTime = int64(idleTime)
	c.checkTime = checkTime
	if connPool {
		c.connPool = new(sync.Pool)
		c.connPool.New = func() interface{} {
			conn := c.New()
			c.Init(conn)
			return conn
		}

	} else {
		c.connPool = nil
	}

	c.loopTime = 0
	c.ConnMap = sync.Map{}
	c.beatBs = c.Processor().Protocol.Rep(REQ_BEAT, "", 0, nil, false, 0)
	return c
}

func (c *ConnMng) ConnM(conn Conn) ConnM {
	return c.hanlderM.ConnM(conn)
}

func (c *ConnMng) OpenConn(client Client) Conn {
	var conn Conn
	if c.connPool == nil {
		conn := c.New()
		c.Init(conn)

	} else {
		conn = c.connPool.Get().(Conn)
	}

	c.Open(conn, client)
	return conn
}

func (c *ConnMng) Init(conn Conn) {
	conn.Get().Init()
	c.handler.Init(conn)
}

func (c *ConnMng) Open(conn Conn, client Client) {
	conn.Get().Open(client, c)
	connM := c.ConnM(conn)
	connM.id = c.idWorker.Generate()
	connM.initTime = time.Now().UnixNano()
	connM.idleTime = connM.initTime
	connM.Get().poolG = APro.PoolOne
	c.handler.Open(conn, client)
}

func (c *ConnMng) OnClose(conn Conn, err error, reason interface{}) {
	if c.connPool != nil {
		c.connPool.Put(conn)
	}

	connM := c.ConnM(conn)
	c.ConnMap.Delete(connM.id)
	c.handler.OnClose(conn, err, reason)
}

func (c *ConnMng) Last(conn Conn, req bool) {
	// 心跳延长
	connM := c.ConnM(conn)
	connM.idleTime = time.Now().UnixNano() + c.idleTime
	c.handler.Last(conn, req)
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
			if connC.Closed() {
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
		pg := connM.ConnC.poolG
		if pg == nil || !pg.StrictAs(poolG) {
			connM.ConnC.poolG = APro.NewPoolLimit(poolG)
		}
	}
}

func (c *ConnMng) UnRegConn(conn Conn, close bool) {
	connM := c.ConnM(conn)
	c.ConnMap.Delete(connM.id)
	if close {
		connM.ConnC.Close(ERR_CLOSED, nil)
	}
}
