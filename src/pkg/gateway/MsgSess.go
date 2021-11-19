package gateway

import (
	"axj/ANet"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"errors"
	"gitee.com/absir_admin/cmap"
	"time"
)

type MsgSess struct {
	// 消息组
	grp *MsgGrp
	// 消息队列
	queue *Util.CircleQueue
	// 队列异步执行
	queueAsync *Util.NotifierAsync
	// last通知异步执行
	lastAsync *Util.NotifierAsync
	// last消息队列
	lastQueue *Util.CircleQueue
	// last消息已载入
	lastQueueLoaded bool
	// 主客户端
	client *MsgClient
	// 多客户端
	clientMap *cmap.CMap
	// 客户端数
	clientNum int
}

type ERpc int

const (
	ER_PUSH ERpc = 0
	ER_LAST ERpc = 2
)

var ERR_NOWAY = ANet.ERR_NOWAY
var ERR_FAIL = errors.New("FAIL")

var Result_IdNone = int32(gw.Result_IdNone)

var SubLastSleep = 20 * time.Millisecond

func (that *MsgSess) getOrNewClientMap() *cmap.CMap {
	if that.clientMap == nil {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.clientMap == nil {
			that.clientMap = new(cmap.CMap)
		}
	}

	return that.clientMap
}

func (that *MsgSess) mdfyClientNum(num int) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	that.clientNum = that.clientNum + num
}

// 消息发送返回处理
func (that *MsgSess) OnResult(rep *gw.Id32Rep, err error, rpc ERpc, client *MsgClient, unique string) bool {
	ret := Server.Id32(rep)
	if ret >= R_SUCC_MIN {
		client.idleTime = time.Now().UnixNano() + MsgMng.IdleDrt
		return true

	} else if ret == Result_IdNone {
		// 要不要立刻剔除呢? Conn 和 Close HASH不一致的情况下
		that.grp.Close(client.cid, unique, client.connVer, false)
	}

	// 消息发送失败
	return false
}

// 发送消息执行
func (that *MsgSess) Push(msgD *MsgD, client *MsgClient, unique string, isolate bool) bool {
	if msgD == nil {
		return true
	}

	if msgD.Uri == "" && msgD.Data == nil {
		if msgD.Fid <= 0 {
			return true
		}

		rep, err := Server.GetProdCid(client.cid).GetGWIClient().Push(Server.Context, &gw.PushReq{
			Cid:     client.cid,
			Uri:     msgD.Uri,
			Data:    msgD.Data,
			Id:      msgD.Fid,
			Isolate: isolate,
		})
		return that.OnResult(rep, err, ER_PUSH, client, unique)

	} else {
		pushReg := &gw.PushReq{
			Cid:     client.cid,
			Uri:     msgD.Uri,
			Data:    msgD.Data,
			Id:      msgD.Id,
			Isolate: isolate,
		}

		compress := Server.CidCompress(client.cid)
		if compress {
			cData, cDid := msgD.CData()
			pushReg.Data = cData
			if cDid {
				// 已压缩
				pushReg.CData = 1

			} else {
				// 无法压缩
				pushReg.CData = 2
			}
		}

		rep, err := Server.GetProdCid(client.cid).GetGWIClient().Push(Server.Context, pushReg)
		return that.OnResult(rep, err, ER_PUSH, client, unique)
	}
}

// 获取或创建队列
func (that *MsgSess) getOrNewQueue() *Util.CircleQueue {
	if that.queue == nil && MsgMng.QueueMax > 0 {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.queue == nil {
			that.queueAsync = Util.NewNotifierAsync(that.queueRun, that.grp.locker)
			that.queue = Util.NewCircleQueue(MsgMng.QueueMax)
		}
	}

	return that.queue
}

// 队列通知
func (that *MsgSess) queueRun() {
	client := that.client
	if client == nil || that.queue == nil {
		return
	}

	for {
		msg := that.queueGet()
		if msg == nil {
			break
		}

		if !that.Push(msg.Get(), client, "", true) {
			break
		}

		that.queueRemove(msg)
	}
}

// 队列消息获取
func (that *MsgSess) queueGet() Msg {
	that.grp.rwLocker.RLock()
	defer that.grp.rwLocker.RUnlock()
	for {
		if that.queue.IsEmpty() {
			return nil
		}

		msg, _ := that.queue.Get(0)
		if msg == nil {
			that.queue.Pop()
			continue
		}

		return msg.(Msg)
	}
}

// 队列消息移除
func (that *MsgSess) queueRemove(msg Msg) {
	that.grp.rwLocker.Lock()
	defer that.grp.rwLocker.Unlock()
	that.queue.Remove(msg)
}

