package asdk

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"container/list"
	"encoding/json"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	CONN  = 0 // 开始连接
	OPEN  = 1 // 连接开启
	LOOP  = 2 // 可以通讯
	CLOSE = 3 // 连接关闭
	ERROR = 4 // 连接错误
	KICK  = 5 // 被剔
)

var SUCC = make([]byte, 0)

func SetBufferPool(pool string) {
	Util.SetBufferPoolsS(pool)
}

type Opt interface {
	// 载入缓存，路由压缩字典
	LoadStorage(name string) string
	// 保存缓存
	SaveStorage(name string, value string)
	// 授权数据
	LoginData(adapter *Adapter) []byte
	// 推送数据处理 !uri && !data && tid 为 fid编号消息发送失败
	OnPush(uri string, data []byte, tid int64, buffer Buffer)
	// 推送消息管道通知 gid 管道编号 connVer 推送消息时，连接版本，调用逻辑服务器Disc方法，附加验证 continues 为发送推送数据时，附加通知
	// 可以在附加消息逻辑 检测当前gid管道 是否监听， 不监听可调用逻辑服务器Disc方法， 防止之前调用逻辑服务器Disc可以未成功的情况
	OnLast(gid string, connVer int32, continues bool)
	// 监听client连接状态编号
	/*
	   gw.state
	   state: {
	       CONN: 0, // 开始连接
	       OPEN: 1, // 连接开启
	       LOOP: 2, // 可以通讯
	       CLOSE: 3, // 连接关闭
	       ERROR: 4, // 连接错误
	       KICK: 5, // 被剔
	   },
	*/
	OnState(adapter *Adapter, state int, err string, data []byte, buffer Buffer)
	// 保留通道消息处理
	OnReserve(adapter *Adapter, req int32, uri string, uriI int32, data []byte, buffer Buffer)
}

type Client struct {
	locker      sync.Locker
	addr        string
	ws          bool
	sendP       bool // 写入内存池
	readP       bool // 读取内存池
	encry       bool
	compress    bool
	checkDrt    time.Duration
	checkTime   int64
	checksAsync *Util.NotifierAsync
	rqIMax      int32
	processor   *ANet.Processor
	opt         Opt
	adapter     *Adapter
	uriMapUriI  map[string]int32
	uriIMapUri  map[int32]string
	uriMapHash  string
	rqAry       *list.List
	rqDict      map[int32]*rqDt
	rqI         int32
	beatIdle    int64
	idleTimeout int64
}

type Adapter struct {
	conn     ANet.Conn
	locker   sync.Locker
	decryKey []byte
	looped   bool
	cid      int64
	unique   string
	gid      string
	closed   bool
	kicked   bool
	lastTime int64
}

func (that *Adapter) IsLooped() bool {
	return that.looped
}

func (that *Adapter) GetCid() int64 {
	return that.cid
}

func (that *Adapter) GetUnique() string {
	return that.unique
}

func (that *Adapter) GetGid() string {
	return that.gid
}

func (that *Adapter) IsClosed() bool {
	return that.closed
}

func (that *Adapter) IsKicked() bool {
	return that.kicked
}

func (that *Adapter) Rep(client *Client, req int32, uri string, uriI int32, data []byte, isolate bool, id int64) error {
	return client.processor.Rep(client.sendP, that.conn, that.decryKey, client.compress, req, uri, uriI, data, isolate, id)
}

type rqDt struct {
	adapter *Adapter
	timeout int64
	send    bool
	back    func(string, []byte, Buffer)
	uri     string
	data    []byte
	encrypt bool
	rqI     int32
}

