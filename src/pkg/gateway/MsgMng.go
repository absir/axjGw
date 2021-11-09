package gateway

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

type msgMng struct {
	QueueMax  int           // 主client消息队列大小
	NextLimit int           // last消息，单次读取列表数
	LastLimit int           // last消息队类，初始化载入列表数
	LastMax   int           // last消息队列大小
	LastLoad  bool          // 是否执行 last消息队类，初始化载入列表数
	LastUri   string        // 消息持久化，数据库连接
	ClearCron string        // 消息清理，执行周期
	ClearDay  int64         // 清理消息过期天数
	CheckDrt  time.Duration // 执行检查逻辑，间隔
	LiveDrt   int64         // 连接断开，存活时间
	IdleDrt   int64         // 连接检查，间隔
	checkLoop int64
	checkTime int64
	Db        MsgDb
	idWorkder *Util.IdWorker
	locker    sync.Locker
	connVer   int32
	grpMap    *sync.Map
}

// 初始变量
var MsgMng *msgMng

const connVerMax = KtBytes.VINT_3_MAX - 1

func initMsgMng() {
	// 消息管理配置
	MsgMng = &msgMng{
		QueueMax:  20,
		NextLimit: 10,
		LastMax:   21, // over load cover msgs
		LastLoad:  true,
		LastUri:   "",
		ClearCron: "0 0 3 * * *",
		ClearDay:  30,
		CheckDrt:  5000,
		LiveDrt:   15000,
		IdleDrt:   30000,
	}

	// 配置处理
	APro.SubCfgBind("msg", MsgMng)
	that := MsgMng
	that.LastLoad = that.LastLoad && that.LastMax > 0
	that.CheckDrt = that.CheckDrt * time.Millisecond
	that.LiveDrt = that.LiveDrt * int64(time.Millisecond)
	that.IdleDrt = that.IdleDrt * int64(time.Millisecond)

	// 属性初始化
	that.idWorkder = Util.NewIdWorkerPanic(Config.WorkId)
	that.locker = new(sync.Mutex)
	that.grpMap = new(sync.Map)

	// 消息持久化
	if that.LastUri != "" {
		db, err := gorm.Open(mysql.Open(that.LastUri), &gorm.Config{})
		Kt.Panic(err)

		msgGorm := &MsgGorm{
			db: db,
		}
		// 自动创建表
		msgGorm.AutoMigrate()
		that.Db = msgGorm
		if that.ClearCron != "" && !strings.HasPrefix(that.ClearCron, "#") {
			Server.Cron().AddFunc(that.ClearCron, that.ClearPass)
			that.ClearPass()
		}
	}
}

type MsgGrp struct {
	gid      string
	ghash    int
	locker   *sync.RWMutex
	laster   sync.Locker
	passTime int64
	sess     *MsgSess
}

type MsgSess struct {
	grp        *MsgGrp
	queue      *Util.CircleQueue
	queuing    bool
	lastQueue  *Util.CircleQueue
	lastLoaded bool
	client     *MsgClient
	clientMap  *sync.Map
	clientNum  int
}

type MsgClient struct {
	cid        int64
	gatewayI   gw.GatewayIClient
	connVer    int32
	idleTime   int64
	lasting    bool
	lastTime   int64
	lastLoop   int64
	lastId     int64
	continuous int32
	cidReq     *gw.CidReq
}

func (that *MsgClient) ConnVer() int32 {
	return that.connVer
}

func (that *msgMng) ClearPass() {
	if that.Db != nil && that.ClearDay > 0 {
		oId := that.idWorkder.Timestamp(time.Now().UnixNano() - that.ClearDay*24*int64(time.Hour))
		that.Db.Clear(oId)
	}
}

// 空闲检测
func (that *msgMng) CheckStop() {
	that.checkLoop = -1
}

func (that *msgMng) CheckLoop() {
	loopTime := time.Now().UnixNano()
	that.checkLoop = loopTime
	for loopTime == that.checkLoop {
		time.Sleep(that.CheckDrt)
		that.checkTime = time.Now().UnixNano()
		that.grpMap.Range(that.checkRange)
	}
}

func (that *msgMng) checkRange(key interface{}, val interface{}) bool {
	msgGrp, _ := val.(*MsgGrp)
	that.checkGrp(key, msgGrp)
	return true
}

