package asdk

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"encoding/json"
	"go.uber.org/zap"
	"net"
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

type Opt interface {
	// 授权数据
	LoginData(adapter *Adapter) []byte
	// 推送数据处理 !uri && !data && tid 为 fid编号消息发送失败
	OnPush(uri string, data []byte, tid int64)
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
	OnState(adapter *Adapter, state int, err string, data []byte)
	// 载入缓存，路由压缩字典
	LoadStorage(name string) string
	// 保存缓存
	SaveStorage(name string, value string)
}

type Client struct {
	locker     sync.Locker
	addr       string
	out        bool
	encry      bool
	compress   bool
	checkDrt   time.Duration
	checkTime  int64
	rqIMax     int32
	opt        Opt
	adapter    *Adapter
	processor  *ANet.Processor
	uriMapUriI map[string]int32
	uriIMapUri map[int32]string
	uriMapHash string
	reqsMap    *sync.Map
	reqIdx     int32
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
}

type req struct {
	adapter *Adapter
	timeout int64
	send    bool
	back    func(string, []byte)
	uri     string
	data    []byte
	encrypt bool
}

func NewClient(addr string, out bool, encry bool, compress bool, checkDrt int32, rqIMax int32, opt Opt) *Client {
	that := new(Client)
	that.locker = new(sync.Mutex)
	that.addr = addr
	that.out = out
	that.encry = encry
	that.compress = compress
	// 检测间隔
	if checkDrt < 1 {
		checkDrt = 1
	}

	that.checkDrt = time.Duration(checkDrt) * time.Second

	// 最大请求编号
	if rqIMax < KtBytes.VINT_1_MAX {
		rqIMax = KtBytes.VINT_1_MAX
	}

	that.rqIMax = rqIMax

	that.opt = opt
	that.uriMapUriI = map[string]int32{}
	that.uriIMapUri = map[int32]string{}
	that.loadUriMapUriI()
	return that
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

	conn, err := net.Dial("tcp", that.addr)
	if that.onError(nil, err) {
		return that.adapter
	}

	adapter := new(Adapter)
	adapter.conn = ANet.NewConnSocket(conn.(*net.TCPConn))
	adapter.locker = new(sync.Mutex)
	that.adapter = adapter
	that.opt.OnState(adapter, CONN, "", nil)
	go that.reqLoop(adapter)
	return adapter
}

func (that *Client) Close() {
	// 关闭check线程
	that.checkTime = 0
	// 关闭连接
	go that.onError(that.adapter, ANet.ERR_CLOSED)
}

func (that *Client) Loop(timeout int32, back func(string, []byte)) {
	that.Req("", nil, false, timeout, back)
}

func (that *Client) Req(uri string, data []byte, encrypt bool, timeout int32, back func(string, []byte)) {
	adapter := that.Conn()
	if adapter == nil {
		// 直接断开
		if back != nil {
			go back("closed", nil)
		}
		return
	}

	// 请求对象
	rq := &req{
		adapter: adapter,
		back:    back,
		uri:     uri,
		data:    data,
		encrypt: encrypt,
	}

	// 超时设置
	if timeout > 0 {
		rq.timeout = time.Now().UnixNano() + int64(timeout)*int64(time.Second)
	}

	// 超时检测
	that.checkStart()

	// 加入请求唯一rqI，需要锁保持唯一
	that.locker.Lock()
	defer that.locker.Unlock()
	rqI := that.reqIdx
	for {
		rqI++
		if rqI <= ANet.REQ_ONEWAY || rqI >= that.rqIMax {
			rqI = ANet.REQ_ONEWAY + 1
		}

		if _, ok := that.reqsMap.Load(rqI); !ok {
			break
		}
	}

	// 加入请求
	that.reqsMap.Store(rqI, rq)
	if adapter.looped {
		// 发送请求
		go that.send(rqI, rq)
	}
}

func (that *Client) checkStart() {
	if that.checkTime == 0 {
		go that.checkLoop()
	}
}

func (that *Client) checkIn() int64 {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.checkTime == 0 {
		checkTime := time.Now().UnixNano()
		that.checkTime = checkTime
		return checkTime
	}

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
	for checkTime == that.checkTime {
		time.Sleep(that.checkDrt)
		that.check()
	}
}

