package gateway

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"errors"
	"time"
)

const (
	MSD_ID_SUB int64 = -1
	MSD_ID_MIN int64 = 65535
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
	// 客户端lastBuffs
	lastBuffs [][]interface{}
	// 客户端lastWait
	lastWait *Util.DoneWait
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
	// 客户端checkBuff
	checkBuff []interface{}
	// 未读消息
	unreads *cmap.CMap
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

func (that *MsgSess) getOrNewClientMap(locker bool) *cmap.CMap {
	if that.clientMap == nil {
		if locker {
			that.grp.locker.Lock()
		}

		if that.clientMap == nil {
			that.clientMap = cmap.NewCMapInit()
		}

		if locker {
			that.grp.locker.Unlock()
		}
	}

	return that.clientMap
}

func (that *MsgSess) dirtyClientNum() {
	that.grp.locker.Lock()
	clientNum := 0
	if that.client != nil {
		clientNum++
	}

	if that.clientMap != nil {
		clientNum += int(that.clientMap.CountFast())
	}
	that.clientNum = clientNum
	that.grp.locker.Unlock()
}

// 消息发送返回处理
func (that *MsgSess) OnResult(rep *gw.Id32Rep, err error, rpc ERpc, client *MsgClient) bool {
	ret := Server.Id32(rep)
	if ret >= R_SUCC_MIN {
		client.idleTime = time.Now().Unix() + _msgMng.IdleDrt
		return true

	} else if ret == Result_IdNone {
		// 要不要立刻剔除呢? Conn 和 Close HASH不一致的情况下
		client.connVer = -1
		that.grp.closeClient(client, client.cid, client.unique, false, false)
	}

	// 消息发送失败
	return false
}

// 发送消息执行
func (that *MsgSess) Push(msgD *MsgD, client *MsgClient, isolate bool) bool {
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
		return that.OnResult(rep, err, ER_PUSH, client)

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
		return that.OnResult(rep, err, ER_PUSH, client)
	}
}

// 获取或创建队列
func (that *MsgSess) getOrNewQueue() *Util.CircleQueue {
	if that.queue == nil && _msgMng.QueueMax > 0 {
		that.grp.locker.Lock()
		if that.queue == nil {
			that.queueAsync = Util.NewNotifierAsync(that.queueRun, that.grp.locker, nil)
			that.queue = Util.NewCircleQueue(_msgMng.QueueMax)
		}

		that.grp.locker.Unlock()
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

		if !that.Push(msg.Get(), client, true) {
			break
		}

		that.queueRemove(msg)
	}
}

// 队列消息获取
func (that *MsgSess) queueGet() Msg {
	that.grp.rwLocker.RLock()
	for {
		if that.queue.IsEmpty() {
			that.grp.rwLocker.RUnlock()
			return nil
		}

		msg, _ := that.queue.Get(0)
		if msg == nil {
			that.queue.Pop()
			continue
		}

		that.grp.rwLocker.RUnlock()
		return msg.(Msg)
	}
}

// 队列消息移除
func (that *MsgSess) queueRemove(msg Msg) {
	that.grp.rwLocker.Lock()
	that.queue.Remove(msg)
	that.grp.rwLocker.Unlock()
}

// 队列启动
func (that *MsgSess) QueueStart() {
	if that.client == nil || that.queue == nil {
		return
	}

	that.queueAsync.Start(nil)
}

// 队列添加消息
func (that *MsgSess) QueuePush(msg Msg, qs int32) (bool, error) {
	if msg == nil {
		return false, nil
	}

	if qs <= 0 || that.getOrNewQueue() == nil {
		client := that.client
		if client != nil {
			msgD := msg.Get()
			rep, err := that.client.gatewayI.Push(Server.Context, &gw.PushReq{
				Cid:     client.cid,
				Uri:     msgD.Uri,
				Data:    msgD.Data,
				Isolate: true,
			})
			return that.OnResult(rep, err, ER_PUSH, client), err
		}

		return false, ERR_NOWAY
	}

	that.grp.rwLocker.Lock()
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

	that.grp.rwLocker.Unlock()
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
		if that.lastAsync == nil {
			that.lastAsync = Util.NewNotifierAsync(that.lastRun, that.grp.locker, nil)
		}

		that.grp.locker.Unlock()
	}

	return that.lastAsync
}