func (that *msgMng) checkGrp(key interface{}, grp *MsgGrp) {
	if grp == nil {
		that.grpMap.Delete(key)
		return
	}

	clientNum := 0
	if Server.IsProdHash(grp.ghash) {
		sess := grp.sess
		if sess != nil && sess.clientNum > 0 {
			// 客户端连接
			grp.checkClients()
			clientNum = sess.clientNum
		}
	}

	time := that.checkTime
	if grp.passTime < time {
		if clientNum > 0 {
			// 还有客户端连接
			grp.retain()
			return
		}

		that.locker.Lock()
		defer that.locker.Unlock()
		grp.locker.Lock()
		defer grp.locker.Unlock()
		if grp.passTime > time {
			return
		}

		that.grpMap.Delete(key)
	}
}

func (that *msgMng) MsgGrp(gid string) *MsgGrp {
	val, _ := that.grpMap.Load(gid)
	msgGrp, _ := val.(*MsgGrp)
	return msgGrp
}

func (that *msgMng) GetMsgGrp(gid string) *MsgGrp {
	that.locker.Lock()
	defer that.locker.Unlock()
	val, _ := that.grpMap.Load(gid)
	grp, _ := val.(*MsgGrp)
	if grp != nil {
		if Server.IsProdHash(grp.ghash) {
			grp.retain()
		}

	} else {
		grp = that.newMsgGrp(gid)
		that.grpMap.Store(gid, grp)
	}

	return grp
}

func (that *msgMng) newMsgGrp(gid string) *MsgGrp {
	grp := new(MsgGrp)
	grp.gid = gid
	grp.ghash = Kt.HashCode(KtUnsafe.StringToBytes(gid))
	grp.locker = new(sync.RWMutex)
	grp.retain()
	return grp
}

func (that *msgMng) newConnVer() int32 {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.connVer < R_SUCC_MIN || that.connVer >= connVerMax {
		that.connVer = R_SUCC_MIN

	} else {
		that.connVer++
	}

	return that.connVer
}

func (that *MsgGrp) Sess() *MsgSess {
	return that.sess
}

func (that *MsgGrp) retain() {
	that.passTime = time.Now().UnixNano() + MsgMng.LiveDrt
}

func (that *MsgGrp) newMsgSess() *MsgSess {
	sess := new(MsgSess)
	sess.grp = that
	return sess
}

func (that *MsgGrp) newMsgClient(cid int64) *MsgClient {
	client := new(MsgClient)
	client.cid = cid
	client.gatewayI = Server.GetProdCid(cid).GetGWIClient()
	client.idleTime = time.Now().UnixNano() + that.passTime
	return client
}

func (that *MsgGrp) getLaster() sync.Locker {
	if that.laster == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.laster == nil {
			that.laster = new(sync.Mutex)
		}
	}

	return that.laster
}

func (that *MsgGrp) getSess(create bool) *MsgSess {
	if that.sess == nil && create {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.sess == nil {
			that.sess = that.newMsgSess()
		}
	}

	return that.sess
}

func (that *MsgGrp) getClient(unique string) *MsgClient {
	sess := that.sess
	if sess == nil {
		return nil
	}

	if unique == "" {
		return sess.client
	}

	clientMap := sess.clientMap
	if clientMap == nil {
		return nil
	}

	client, _ := clientMap.Load(unique)
	msgClient, _ := client.(*MsgClient)
	return msgClient
}

func (that *MsgGrp) closeOld(old *MsgClient, cid int64, unique string, kick bool) bool {
	sess := that.sess
	if sess == nil {
		return false
	}

	if old == nil || (cid > 0 && cid != old.cid) {
		return false
	}

	if unique == "" {
		sess.client = nil

	} else {
		sess.clientMap.Delete(unique)
	}

	old.connVer = 0
	sess.clientNum--
	if kick {
		// 关闭通知
		go old.gatewayI.Kick(Server.Context, &gw.KickReq{Cid: old.cid}, nil)
	}

	return true
}

func (that *MsgGrp) checkClients() {
	sess := that.sess
	if sess == nil {
		return
	}

	that.checkClient(sess.client, "")
	if sess.clientMap != nil {
		sess.clientMap.Range(that.checkRange)
	}
}

func (that *MsgGrp) checkRange(key, val interface{}) bool {
	client, _ := val.(*MsgClient)
	unique, _ := key.(string)
	that.checkClient(client, unique)
	return true
}

