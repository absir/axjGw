package gateway

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gw"
	"sync"
	"time"
)

type msgMng struct {
	QueueMax  int
	NextLimit int
	LastLimit int
	LastMax   int
	LastLoad  bool
	LastLoop  int
	LastUrl   string
	CheckDrt  time.Duration
	LiveDrt   int64
	IdleDrt   int64
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
		LastMax:   20,
		LastLoad:  false,
		LastLoop:  10,
		LastUrl:   "",
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

	// 消息持久化
	if that.LastUrl != "" {
		db, err := gorm.Open(mysql.Open(that.LastUrl), &gorm.Config{})
		Kt.Panic(err)

		msgGorm := MsgGorm{
			db: db,
		}
		// 自动创建表
		msgGorm.AutoMigrate()
		that.Db = msgGorm
	}

	// 属性初始化
	that.idWorkder = Util.NewIdWorkerPanic(Config.WorkId)
	that.locker = new(sync.Mutex)
	that.grpMap = new(sync.Map)
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
	cid      int64
	gatewayI gw.GatewayI
	connVer  int32
	idleTime int64
	lasting  int64
	lastLoop int64
}

// 空闲检测
func (that msgMng) CheckStop() {
	that.checkLoop = -1
}

func (that msgMng) CheckLoop() {
	loopTime := time.Now().UnixNano()
	that.checkLoop = loopTime
	for loopTime == that.checkLoop {
		time.Sleep(that.CheckDrt)
		that.checkTime = time.Now().UnixNano()
		that.grpMap.Range(that.checkRange)
	}
}

func (that msgMng) checkRange(key interface{}, val interface{}) bool {
	that.checkGrp(key, val.(*MsgGrp))
	return true
}