// 队列启动
func (that *MsgSess) QueueStart() {
	if that.client == nil || that.queue == nil {
		return
	}

	that.queueAsync.Start(nil)
}

// 队列添加消息
func (that *MsgSess) QueuePush(msg Msg) (bool, error) {
	if msg == nil {
		return false, nil
	}

	if that.getOrNewQueue() == nil {
		client := that.client
		if client != nil {
			msgD := msg.Get()
			rep, err := that.client.gatewayI.Push(Server.Context, &gw.PushReq{
				Cid:     client.cid,
				Uri:     msgD.Uri,
				Data:    msgD.Data,
				Isolate: true,
			})
			return that.OnResult(rep, err, ER_PUSH, client, ""), err
		}

		return false, ERR_NOWAY
	}

	that.grp.rwLocker.Lock()
	defer that.grp.rwLocker.Unlock()
	unique := msg.Unique()
	if unique != "" {
		// 消息唯一标识去重
		for i := that.queue.Size() - 1; i >= 0; i-- {
			m, _ := that.queue.Get(i)
			if m != nil && m.(Msg).Unique() == unique {
				that.queue.Set(i, nil)
				msg = nil
				break
			}
		}
	}

	if msg != nil {
		that.queue.Push(msg, true)
	}

	that.QueueStart()
	return true, nil
}

// last通知异步执行
func (that *MsgSess) LastStart() {
	that.getOrNewLastAsync().Start(nil)
}

// 获取或创建last通知异步执行
func (that *MsgSess) getOrNewLastAsync() *Util.NotifierAsync {
	if that.lastAsync == nil {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.lastAsync == nil {
			that.lastAsync = Util.NewNotifierAsync(that.lastRun, that.grp.locker)
		}
	}

	return that.lastAsync
}

// last通知执行
func (that *MsgSess) lastRun() {
	client := that.client
	if client != nil {
		that.LastClient(client, "")
	}

	if that.clientMap != nil {
		that.clientMap.Range(that.lastRange)
	}
}

func (that *MsgSess) lastRange(key, val interface{}) bool {
	// 数据转换或保障
	if val == nil {
		that.clientMap.Delete(key)
		return true
	}

	client, _ := val.(*MsgClient)
	if client == nil {
		that.clientMap.Delete(key)
		return true
	}

	unique, _ := key.(string)
	if unique == "" {
		that.clientMap.Delete(key)
		return true
	}

	that.LastClient(client, unique)
	return true
}

// 获取或创建last队列
func (that *MsgSess) getOrNewLastQueue() *Util.CircleQueue {
	if that.lastQueue == nil && MsgMng.LastMax > 0 {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.lastQueue == nil {
			that.lastQueue = Util.NewCircleQueue(MsgMng.LastMax)
		}
	}

	return that.lastQueue
}

// last消息队列载入
func (that *MsgSess) lastQueueLoad() {
	if !MsgMng.LastLoad || MsgMng.Db == nil || that.lastQueueLoaded {
		return
	}

	if that.lastQueueLoaded {
		return
	}

	lastQueue := that.getOrNewLastQueue()
	if lastQueue != nil && lastQueue.IsEmpty() {
		// 加强消息不丢失，读写锁
		dbLocker := that.grp.getOrNewDbLocker()
		if dbLocker != nil {
			// 保证写入线程执行执行完成后才载入
			dbLocker.Lock()
			dbLocker.Unlock()
			that.grp.dbLocker = nil
		}

		if that.lastQueueLoaded {
			return
		}

		msgDs := MsgMng.Db.Last(that.grp.gid, MsgMng.LastMax)
		// 锁放在io之后
		that.grp.rwLocker.Lock()
		defer that.grp.rwLocker.Unlock()
		if that.lastQueueLoaded {
			return
		}

		if that.lastQueue.IsEmpty() {
			// 空队列才加入
			mLen := len(msgDs)
			for i := 0; i < mLen; i++ {
				that.lastQueue.Push(&msgDs[i], true)
			}
		}
	}

	that.lastQueueLoaded = true
}

// last消息通知客户端
func (that *MsgSess) LastClient(client *MsgClient, unique string) {
	if client == nil {
		return
	}

	if client.subLastId <= 0 {
		// 尚未loop监听
		return
	}

	lastTime := time.Now().UnixNano()
	client.locker.Lock()
	defer client.locker.Unlock()
	// 保证lastTime递增
	if client.lastTime < lastTime {
		client.lastTime = lastTime

	} else {
		client.lastTime++
	}

	if client.subLastTime > 0 {
		// lastSubRun执行中, 在结束后，会再启动LastClient
		return
	}

	if !client.lasting {
		// 启动通知线程，包含会执行LastSubRun
		go that.LastClientRun(client, unique)
	}
}