func (that *MsgGrp) checkClient(client *MsgClient, unique string) {
	if client == nil {
		return
	}

	if client.idleTime > MsgMng.checkTime {
		// 未空闲
		return
	}

	rep, _ := client.gatewayI.Alive(Server.Context, client.getCidReq())
	ret := Server.Id32(rep)
	if ret < R_SUCC_MIN {
		that.locker.Lock()
		defer that.locker.Unlock()
		that.closeOld(client, client.cid, unique, false)
	}
}

func (that *MsgGrp) Conn(cid int64, unique string, kick bool, newVer bool) *MsgClient {
	client := that.getClient(unique)
	if client != nil {
		if client.cid == cid {
			if newVer {
				client.connVer = MsgMng.newConnVer()
			}

			return client

		} else if client.cid > cid {
			return nil
		}

		that.closeOld(client, client.cid, unique, kick)
	}

	sess := that.getSess(true)
	client = that.newMsgClient(cid)
	client.connVer = MsgMng.newConnVer()
	if unique == "" {
		sess.client = client

	} else {
		if sess.clientMap == nil {
			sess.clientMap = new(sync.Map)
		}

		sess.clientMap.Store(unique, client)
	}

	sess.clientNum++
	return client
}

func (that *MsgGrp) Close(cid int64, unique string, connVer int32, kick bool) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	client := that.getClient(unique)
	if client != nil && (connVer == 0 || connVer == client.connVer) {
		return that.closeOld(client, cid, unique, kick)
	}

	return false
}

func (that *MsgGrp) Clear(queue bool, last bool) {
	sess := that.sess
	if sess != nil {
		locked := false
		if queue && sess.queue != nil {
			if !locked {
				locked = true
				that.locker.Lock()
				defer that.locker.Unlock()
			}

			sess.queue.Clear()
		}

		if last && sess.lastQueue != nil {
			if !locked {
				locked = true
				that.locker.Lock()
				defer that.locker.Unlock()
			}

			sess.lastQueue.Clear()
			sess.lastLoaded = false
		}
	}
}

func (that *MsgGrp) Push(uri string, data []byte, isolate bool, qs int32, queue bool, unique string, fid int64) (int64, bool, error) {
	if qs >= 2 {
		if MsgMng.LastMax <= 0 {
			qs = 1
		}
	}

	sess := that.getSess(queue)
	if qs <= 1 {
		if sess == nil {
			return 0, false, ERR_NOWAY
		}

		msg := NewMsg(uri, data, unique)
		succ, err := sess.QueuePush(msg)
		return msg.Get().Id, succ, err

	} else {
		if qs >= 0 && fid > 0 && MsgMng.Db != nil {
			// 唯一性校验
			id := MsgMng.Db.FidGet(fid, that.gid)
			if id > 0 {
				return id, true, nil
			}
		}

		var err error = nil
		msg := NewMsg(uri, data, unique)
		that.lastPush(sess, msg, fid)
		msgD := msg.Get()
		if qs >= 3 && MsgMng.Db != nil {
			err = that.insertMsgD(msgD)
		}

		if sess != nil && msg.Get().Id > 0 {
			sess.LastsStart()
		}

		return msg.Get().Id, true, err
	}
}

func (that *MsgGrp) insertMsgD(msgD *MsgD) error {
	if msgD.Id > 0 {
		// sess.lastQueue加强不漏消息
		return MsgMng.Db.Insert(msgD)

	} else {
		// laster加强消息顺序写入
		that.getLaster().Lock()
		defer that.getLaster().Unlock()
		msgD.Id = MsgMng.idWorkder.Generate()
		return MsgMng.Db.Insert(msgD)
	}
}

// 插入顺序消息
func (that *MsgGrp) lastPush(sess *MsgSess, msg Msg, fid int64) {
	msgD := msg.Get()
	msgD.Gid = that.gid
	msgD.Fid = fid

	// 顺序队列
	lastQueue := sess != nil && sess.getLastQueue() != nil
	if !lastQueue {
		return
	}

	sess.lastLoad()
	// 锁加入队列
	that.locker.Lock()
	defer that.locker.Unlock()
	msgD.Id = MsgMng.idWorkder.Generate()
	sess.lastQueue.Push(msg, true)
}

type ERpc int

const (
	ER_PUSH ERpc = 0
	ER_LAST ERpc = 2
)

var ERR_NOWAY = errors.New("ERR_NOWAY")
var ERR_FAIL = errors.New("ERR_FAIL")