func NewClient(addr string, sendP bool, readP bool, encry bool, compressMin int, dataMax int, checkDrt int, rqIMax int, opt Opt) *Client {
	that := new(Client)
	that.locker = new(sync.Mutex)
	that.addr = addr
	that.ws = strings.HasPrefix(addr, "ws:") || strings.HasPrefix(addr, "wss:")
	that.sendP = sendP
	that.readP = readP
	that.encry = encry
	that.compress = compressMin > 0
	// 检测间隔
	if checkDrt < 1 {
		checkDrt = 1
	}

	that.checkDrt = time.Duration(checkDrt)

	// 检查异步执行
	that.checksAsync = Util.NewNotifierAsync(that.doChecks, that.locker, nil)

	// 最大请求编号
	that.rqIMax = int32(rqIMax)
	if that.rqIMax < KtBytes.VINT_2_MAX {
		that.rqIMax = KtBytes.VINT_2_MAX
	}

	processor := &ANet.Processor{
		Protocol:    &ANet.ProtocolV{},
		Compress:    &ANet.CompressZip{},
		CompressMin: compressMin,
		Encrypt:     &ANet.EncryptSr{},
		DataMax:     int32(dataMax),
	}

	that.processor = processor
	that.opt = opt
	that.uriMapUriI = map[string]int32{}
	that.uriIMapUri = map[int32]string{}
	that.loadUriMapUriI()
	that.rqAry = new(list.List)
	that.rqDict = map[int32]*rqDt{}
	that.SetIdleTime(30, 0)
	return that
}

// 设置连接地址
func (that *Client) SetAddr(addr string) {
	that.addr = addr
}

// 空闲检查配置
func (that *Client) SetIdleTime(beatIdle int32, idleTimeout int32) {
	that.beatIdle = int64(beatIdle)
	that.idleTimeout = int64(idleTimeout)
}

// interface{}保护，避免sdk导出类型复杂
func (that *Client) GetProcessor() interface{} {
	return that.processor
}

// interface{}保护，避免sdk导出类型复杂
func (that *Client) DialConn() (interface{}, error) {
	if that.ws {
		return wsDial(that.addr)

	} else {
		conn, err := net.Dial("tcp", that.addr)
		if conn == nil || err != nil {
			return nil, err
		}

		return ANet.NewConnSocket(conn.(*net.TCPConn)), err
	}
}

func (that *Client) loadUriMapUriI() {
	uriRoute := that.opt.LoadStorage("uriRoute")
	idx := KtStr.IndexByte(uriRoute, ',', 0)
	if idx >= 0 {
		that.setUriMapUriI(uriRoute[idx+1:], uriRoute[0:idx], false)
	}
}

func (that *Client) setUriMapUriI(uriMapJson string, uriMapHash string, save bool) {
	// uriMapUriI map[string]int32
	json.Unmarshal(KtUnsafe.StringToBytes(uriMapJson), &that.uriMapUriI)
	uriIMapUri := map[int32]string{}
	for key, val := range that.uriMapUriI {
		uriIMapUri[val] = key
	}

	that.uriIMapUri = uriIMapUri
	that.uriMapHash = uriMapHash
	if save {
		that.opt.SaveStorage("uriRoute", uriMapHash+","+uriMapJson)
	}
}

func (that *Client) Conn() *Adapter {
	if that.adapter != nil {
		return that.adapter
	}

	that.locker.Lock()
	defer that.locker.Unlock()
	if that.adapter != nil {
		return that.adapter
	}

	conn, err := that.DialConn()
	aConn, _ := conn.(ANet.Conn)
	if that.onError(nil, err, false) || aConn == nil {
		return that.adapter
	}

	adapter := new(Adapter)
	adapter.conn = aConn
	adapter.locker = new(sync.Mutex)
	// 空闲检查
	adapter.lastTime = time.Now().Unix()
	if that.idleTimeout > 0 {
		that.checkStart()
	}

	that.adapter = adapter
	that.opt.OnState(adapter, CONN, "", nil, nil)
	go that.reqLoop(adapter)
	return adapter
}

func (that *Client) Close() {
	// 关闭check线程
	that.checkTime = 0
	// 关闭连接
	go that.onError(that.adapter, ANet.ERR_CLOSED, true)
}