// last消息通知客户端进入
func (that *MsgSess) LastClientIn(client *MsgClient) bool {
	client.locker.Lock()
	defer client.locker.Unlock()
	if client.lasting {
		return false
	}

	client.lasting = true
	return true
}

// last消息通知客户端退出
func (that *MsgSess) LastClientOut(client *MsgClient, unique string, lastTime int64) {
	client.locker.Lock()
	defer client.locker.Unlock()
	client.lasting = false
	if client.lastTime > lastTime {
		// 漏掉重启
		go that.LastClientRun(client, unique)
	}
}

// last消息通知客户端完成
func (that *MsgSess) LastClientDone(client *MsgClient, lastTime int64) bool {
	client.locker.Lock()
	defer client.locker.Unlock()
	return client.lastTime <= lastTime
}

// last消息通知客户端执行
func (that *MsgSess) LastClientRun(client *MsgClient, unique string) {
	if !that.LastClientIn(client) {
		return
	}

	lastTime := client.lastTime
	defer that.LastClientOut(client, unique, lastTime)
	for {
		if client.subLastTime > 0 {
			// lastSubRun执行中, 在结束后，会再启动LastClient
			return
		}

		// 不执行lastLoop才通知
		lastTime = client.lastTime
		if client.subLastId > 1 {
			// 执行LastSubRun
			that.subLastRun(client.subLastId, client, unique, client.subContinuous)

		} else {
			// last通知发送
			rep, err := client.gatewayI.Last(Server.Context, &gw.ILastReq{
				Cid:        client.cid,
				Gid:        that.grp.gid,
				ConnVer:    client.connVer,
				Continuous: false,
			})
			that.OnResult(rep, err, ER_LAST, client, unique)
			// 休眠一秒， 防止通知过于频繁|必需
			time.Sleep(time.Second)
		}

		if that.LastClientDone(client, lastTime) {
			break
		}
	}
}

// last消息队列消息订阅
//lastId < 65535, 最近消息数量，
//continuous <= 0 不连续监听 同时lastId<=0时 只是激活Last消息通知
//continuous == 1 连续监听不发送LastC消息
//continuous > 1 多少条间隔发送LastC消息
func (that *MsgSess) SubLast(lastId int64, client *MsgClient, unique string, continuous int32) {
	if client == nil {
		return
	}

	// lastId <= 0 && lastId <= 0
	if continuous <= 0 && lastId <= 0 {
		// 只监听last通知，不接受last消息推送
		client.subLastId = 1
		return
	}

	go that.subLastRun(lastId, client, unique, continuous)
}

// last消息队列消息订阅进入
func (that *MsgSess) subLastIn(client *MsgClient) int64 {
	subLastTime := time.Now().UnixNano()
	client.locker.Lock()
	defer client.locker.Unlock()
	// 保证lastTime递增
	if client.subLastTime < subLastTime {
		client.subLastTime = subLastTime

	} else {
		client.subLastTime++
	}

	return client.subLastTime
}

// last消息队列消息订阅退出
func (that *MsgSess) subLastOut(client *MsgClient, unique string, subLastTime int64, lastTime int64) {
	client.locker.Lock()
	defer client.locker.Unlock()
	if client.subLastTime == subLastTime {
		client.subLastTime = 0
		if lastTime < client.lastTime {
			// last通知触发，last通知或lastSubRun由LastClient负责
			go that.LastClientRun(client, unique)
		}
	}
}

// last消息队列消息订阅完成
func (that *MsgSess) subLastDone(client *MsgClient, lastTime int64) bool {
	client.locker.Lock()
	defer client.locker.Unlock()
	return client.lastTime <= lastTime
}