func (that *MsgSess) OnResult(rep *gw.Id32Rep, err error, rpc ERpc, client *MsgClient, unique string) bool {
	if Server.Id32(rep) >= R_SUCC_MIN {
		client.idleTime = time.Now().UnixNano() + MsgMng.IdleDrt
		return true
	}

	// 消息发送失败
	return false
}

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
		rep, err := Server.GetProdCid(client.cid).GetGWIClient().Push(Server.Context, &gw.PushReq{
			Cid:     client.cid,
			Uri:     msgD.Uri,
			Data:    msgD.Data,
			Id:      msgD.Id,
			Isolate: isolate,
		})
		return that.OnResult(rep, err, ER_PUSH, client, unique)
	}
}

func (that *MsgSess) getQueue() *Util.CircleQueue {
	if that.queue == nil && MsgMng.QueueMax > 0 {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.queue == nil {
			that.queue = Util.NewCircleQueue(MsgMng.QueueMax)
		}
	}

	return that.queue
}

func (that *MsgSess) getLastQueue() *Util.CircleQueue {
	if that.lastQueue == nil && MsgMng.LastMax > 0 {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.lastQueue == nil {
			that.lastQueue = Util.NewCircleQueue(MsgMng.LastMax)
		}
	}

	return that.lastQueue
}

func (that *MsgSess) QueuePush(msg Msg) (bool, error) {
	if msg == nil {
		return false, nil
	}

	if that.getQueue() == nil {
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

	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
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

func (that *MsgSess) QueueStart() {
	if that.client == nil || that.queuing {
		return
	}

	go that.queueRun()
}

func (that *MsgSess) queueIn() bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	if that.queuing {
		return false
	}

	that.queuing = true
	return true
}

func (that *MsgSess) queueOut() {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	that.queuing = false
}

func (that *MsgSess) queueRun() {
	client := that.client
	if client == nil {
		return
	}

	if !that.queueIn() {
		return
	}

	defer that.queueOut()
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

func (that *MsgSess) queueGet() Msg {
	that.grp.locker.RLock()
	defer that.grp.locker.RUnlock()
	msg, _ := that.queue.Get(0)
	if msg == nil {
		return nil
	}

	return msg.(Msg)
}

func (that *MsgSess) queueRemove(msg Msg) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	that.queue.Remove(msg)
}

func (that *MsgSess) LastsStart() {
	that.LastStart(that.client, "")
	if that.clientMap != nil {
		that.clientMap.Range(that.LastsStartRange)
	}
}

func (that *MsgSess) LastsStartRange(key, val interface{}) bool {
	client, _ := val.(*MsgClient)
	unique, _ := key.(string)
	that.LastStart(client, unique)
	return true
}

func (that *MsgSess) LastStart(client *MsgClient, unique string) {
	if client == nil {
		return
	}

	if client.lastId <= 0 {
		// 尚未loop监听
		return
	}

	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	client.lastTime = time.Now().UnixNano()
	if client.lastLoop > 0 && client.lastId > 1 {
		// lastLoop执行中
		return
	}

	if !client.lasting {
		go that.lastRun(client, unique)
	}
}

func (that *MsgSess) lastIn(client *MsgClient) bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	if client.lasting {
		return false
	}

	client.lasting = true
	return true
}

func (that *MsgSess) lastOut(client *MsgClient) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	client.lasting = false
}

func (that *MsgSess) lastDone(client *MsgClient, lastTime int64) bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	return client.lastTime <= lastTime
}

func (that *MsgSess) lastRun(client *MsgClient, unique string) {
	if !that.lastIn(client) {
		return
	}

	defer that.lastOut(client)
	for {
		if client.lastLoop > 0 {
			continue
		}

		// 不执行lastLoop才通知
		lastTime := client.lastTime
		if client.lastId > 0 {
			that.Lasts(client.lastId, client, unique, client.continuous)

		} else {
			rep, err := client.gatewayI.Last(Server.Context, &gw.ILastReq{
				Cid:        client.cid,
				Gid:        that.grp.gid,
				ConnVer:    client.connVer,
				Continuous: false,
			})
			that.OnResult(rep, err, ER_LAST, client, unique)
		}

		// 休眠一秒， 防止通知过于频繁|必需
		time.Sleep(time.Second)
		if that.lastDone(client, lastTime) {
			break
		}
	}
}

func (that *MsgSess) Lasts(lastId int64, client *MsgClient, unique string, continuous int32) {
	if client == nil {
		return
	}

	if continuous > 0 && lastId == 1 {
		// 光loop监听，last通知
		client.lastId = 1
		return
	}

	go that.lastLoop(lastId, client, unique, continuous)
}