func (that *Client) Loop(timeout int32, back func(string, []byte, Buffer)) {
	that.Req("", nil, false, timeout, back)
}

func (that *Client) Req(uri string, data []byte, encrypt bool, timeout int32, back func(string, []byte, Buffer)) {
	adapter := that.Conn()
	if adapter == nil {
		// 直接断开
		if back != nil {
			go back("closed", nil, nil)
		}
		return
	}

	// 请求对象
	rq := &rqDt{
		adapter: adapter,
		back:    back,
		uri:     uri,
		data:    data,
		encrypt: encrypt,
	}

	// 超时设置
	if timeout > 0 {
		rq.timeout = time.Now().Unix() + int64(timeout)
	}

	// 请求队列
	that.rqAdd(rq)
	// 超时检测
	that.checkStart()
	// 发送触发
	if adapter.looped {
		that.checksAsync.Start(nil)
	}
}

func (that *Client) rqGet(rqI int32) *rqDt {
	that.locker.Lock()
	dt := that.rqDict[rqI]
	that.locker.Unlock()
	return dt
}

func (that *Client) rqDelDict(rqI int32) {
	if rqI <= ANet.REQ_ONEWAY {
		return
	}

	that.locker.Lock()
	delete(that.rqDict, rqI)
	that.locker.Unlock()
}

func (that *Client) onRep(rqI int32, rq *rqDt, err string, data []byte, buffer Buffer) {
	if rq == nil {
		rq = that.rqGet(rqI)

	} else if rqI <= 0 {
		rqI = rq.rqI
	}

	if rq != nil {
		// 已发送
		rq.send = true
		back := rq.back
		that.rqDelDict(rqI)
		if back != nil {
			// 已回调
			rq.back = nil
			defer that.onRepRcvr()
			back(err, data, buffer)
			return
		}
	}

	BufferFree(buffer)
}

func (that *Client) onRepRcvr() {
	if err := recover(); err != nil {
		AZap.LoggerS.Warn("OnRep Err", zap.Reflect("err", err))
	}
}

func (that *Client) rqAdd(rq *rqDt) {
	// 加入请求唯一rqI，需要锁保持唯一
	that.locker.Lock()
	if rq.back == nil {
		rq.rqI = ANet.REQ_ONEWAY

	} else {
		rqI := that.rqI
		var num int32 = 0
		for {
			rqI++
			if rqI <= ANet.REQ_ONEWAY || rqI >= that.rqIMax {
				rqI = ANet.REQ_ONEWAY + 1
			}

			if that.rqDict[rqI] == nil {
				break
			}

			num++
			if num >= that.rqIMax {
				num = 0
				time.Sleep(time.Millisecond)
			}
		}

		that.rqI = rqI
		rq.rqI = rqI
		that.rqDict[rqI] = rq
	}

	that.rqAry.PushBack(rq)
	that.locker.Unlock()
}

func (that *Client) rqNext(el *list.Element) *list.Element {
	that.locker.Lock()
	if el == nil {
		el = that.rqAry.Front()

	} else {
		el = el.Next()
	}

	that.locker.Unlock()
	return el
}

func (that *Client) rqDelAry(el *list.Element) {
	that.locker.Lock()
	that.rqAry.Remove(el)
	that.locker.Unlock()
}