// last通知执行
func (that *MsgSess) lastRun() {
	client := that.client
	if client != nil {
		that.LastClient(client)
	}

	if that.clientMap != nil {
		if _msgMng.LastRangeWait {
			that.clientMap.RangeBuffs(that.lastRange, &that.lastBuffs, Config.ClientPMax, &that.lastWait)

		} else {
			if that.lastBuffs == nil {
				that.lastBuffs = make([][]interface{}, 1)
			}

			that.clientMap.RangeBuff(that.lastRange, &that.lastBuffs[0], Config.ClientPMax)
		}
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

	that.LastClient(client)
	return true
}

// 获取或创建last队列
func (that *MsgSess) getOrNewLastQueue() *Util.CircleQueue {
	if that.lastQueue == nil && _msgMng.LastMax > 0 {
		that.grp.locker.Lock()
		if that.lastQueue == nil {
			that.lastQueue = Util.NewCircleQueue(_msgMng.LastMax)
		}

		that.grp.locker.Unlock()
	}

	return that.lastQueue
}

// last消息队列载入
func (that *MsgSess) lastQueueLoad() {
	if !_msgMng.LastLoad || _msgMng.Db == nil || that.lastQueueLoaded {
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

		msgDs := _msgMng.Db.Last(that.grp.gid, _msgMng.LastMax)
		// 锁放在io之后
		that.grp.rwLocker.Lock()
		if that.lastQueueLoaded {
			that.grp.rwLocker.Unlock()
			return
		}

		if that.lastQueue.IsEmpty() {
			// 空队列才加入
			mLen := len(msgDs)
			for i := 0; i < mLen; i++ {
				that.lastQueue.Push(msgDs[i], true)
			}
		}
	}

	that.lastQueueLoaded = true
	that.grp.rwLocker.Unlock()
}

// last消息通知客户端
func (that *MsgSess) LastClient(client *MsgClient) {
	if client == nil {
		return
	}

	if client.subLastId == 0 {
		// 未读消息
		client.unreadPush()
		// 还未sub监听
		return
	}

	lastTime := time.Now().UnixNano()
	client.locker.Lock()
	// 保证lastTime递增
	if client.lastTime < lastTime {
		client.lastTime = lastTime

	} else {
		client.lastTime++
	}

	if client.subLastTime > 0 {
		// lastSubRun执行中, 在结束后，会再启动LastClient
		client.locker.Unlock()
		return
	}

	if !client.lasting {
		// 启动通知线程，包含会执行LastSubRun
		Util.GoSubmit(func() {
			that.lastClientRun(client)
		})
		client.locker.Unlock()
		return
	}

	client.locker.Unlock()
}

// last消息通知客户端进入
func (that *MsgSess) lastClientIn(client *MsgClient) bool {
	client.locker.Lock()
	if client.lasting {
		client.locker.Unlock()
		return false
	}

	client.lasting = true
	client.locker.Unlock()
	return true
}

// last消息通知客户端退出
func (that *MsgSess) lastClientOut(client *MsgClient, lastTime int64) {
	client.locker.Lock()
	client.lasting = false
	if client.lastTime > lastTime {
		// 漏掉重启
		Util.GoSubmit(func() {
			that.lastClientRun(client)
		})
		client.locker.Unlock()
		return
	}

	client.locker.Unlock()
}

// last消息通知客户端完成
func (that *MsgSess) lastClientDone(client *MsgClient, lastTime int64) bool {
	client.locker.Lock()
	done := client.lastTime <= lastTime
	client.locker.Unlock()
	return done
}

// last消息通知客户端执行
func (that *MsgSess) lastClientRun(client *MsgClient) {
	if !that.lastClientIn(client) {
		return
	}

	lastTime := client.lastTime
	defer that.lastClientOut(client, lastTime)
	for {
		if client.subLastTime > 0 {
			// lastSubRun执行中, 在结束后，会Async启动LastClient
			return
		}

		// 执行lastTime
		lastTime = client.lastTime
		if client.subLastId >= MSD_ID_MIN {
			// 执行LastSubRun
			that.subLastRun(client.subLastId, client, client.subContinuous)

		} else {
			// last通知发送
			rep, err := client.gatewayI.Last(Server.Context, &gw.ILastReq{
				Cid:        client.cid,
				Gid:        that.grp.gid,
				ConnVer:    client.connVer,
				Continuous: false,
			})
			that.OnResult(rep, err, ER_LAST, client)
			// 休眠一秒， 防止通知过于频繁|必需
			time.Sleep(time.Second)
		}

		// 未读消息推送
		client.unreadPush()

		if that.lastClientDone(client, lastTime) {
			break
		}
	}
}

// last消息队列消息订阅
//lastId < 65535, 最近消息数量，
//continuous <= 0 不连续监听 同时lastId<=0时 只是激活Last消息通知
//continuous == 1 连续监听不发送LastC消息
//continuous > 1 多少条间隔发送LastC消息
func (that *MsgSess) SubLast(lastId int64, client *MsgClient, continuous int32) {
	if client == nil {
		return
	}

	AZap.Debug("Grp SubLast %s : %d, %d, %d", that.grp.gid, client.cid, lastId, continuous)

	// lastId <= 0 && lastId <= 0
	if continuous <= 0 && lastId <= 0 {
		// 只监听last通知，不接受subLast消息推送
		client.subLastId = MSD_ID_SUB
		return
	}

	Util.GoSubmit(func() {
		that.subLastRun(lastId, client, continuous)
	})
}

// last消息队列消息订阅进入
func (that *MsgSess) subLastIn(client *MsgClient) int64 {
	subLastTime := time.Now().UnixNano()
	client.locker.Lock()
	// 保证lastTime递增
	if client.subLastTime < subLastTime {
		client.subLastTime = subLastTime

	} else {
		client.subLastTime++
	}

	subLastTime = client.subLastTime
	client.locker.Unlock()
	return subLastTime
}

// last消息队列消息订阅退出
func (that *MsgSess) subLastOut(client *MsgClient, subLastTime int64, lastTime int64) {
	client.locker.Lock()
	if client.subLastTime == subLastTime {
		client.subLastTime = 0
		if lastTime < client.lastTime {
			// last通知触发，last通知或lastSubRun由LastClient负责
			Util.GoSubmit(func() {
				that.lastClientRun(client)
			})
			client.locker.Unlock()
			return
		}
	}

	client.locker.Unlock()
}

// last消息队列消息订阅完成
func (that *MsgSess) subLastDone(client *MsgClient, lastTime int64) bool {
	client.locker.Lock()
	done := client.lastTime <= lastTime
	client.locker.Unlock()
	return done
}

// last消息队列消息订阅执行
func (that *MsgSess) subLastRun(lastId int64, client *MsgClient, continuous int32) {
	if client == nil {
		return
	}

	connVer := client.connVer
	lastTime := client.lastTime
	subLastTime := that.subLastIn(client)
	defer that.subLastOut(client, subLastTime, lastTime)
	if lastId >= 0 && lastId < MSD_ID_MIN {
		subLastId := that.lastSubLastId(int(lastId))
		if lastId == subLastId && _msgMng.Db != nil {
			// 从最近多少条开始
			subLastId = _msgMng.Db.LastId(that.grp.gid, int(lastId))
		}

		lastId = subLastId
	}

	client.subLastId = lastId
	client.subContinuous = continuous
	if continuous <= 0 {
		// sub监听下不为0
		client.subLastId = MSD_ID_SUB

	} else if client.subLastId < MSD_ID_MIN {
		// 监听连续发送subLastId最小为2
		client.subLastId = MSD_ID_MIN
	}

	var pushNum int32 = 0
	// 修复 lastQueue 中有可能会有qs=2的内存消息
	var dbNexted = lastId <= MSD_ID_MIN
	for subLastTime == client.subLastTime && connVer == client.connVer {
		lastTime = client.lastTime
		msg, lastIn := that.lastQueueGet(client, subLastTime, lastId)
		if !lastIn && _msgMng.Db != nil && !dbNexted {
			msg = nil
		}

		if msg == nil {
			if lastIn || _msgMng.Db == nil || dbNexted {
				// 消息已读取完毕
				if !that.subLastDone(client, lastTime) {
					// 必须要间隔, 最小毫秒
					time.Sleep(time.Millisecond)
					continue
				}

				break

			} else {
				// 缓冲消息
				msgDs := _msgMng.Db.Next(that.grp.gid, lastId, _msgMng.NextLimit)
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
					if !that.lastQueuePush(subLastTime, client, msgDs[j], &lastId, false, continuous, &pushNum) {
						return
					}
				}
			}

		} else {
			if !that.lastQueuePush(subLastTime, client, msg, &lastId, true, continuous, &pushNum) {
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
		that.OnResult(rep, err, ER_LAST, client)
	}
}

// last消息队列消息lastId计算
func (that *MsgSess) lastSubLastId(lastId int) int64 {
	if lastId < _msgMng.LastMaxAll && lastId >= 0 && that.lastQueue != nil {
		that.grp.rwLocker.RLock()
		size := that.lastQueue.Size()
		if size > lastId {
			val, _ := that.lastQueue.Get(size - lastId - 1)
			msg := val.(Msg)
			that.grp.rwLocker.RUnlock()
			return msg.Get().Id
		}

		that.grp.rwLocker.RUnlock()
	}

	return int64(lastId)
}

// last消息队列获取 return bool lastIn, 为true 则内存缓存已覆盖lastId，不需要从db读取列表
func (that *MsgSess) lastQueueGet(client *MsgClient, subLastTime int64, lastId int64) (Msg, bool) {
	// 预加载
	that.lastQueueLoad()
	// lock锁提前初始化
	that.getOrNewLastQueue()
	// 锁查找
	that.grp.rwLocker.RLock()
	if client.subLastTime != subLastTime {
		that.grp.rwLocker.RUnlock()
		return nil, true
	}

	if that.lastQueue == nil {
		that.grp.rwLocker.RUnlock()
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
			that.grp.rwLocker.RUnlock()
			return msg, lastIn

		} else {
			lastIn = true
		}
	}

	that.grp.rwLocker.RUnlock()
	return nil, lastIn
}

// last消息队列消息推送执行
func (that *MsgSess) lastQueuePush(subLastTime int64, client *MsgClient, msg Msg, lastId *int64, isolate bool, continuous int32, pushNum *int32) bool {
	if subLastTime != client.subLastTime {
		return false
	}

	msgD := msg.Get()
	if !that.Push(msg.Get(), client, isolate) {
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
			if !that.OnResult(rep, err, ER_LAST, client) {
				return false
			}

			num = 0
		}

		*pushNum = num
	}

	return true
}