func (that *MsgSess) lastLoop(lastId int64, client *MsgClient, unique string, continuous int32) {
	if client == nil {
		return
	}

	lastTime := client.lastTime
	lastLoop := that.inLastLoop(client)
	connVer := client.connVer
	defer that.outLastLoop(client, unique, lastLoop, lastTime)
	if lastId < 65535 && MsgMng.Db != nil {
		// 从最近多少条开始
		lastId = MsgMng.Db.LastId(that.grp.gid, int(lastId))
	}

	client.continuous = continuous
	if continuous <= 0 {
		// loop监听下不为0
		client.lastId = 1

	} else if client.lastId < 2 {
		// 最小连续监听id为2
		client.lastId = 2
	}

	var pushNum int32 = 0
	for lastLoop == client.lastLoop && connVer == client.connVer {
		lastTime = client.lastTime
		msg, lastIn := that.lastGet(client, lastLoop, lastId)
		if !lastIn && MsgMng.Db != nil {
			msg = nil
		}

		if msg == nil {
			if lastIn || MsgMng.Db == nil {
				// 消息已读取完毕
				if that.conLastLoop(client, lastTime) {
					continue
				}

				break

			} else {
				// 缓冲消息
				msgDs := MsgMng.Db.Next(that.grp.gid, lastId, MsgMng.NextLimit)
				mLen := len(msgDs)
				if mLen <= 0 {
					if that.conLastLoop(client, lastTime) {
						continue
					}

					break
				}

				for j := 0; j < mLen; j++ {
					if !that.lastMsg(lastLoop, client, &msgDs[j], &lastId, unique, false, continuous, &pushNum) {
						return
					}
				}
			}

		} else {
			if !that.lastMsg(lastLoop, client, msg, &lastId, unique, true, continuous, &pushNum) {
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

func (that *MsgSess) lastMsg(lastLoop int64, client *MsgClient, msg Msg, lastId *int64, unique string, isolate bool, continuous int32, pushNum *int32) bool {
	if lastLoop != client.lastLoop {
		return false
	}

	msgD := msg.Get()
	if !that.Push(msg.Get(), client, "", isolate) {
		return false
	}

	// 遍历Next
	*lastId = msgD.Id
	if continuous > 0 {
		client.lastId = *lastId
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

func (that *MsgSess) inLastLoop(client *MsgClient) int64 {
	lastLoop := time.Now().UnixNano()
	that.grp.locker.RLock()
	defer that.grp.locker.RUnlock()
	client.lastLoop = lastLoop
	return lastLoop
}

func (that *MsgSess) outLastLoop(client *MsgClient, unique string, lastLoop int64, lastTime int64) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	if client.lastLoop == lastLoop {
		client.lastLoop = 0
	}

	if lastTime < client.lastTime {
		if client.lastId > 1 {
			// 优化，避免lasting->Lasts 挑来挑去
			that.Lasts(client.lastId, client, unique, client.continuous)

		} else {
			// lasting通知
			that.LastStart(client, unique)
		}
	}
}

func (that *MsgSess) conLastLoop(client *MsgClient, lastTime int64) bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	return lastTime < client.lastTime
}

func (that *MsgSess) lastLoad() {
	if !MsgMng.LastLoad || MsgMng.Db == nil || that.lastLoaded {
		return
	}

	if that.lastLoaded {
		return
	}

	lastQueue := that.getLastQueue()
	if lastQueue != nil && lastQueue.IsEmpty() {
		// 加强消息不丢失，读写锁
		laster := that.grp.laster
		if laster != nil {
			laster.Lock()
			defer laster.Unlock()
		}

		if that.lastLoaded {
			return
		}

		msgDs := MsgMng.Db.Last(that.grp.gid, MsgMng.LastMax)
		// 锁放在io之后
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.lastLoaded {
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

	that.lastLoaded = true
}

// return bool lastIn, 为true 则内存缓存已覆盖lastId，不需要从db读取列表
func (that *MsgSess) lastGet(client *MsgClient, lastLoop int64, lastId int64) (Msg, bool) {
	// 预加载
	that.lastLoad()
	// lock锁提前初始化
	that.getLastQueue()
	// 锁查找
	that.grp.locker.RLock()
	defer that.grp.locker.RUnlock()
	if client.lastLoop != lastLoop {
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

func (that *MsgClient) getCidReq() *gw.CidReq {
	if that.cidReq == nil {
		that.cidReq = &gw.CidReq{
			Cid: that.cid,
		}
	}

	return that.cidReq
}