func (that *Client) doChecks() {
	adapter := that.adapter
	looped := adapter != nil && adapter.looped
	time := time.Now().Unix()
	var el *list.Element
	nxt := that.rqNext(nil)
	for {
		if el = nxt; el == nil {
			break
		}

		nxt = that.rqNext(el)
		rq, _ := el.Value.(*rqDt)
		if rq == nil {
			that.rqDelAry(el)
			continue
		}

		if rq.adapter != adapter {
			// 请求已关闭
			that.onRep(0, rq, "closed", nil, nil)

		} else if rq.timeout <= time {
			// 请求已关闭
			that.onRep(0, rq, "timeout", nil, nil)

		} else if !rq.send && looped {
			// 发送时间
			adapter.lastTime = time
			// 发送请求
			that.rqSend(rq)
		}

		if rq.send && rq.back == nil {
			// 已发送返回, 队列删除
			that.rqDelAry(el)
		}
	}

	if adapter != nil {
		// 心跳超时检查
		if adapter.looped && that.beatIdle > 0 && adapter.lastTime < (time-that.beatIdle) {
			adapter.lastTime = time
			that.processor.Rep(that.sendP, adapter.conn, nil, that.compress, ANet.REQ_BEAT, "", 0, nil, false, 0)
		}

		// 接收超时检查
		if that.idleTimeout > 0 && adapter.lastTime < (time-that.idleTimeout) {
			that.onError(adapter, Kt.NewErrReason("idleTimeout"), true)
		}
	}
}

func (that *Client) rqSend(rq *rqDt) {
	if rq.send {
		return
	}

	adapter := rq.adapter
	if !adapter.looped {
		return
	}

	rq.send = true
	if rq.uri == "" && rq.data == nil {
		// 无数据请求 loop回调
		if rq.back != nil {
			that.onRep(0, rq, "", SUCC, nil)
		}

		return
	}

	decryKey := adapter.decryKey
	if !rq.encrypt {
		decryKey = nil
	}

	err := that.processor.Rep(that.sendP, adapter.conn, decryKey, that.compress, rq.rqI, rq.uri, 0, rq.data, false, 0)
	// 发送错误处理
	that.onError(adapter, err, true)
}

func (that *Client) checkStart() {
	if that.checkTime == 0 {
		go that.checkLoop()
	}
}

func (that *Client) checkIn() int64 {
	that.locker.Lock()
	if that.checkTime == 0 {
		checkTime := time.Now().Unix()
		if that.checkTime >= checkTime {
			checkTime = that.checkTime + 1
		}

		that.checkTime = checkTime
		that.locker.Unlock()
		return checkTime
	}

	that.locker.Unlock()
	return 0
}

func (that *Client) checkOut() {
	that.checkTime = 0
}

func (that *Client) checkLoop() {
	checkTime := that.checkIn()
	if checkTime <= 0 {
		return
	}

	defer that.checkOut()
	checkDrt := that.checkDrt * time.Second
	for Kt.Active && checkTime == that.checkTime {
		time.Sleep(checkDrt)
		that.checksAsync.Start(nil)
	}
}

func (that *Client) closeAdapter(adapter *Adapter, lock bool) bool {
	if lock {
		that.locker.Lock()
	}

	if that.adapter == adapter {
		that.adapter = nil
		if lock {
			that.locker.Unlock()
		}

		return true
	}

	if lock {
		that.locker.Unlock()
	}

	return false
}

func (that *Client) onError(adapter *Adapter, err error, lock bool) bool {
	if err == nil {
		return false
	}

	if adapter == nil {
		// 连接错误
		that.opt.OnState(adapter, ERROR, err.Error(), nil, nil)

	} else {
		// 关闭连接
		adapter.conn.Close(true)
		if that.closeAdapter(adapter, lock) {
			if !adapter.kicked {
				// 未踢开，CLOSE状态通知
				that.opt.OnState(adapter, CLOSE, err.Error(), nil, nil)
			}

			that.checksAsync.StartLock(nil, lock)
		}
	}

	return true
}