// 已读消息
func (that *MsgSess) ReadLastId(gid string, lastId int64) {
	if _msgMng.Db != nil {
		_msgMng.Db.Read(_msgMng.GidForTid(that.grp.gid, gid), lastId)
	}

	unreads := that.unreads
	if unreads != nil {
		val, _ := unreads.Load(gid)
		unread, _ := val.(*SessUnread)
		if unread != nil {
			that.grp.locker.Lock()
			if unread.lastId < lastId {
				unread.lastId = lastId
				unread.num = 0
			}

			that.grp.locker.Unlock()
		}
	}
}

// 未读消息 id > 0 增加一条未读消息
func (that *MsgSess) UnreadRecv(gid string, num int32, lastId int64, uri string, data []byte, entry bool) {
	unreads := that.unreads
	if unreads == nil {
		that.grp.locker.Lock()
		unreads = that.unreads
		if unreads == nil {
			unreads = cmap.NewCMapInit()
			that.unreads = unreads
		}

		that.grp.locker.Unlock()
	}

	val, _ := unreads.Load(gid)
	unread, _ := val.(*SessUnread)
	if unread == nil {
		unread = &SessUnread{num: num}
		unreads.Store(gid, unread)

	} else if num > 0 {
		unread.num = num
	}

	if uri != "" || data != nil {
		unread.data = &SessUnreadData{
			uri:   uri,
			data:  data,
			entry: entry,
		}
	}

	if lastId > 0 {
		if unread.lastId < lastId {
			unread.lastId = lastId
			if num <= 0 {
				// 未读消息数++
				unread.num++
			}
		}
	}

	// 未读版本
	unread.ver = _msgMng.newUnreadVer()
	// 未读消息数通知
	that.LastStart()
}

type SessUnread struct {
	ver    int32
	num    int32
	lastId int64
	data   *SessUnreadData
}

type SessUnreadData struct {
	// 最后一条消息内容
	uri    string
	data   []byte
	cData  []byte
	cDid   bool
	entry  bool
	gidUri string
}