func (that msgMng) checkGrp(key interface{}, grp *MsgGrp) {
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

func (that msgMng) MsgGrp(gid string) *MsgGrp {
	val, _ := that.grpMap.Load(gid)
	return val.(*MsgGrp)
}

func (that msgMng) GetMsgGrp(gid string) *MsgGrp {
	that.locker.Lock()
	defer that.locker.Unlock()
	val, _ := that.grpMap.Load(gid)
	grp := val.(*MsgGrp)
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

func (that msgMng) newMsgGrp(gid string) *MsgGrp {
	grp := new(MsgGrp)
	grp.gid = gid
	grp.ghash = Kt.HashCode(KtUnsafe.StringToBytes(gid))
	grp.locker = new(sync.RWMutex)
	grp.retain()
	return grp
}

func (that msgMng) newConnVer() int32 {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.connVer++
	if that.connVer > connVerMax {
		that.connVer = 1
	}

	return that.connVer
}

func (that MsgGrp) Sess() *MsgSess {
	return that.sess
}

func (that MsgGrp) retain() {
	that.passTime = time.Now().UnixNano() + MsgMng.LiveDrt
}

func (that *MsgGrp) newMsgSess() *MsgSess {
	sess := new(MsgSess)
	sess.grp = that
	return sess
}

func (that MsgGrp) newMsgClient(cid int64) *MsgClient {
	client := new(MsgClient)
	client.cid = cid
	client.gatewayI = Server.GetProdCid(cid).GetGWIClient()
	client.idleTime = time.Now().UnixNano() + that.passTime
	return client
}

func (that MsgGrp) getLaster() sync.Locker {
	if that.laster == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.laster == nil {
			that.laster = new(sync.Mutex)
		}
	}

	return that.laster
}

func (that MsgGrp) getSess(create bool) *MsgSess {
	if that.sess == nil && create {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.sess == nil {
			that.sess = that.newMsgSess()
		}
	}

	return that.sess
}

func (that MsgGrp) getClient(unique string) *MsgClient {
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
	return client.(*MsgClient)
}

func (that MsgGrp) closeOld(old *MsgClient, cid int64, unique string) bool {
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

	sess.clientNum--
	// 关闭通知
	go old.gatewayI.Kick(Server.Context, old.cid, nil)
	return true
}

func (that MsgGrp) checkClients() {
	sess := that.sess
	if sess == nil {
		return
	}

	that.checkClient(sess.client, "")
	if sess.clientMap != nil {
		sess.clientMap.Range(that.checkRange)
	}
}

func (that MsgGrp) checkRange(key, val interface{}) bool {
	that.checkClient(val.(*MsgClient), key.(string))
	return true
}

func (that MsgGrp) checkClient(client *MsgClient, unique string) {
	if client == nil {
		return
	}

	if client.idleTime > MsgMng.checkTime {
		// 未空闲
		return
	}

	result, _ := client.gatewayI.Alive(Server.Context, client.cid)
	if result != gw.Result__Succ {
		that.locker.Lock()
		defer that.locker.Unlock()
		that.closeOld(client, client.cid, unique)
	}
}

func (that MsgGrp) Conn(cid int64, unique string) *MsgClient {
	that.locker.Lock()
	defer that.locker.Unlock()
	client := that.getClient(unique)
	if client != nil {
		if client.cid == cid {
			client.connVer = MsgMng.newConnVer()
			return client

		} else if client.cid > cid {
			return nil
		}

		that.closeOld(client, client.cid, unique)
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

func (that MsgGrp) Close(cid int64, unique string, connVer int32) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	client := that.getClient(unique)
	if client != nil && (connVer == 0 || connVer == client.connVer) {
		return that.closeOld(client, cid, unique)
	}

	return false
}

func (that MsgGrp) Clear(queue bool, last bool) {
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

func (that MsgGrp) Push(uri string, bytes []byte, isolate bool, qs int32, queue bool, unique string, fid int64) (Msg, bool, error) {
	if qs >= 2 {
		if MsgMng.LastMax <= 0 {
			qs = 1
		}
	}

	sess := that.getSess(queue)
	if qs <= 1 {
		if sess == nil {
			return nil, false, NOWAY
		}

		msg := NewMsg(uri, bytes, unique)
		succ, err := sess.QueuePush(msg)
		return msg, succ, err

	} else {
		var err error = nil
		msg := NewMsg(uri, bytes, unique)
		that.lastPush(sess, msg, fid)
		msgD := msg.Get()
		if qs >= 3 && MsgMng.Db != nil {
			err = that.insertMsgD(msgD)
		}

		if sess != nil && msg.Get().Id > 0 {
			sess.LastsStart()
		}

		return msg, true, err
	}
}

func (that MsgGrp) insertMsgD(msgD *MsgD) error {
	if msgD.Id > 0 {
		// sess.lastQueue加强不漏消息
		return MsgMng.Db.Insert(*msgD)

	} else {
		// laster加强消息顺序写入
		that.getLaster().Lock()
		defer that.getLaster().Unlock()
		msgD.Id = MsgMng.idWorkder.Generate()
		return MsgMng.Db.Insert(*msgD)
	}
}

// 插入顺序消息
func (that MsgGrp) lastPush(sess *MsgSess, msg Msg, fid int64) {
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

var NOWAY = errors.New("NOWAY")

func (that MsgSess) OnResult(ret gw.Result_, err error, rpc ERpc, client *MsgClient, unique string) bool {
	if ret == gw.Result__Succ {
		client.idleTime = time.Now().UnixNano() + MsgMng.IdleDrt
		return true
	}

	// 消息发送失败
	return false
}

func (that MsgSess) getQueue() *Util.CircleQueue {
	if that.queue == nil && MsgMng.QueueMax > 0 {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.queue == nil {
			that.queue = Util.NewCircleQueue(MsgMng.QueueMax)
		}
	}

	return that.queue
}

func (that MsgSess) getLastQueue() *Util.CircleQueue {
	if that.lastQueue == nil && MsgMng.LastMax > 0 {
		that.grp.locker.Lock()
		defer that.grp.locker.Unlock()
		if that.lastQueue == nil {
			that.lastQueue = Util.NewCircleQueue(MsgMng.LastMax)
		}
	}

	return that.lastQueue
}

func (that MsgSess) QueuePush(msg Msg) (bool, error) {
	if msg == nil {
		return false, nil
	}

	if that.getQueue() == nil {
		client := that.client
		if client != nil {
			msgD := msg.Get()
			ret, err := that.client.gatewayI.Push(Server.Context, client.cid, msgD.Uri, msgD.Data, true, 0)
			return that.OnResult(ret, err, ER_PUSH, client, ""), err
		}

		return false, NOWAY
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

func (that MsgSess) QueueStart() {
	if that.client == nil || that.queuing {
		return
	}

	go that.queueRun()
}

func (that MsgSess) queueIn() bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	if that.queuing {
		return false
	}

	that.queuing = true
	return true
}

func (that MsgSess) queueOut() {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	that.queuing = false
}

func (that MsgSess) queueRun() {
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

		msgD := msg.Get()
		ret, err := client.gatewayI.Push(Server.Context, client.cid, msgD.Uri, msgD.Data, true, 0)
		if !that.OnResult(ret, err, ER_PUSH, client, "") {
			break
		}

		that.queueRemove(msg)
	}
}

func (that MsgSess) queueGet() Msg {
	that.grp.locker.RLocker()
	defer that.grp.locker.RUnlock()
	msg, _ := that.queue.Get(0)
	if msg == nil {
		return nil
	}

	return msg.(Msg)
}

func (that MsgSess) queueRemove(msg Msg) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	that.queue.Remove(msg)
}

func (that MsgSess) LastsStart() {
	that.LastStart(that.client, "")
	if that.clientMap != nil {
		that.clientMap.Range(that.LastsStartRange)
	}
}

func (that MsgSess) LastsStartRange(key, val interface{}) bool {
	that.LastStart(val.(*MsgClient), key.(string))
	return true
}

func (that MsgSess) LastStart(client *MsgClient, unique string) {
	if client == nil {
		return
	}

	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	if client.lasting == 0 {
		go that.lastRun(client, unique)
		return
	}

	client.lasting = time.Now().UnixNano()
}

func (that MsgSess) lastIn(client *MsgClient) bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	if client.lasting != 0 {
		return false
	}

	client.lasting = time.Now().UnixNano()
	return true
}

func (that MsgSess) lastOut(client *MsgClient) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	client.lasting = 0
}

func (that MsgSess) lastDone(client *MsgClient, lasting int64) bool {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	return client.lasting <= lasting
}

func (that MsgSess) lastRun(client *MsgClient, unique string) {
	if !that.lastIn(client) {
		return
	}

	defer that.lastOut(client)

	for {
		lasting := client.lasting
		ret, err := client.gatewayI.Last(Server.Context, client.cid, that.grp.gid, client.connVer)
		that.OnResult(ret, err, ER_LAST, client, unique)
		if that.lastDone(client, lasting) {
			break
		}
	}
}

func (that MsgSess) Lasts(lastId int64, client *MsgClient, unique string) {
	if client == nil {
		return
	}

	go that.lastLoop(lastId, client, unique)
}

func (that MsgSess) lastLoop(lastId int64, client *MsgClient, unique string) {
	if client == nil {
		return
	}

	lastLoop := that.inLastLoop(client)
	for i := 0; i < MsgMng.LastLoop; i++ {
		msg, lastIn := that.lastGet(client, lastLoop, lastId)
		if msg == nil {
			if lastIn {
				// 消息已读取完毕
				return

			} else {
				// 缓冲消息
				msgDs := MsgMng.Db.Next(that.grp.gid, lastId, MsgMng.NextLimit)
				mLen := len(msgDs)
				if mLen <= 0 {
					return
				}

				var msgD MsgD
				for j := 0; j < mLen; j++ {
					msgD = msgDs[j]
					if lastLoop != client.lastLoop {
						return
					}

					ret, err := client.gatewayI.Push(Server.Context, client.cid, msgD.Uri, msgD.Data, false, msgD.Id)
					if !that.OnResult(ret, err, ER_PUSH, client, unique) {
						return
					}
				}

				// 遍历Next
				lastId = msgD.Id
				break
			}

		} else {
			msgD := msg.Get()
			if lastLoop != client.lastLoop {
				return
			}

			ret, err := client.gatewayI.Push(Server.Context, client.cid, msgD.Uri, msgD.Data, true, msgD.Id)
			if !that.OnResult(ret, err, ER_PUSH, client, unique) {
				return
			}

			// 遍历Next
			lastId = msgD.Id
		}
	}
}

func (that MsgSess) inLastLoop(client *MsgClient) int64 {
	lastLoop := time.Now().UnixNano()
	that.grp.locker.RLocker()
	defer that.grp.locker.RUnlock()
	client.lastLoop = lastLoop
	return lastLoop
}

func (that MsgSess) lastLoad() {
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
func (that MsgSess) lastGet(client *MsgClient, lastLoop int64, lastId int64) (Msg, bool) {
	// 预加载
	that.lastLoad()
	// 锁查找
	that.grp.locker.RLocker()
	defer that.grp.locker.RUnlock()
	if client.lastLoop != lastLoop {
		return nil, true
	}

	size := that.getLastQueue().Size()
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

func (that msgMng) dbMsgFail(id int64, gid string) {
	Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, gid,)
}