func (that *Client) reqLoop(adapter *Adapter) {
	{
		// 客户端状态发送
		var flag int32 = 0
		if that.encry {
			flag |= ANet.FLG_ENCRYPT
		}

		if that.compress {
			flag |= ANet.FLG_COMPRESS
		}

		err := that.processor.Rep(that.sendP, adapter.conn, adapter.decryKey, that.compress, 0, "", flag, nil, false, 0)
		if that.onError(adapter, err, true) {
			return
		}

		that.opt.OnState(adapter, OPEN, "", nil, nil)
	}

	// 读取内存池
	var buffer *KtBuffer.Buffer = nil
	var pBuffer **KtBuffer.Buffer = nil
	if that.readP {
		pBuffer = &buffer
	}

	// catch err
	defer that.reqLoopRcvr(adapter, buffer)
	for adapter == that.adapter {
		Util.PutBuffer(buffer)
		buffer = nil
		err, req, uri, uriI, pid, data := that.processor.ReqPId(pBuffer, adapter.conn, adapter.decryKey)
		if that.onError(adapter, err, true) {
			break
		}

		// 接收时间
		adapter.lastTime = time.Now().Unix()
		// 非返回值 才需要路由压缩解密
		if req < ANet.REQ_ONEWAY {
			if uri == "" && uriI > 0 {
				// 路由压缩解密
				uri = that.uriIMapUri[uriI]
			}
		}

		switch req {
		case ANet.REQ_PUSH:
			// 推送消息
			that.opt.OnPush(uri, data, 0, buffer)
			buffer = nil
			continue
		case ANet.REQ_PUSHI:
			// 推送消息I
			that.opt.OnPush(uri, data, pid, buffer)
			buffer = nil
			continue
		case ANet.REQ_KICK:
			// 被踢
			adapter.kicked = true
			that.adapter = nil
			adapter.conn.Close(true)
			that.opt.OnState(adapter, KICK, "", data, buffer)
			buffer = nil
			continue
		case ANet.REQ_LAST:
			// 推送状态
			that.opt.OnLast(uri, uriI, false)
			continue
		case ANet.REQ_LASTC:
			// 推送状态C
			that.opt.OnLast(uri, uriI, true)
			continue
		case ANet.REQ_KEY:
			// 传输秘钥
			if that.readP {
				data = KtBytes.Copy(data)
			}

			adapter.decryKey = data
			continue
		case ANet.REQ_ACL:
			// 登录请求
			data = that.opt.LoginData(adapter)
			err = that.processor.Rep(that.sendP, adapter.conn, adapter.decryKey, that.compress, 0, that.uriMapHash, 1, data, false, 0)
			that.onError(adapter, err, true)
			continue
		case ANet.REQ_BEAT:
			// 心跳
			continue
		case ANet.REQ_ROUTE:
			// 路由压缩
			that.setUriMapUriI(KtUnsafe.BytesToString(data), uri, true)
			continue
		case ANet.REQ_LOOP:
			// 连接完成
			that.onLoop(adapter, uri)
			that.opt.OnState(adapter, LOOP, "", data, buffer)
			buffer = nil
			// 心跳检查
			if that.beatIdle > 0 {
				that.checkStart()
			}
			continue
		}

		if req > ANet.REQ_ONEWAY {
			errS := ""
			switch uriI {
			case 0:
				break
			case 1:
				errS = "PROD_NO"
				break
			case 2:
				errS = "PROD_ERR"
				break
			default:
				errS = "SERV_ERR"
				break
			}

			that.onRep(req, nil, errS, data, buffer)
			buffer = nil

		} else {
			that.opt.OnReserve(adapter, req, uri, uriI, data, buffer)
			buffer = nil
		}
	}

	Util.PutBuffer(buffer)
}

func (that *Client) reqLoopRcvr(adapter *Adapter, buffer *KtBuffer.Buffer) {
	// 内存池回收
	Util.PutBuffer(buffer)
	if err := recover(); err != nil {
		AZap.LoggerS.Warn("ReqLoop Err", zap.Reflect("err", err))
	}
}

func (that *Client) onLoop(adapter *Adapter, uri string) {
	adapter.looped = true
	ids := KtStr.SplitByte(uri, '/', false, 0, 3)
	if len(ids) > 0 {
		adapter.cid = KtCvt.ToInt64(ids[0])
	}

	if len(ids) > 1 {
		adapter.unique = ids[1]
	}

	if len(ids) > 2 {
		adapter.gid = ids[2]
	}

	that.checksAsync.Start(nil)
}