func (that *Client) check() {
	reqsMap := that.reqsMap
	if reqsMap != nil {
		time := time.Now().UnixNano()
		reqsMap.Range(func(key, value interface{}) bool {
			if rq, ok := value.(*req); ok && rq != nil {
				if rq.adapter != that.adapter {
					// adapter 关闭
					that.onRep(key, "closed", nil)

				} else if rq.timeout <= time {
					// 超时
					that.onRep(key, "timeout", nil)

				} else if !rq.send && rq.adapter.looped {
					// 发送
					that.send(key, rq)
				}

			} else {
				// 移除
				reqsMap.Delete(key)
			}

			return true
		})
	}
}

func (that *Client) send(rqI interface{}, rq *req) {
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
		that.onRep(rqI, "", SUCC)
		return
	}

	decryKey := adapter.decryKey
	if !rq.encrypt {
		decryKey = nil
	}

	err := that.processor.Rep(adapter.locker, that.out, adapter.conn, decryKey, that.compress, KtCvt.ToInt32(Kt.If(rq.back == nil, ANet.REQ_ONEWAY, rqI)), rq.uri, 0, rq.data, false, 0)
	if rq.back == nil {
		// 无回调请求
		that.onRep(rqI, "", SUCC)
	}

	// 发送错误处理
	that.onError(adapter, err)
}

func (that *Client) onRep(rqI interface{}, err string, data []byte) {
	if val, ok := that.reqsMap.Load(rqI); ok {
		that.reqsMap.Delete(rqI)
		rq, _ := val.(*req)
		if rq != nil && rq.back != nil {
			go rq.back(err, data)
		}
	}
}

func (that *Client) closeAdapter(adapter *Adapter) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.adapter == adapter {
		that.adapter = nil
		return true
	}

	return false
}

func (that *Client) onError(adapter *Adapter, err error) bool {
	if err == nil {
		return false
	}

	if adapter == nil {
		// 连接错误
		that.opt.OnState(adapter, ERROR, err.Error(), nil)

	} else {
		// 关闭连接
		adapter.conn.Close()
		if that.closeAdapter(adapter) {
			if !adapter.kicked {
				// 未踢开，CLOSE状态通知
				that.opt.OnState(adapter, CLOSE, err.Error(), nil)
			}
		}
	}

	// 请求检查
	that.check()
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

		err := that.processor.Rep(adapter.locker, that.out, adapter.conn, adapter.decryKey, that.compress, 0, "", flag, nil, false, 0)
		if that.onError(adapter, err) {
			return
		}

		that.opt.OnState(adapter, OPEN, "", nil)
	}

	for adapter == that.adapter {
		err, req, uri, uriI, pid, data := that.processor.ReqPId(adapter.conn, adapter.decryKey)
		if that.onError(adapter, err) {
			break
		}

		// catch err
		defer that.reqLoopRecr(adapter)

		// 非返回值 才需要路由压缩解密
		if req < ANet.REQ_ONEWAY {
			if uri == "" && uriI > 0 {
				// 路由压缩解密
				uri = that.uriIMapUri[uriI]
			}
		}

		switch req {
		case ANet.REQ_BEAT:
			// 心跳
			break
		case ANet.REQ_KEY:
			// 传输秘钥
			adapter.decryKey = data
			break
		case ANet.REQ_ROUTE:
			// 路由压缩
			that.setUriMapUriI(KtUnsafe.BytesToString(data), uri, true)
			break
		case ANet.REQ_ACL:
			// 登录请求
			data = that.opt.LoginData(adapter)
			err = that.processor.Rep(adapter.locker, that.out, adapter.conn, adapter.decryKey, that.compress, 0, that.uriMapHash, 1, data, false, 0)
			that.onError(adapter, err)
			break
		case ANet.REQ_LOOP:
			// 连接完成
			that.onLoop(adapter, uri)
			that.opt.OnState(adapter, LOOP, "", data)
			break
		case ANet.REQ_KICK:
			// 被踢
			adapter.kicked = true
			that.adapter = nil
			adapter.conn.Close()
			that.opt.OnState(adapter, KICK, "", data)
			break
		case ANet.REQ_PUSH:
			// 推送消息
			that.opt.OnPush(uri, data, 0)
			break
		case ANet.REQ_PUSHI:
			// 推送消息I
			that.opt.OnPush(uri, data, pid)
			break
		case ANet.REQ_LAST:
			// 推送状态
			that.opt.OnLast(uri, uriI, false)
			break
		case ANet.REQ_LASTC:
			// 推送状态C
			that.opt.OnLast(uri, uriI, true)
			break
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

			that.onRep(req, errS, data)
		}
	}
}

func (that *Client) reqLoopRecr(adapter *Adapter) {
	if err := recover(); err != nil {
		AZap.Logger.Warn("reqLoop err", zap.Reflect("err", err))
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

	that.check()
}