// last消息队列消息订阅执行
func (that *MsgSess) subLastRun(lastId int64, client *MsgClient, unique string, continuous int32) {
	if client == nil {
		return
	}

	connVer := client.connVer
	lastTime := client.lastTime
	subLastTime := that.subLastIn(client)
	defer that.subLastOut(client, unique, subLastTime, lastTime)
	if lastId < 65535 {
		subLastId := that.lastSubLastId(int(lastId))
		if lastId == subLastId && MsgMng.Db != nil {
			// 从最近多少条开始
			subLastId = MsgMng.Db.LastId(that.grp.gid, int(lastId))
		}

		lastId = subLastId
	}

	client.subLastId = lastId
	client.subContinuous = continuous
	if continuous <= 0 {
		// sub监听下不为0
		client.subLastId = 1

	} else if client.subLastId < 2 {
		// 监听连续发送subLastId最小为2
		client.subLastId = 2
	}

	var pushNum int32 = 0
	// 修复 lastQueue 中有可能会有qs=2的内存消息
	var dbNexted = false
	for subLastTime == client.subLastTime && connVer == client.connVer {
		lastTime = client.lastTime
		msg, lastIn := that.lastQueueGet(client, subLastTime, lastId)
		if !lastIn && MsgMng.Db != nil && !dbNexted {
			msg = nil
		}

		if msg == nil {
			if lastIn || MsgMng.Db == nil {
				// 消息已读取完毕
				if !that.subLastDone(client, lastTime) {
					continue
				}

				break

			} else {
				// 缓冲消息
				msgDs := MsgMng.Db.Next(that.grp.gid, lastId, MsgMng.NextLimit)
				mLen := len(msgDs)
				if mLen <= 0 {
					if !dbNexted {
						dbNexted = true
						continue
					}

					if !that.subLastDone(client, lastTime) {
						// 必须要间隔, 最小毫秒
						time.Sleep(time.Millisecond)
						continue
					}

					break
				}

				for j := 0; j < mLen; j++ {
					if !that.lastQueuePush(subLastTime, client, &msgDs[j], &lastId, unique, false, continuous, &pushNum) {
						return
					}
				}
			}

		} else {
			if !that.lastQueuePush(subLastTime, client, msg, &lastId, unique, true, continuous, &pushNum) {
				return
			}
		}
	}

	if pushNum > 0 {
		rep, err := Server.GetProdCid(client.cid).GetGWIClient().Last(Server.Context, &gw.ILastReq{
			Cid:        client.cid,
			Gid:        that.grp.gid,
			ConnVer:    client.connVer,
			Continuous: true,
		})
		that.OnResult(rep, err, ER_LAST, client, unique)
	}
}

// last消息队列消息lastId计算
func (that *MsgSess) lastSubLastId(lastId int) int64 {
	if lastId < MsgMng.LastMaxAll && lastId >= 0 && that.lastQueue != nil {
		that.grp.rwLocker.RLock()
		defer that.grp.rwLocker.RUnlock()
		size := that.lastQueue.Size()
		if size > lastId {
			val, _ := that.lastQueue.Get(size - lastId - 1)
			msg := val.(Msg)
			return msg.Get().Id
		}
	}

	return int64(lastId)
}

// last消息队列获取 return bool lastIn, 为true 则内存缓存已覆盖lastId，不需要从db读取列表
func (that *MsgSess) lastQueueGet(client *MsgClient, lastLoop int64, lastId int64) (Msg, bool) {
	// 预加载
	that.lastQueueLoad()
	// lock锁提前初始化
	that.getOrNewLastQueue()
	// 锁查找
	that.grp.rwLocker.RLock()
	defer that.grp.rwLocker.RUnlock()
	if client.subLastTime != lastLoop {
		return nil, true
	}

	if that.lastQueue == nil {
		return nil, false
	}

	size := that.lastQueue.Size()
	i := 0
	lastIn := false
	for ; i < size; i++ {
		val, _ := that.lastQueue.Get(i)
		msg := val.(Msg)
		msgD := msg.Get()
		if msgD.Id > lastId {
			return msg, lastIn

		} else {
			lastIn = true
		}
	}

	return nil, lastIn
}

// last消息队列消息推送执行
func (that *MsgSess) lastQueuePush(lastLoop int64, client *MsgClient, msg Msg, lastId *int64, unique string, isolate bool, continuous int32, pushNum *int32) bool {
	if lastLoop != client.subLastTime {
		return false
	}

	msgD := msg.Get()
	if !that.Push(msg.Get(), client, unique, isolate) {
		return false
	}

	// 遍历Next
	*lastId = msgD.Id
	if continuous > 0 {
		// 连续监听
		client.subLastId = *lastId
	}

	if continuous > 1 {
		num := *pushNum + 1
		if num >= continuous {
			rep, err := Server.GetProdCid(client.cid).GetGWIClient().Last(Server.Context, &gw.ILastReq{
				Cid:        client.cid,
				Gid:        that.grp.gid,
				ConnVer:    client.connVer,
				Continuous: true,
			})
			if !that.OnResult(rep, err, ER_LAST, client, unique) {
				return false
			}

			num = 0
		}

		*pushNum = num
	}

	return true
}
