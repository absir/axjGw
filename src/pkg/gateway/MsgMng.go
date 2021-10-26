package gateway

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
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
	db        MsgDb
	idWorkder *Util.IdWorker
	locker    sync.Locker
	grpMap    *sync.Map
}

// 初始变量
var MsgMng *msgMng

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
		that.db = msgGorm
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
	passTime int64
	sess     *MsgSess
}

type MsgSess struct {
	grp        *MsgGrp
	queue      *Util.CircleQueue
	queuing    int64
	lastQueue  *Util.CircleQueue
	lastLoaded bool
	client     *MsgClient
	clientMap  *sync.Map
	clientNum  int
}

type MsgClient struct {
	cid      int64
	gatewayI gw.GatewayI
	idleTime int64
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

func (that msgMng) newMsgGrp(gid string) Msg {

}

func (that MsgGrp) retain() {
	that.passTime = time.Now().UnixNano() + MsgMng.LiveDrt
}

func (that *MsgGrp) newMsgSess() *MsgSess {
	sess := new(MsgSess)
	sess.grp = that
	if MsgMng.QueueMax > 0 {
		sess.queue = Util.NewCircleQueue(MsgMng.QueueMax)
	}

	if MsgMng.LastMax > 0 {
		sess.lastQueue = Util.NewCircleQueue(MsgMng.LastMax)
	}

	return sess
}

func (that MsgGrp) newMsgClient(cid int64) *MsgClient {
	client := new(MsgClient)
	client.cid = cid
	client.gatewayI = Server.GetProdCid(cid).GetGWIClient()
	client.idleTime = time.Now().UnixNano() + that.passTime
	return client
}

func (that MsgGrp) getSess() *MsgSess {
	if that.sess == nil {
		that.sess = that.newMsgSess()
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
	old.gatewayI.Kick(Server.Context, old.cid, nil)
	return true
}

func (that MsgGrp) checkClients() {
	sess := that.sess
	if sess == nil {
		return
	}

	that.locker.Lock()
	defer that.locker.Unlock()
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
		that.closeOld(client, client.cid, unique)
	}
}

func (that MsgGrp) Conn(cid int64, unique string) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	client := that.getClient(unique)
	if client != nil {
		if client.cid == cid {
			return true

		} else if client.cid > cid {
			return false
		}

		that.closeOld(client, client.cid, unique)
	}

	sess := that.getSess()
	client = that.newMsgClient(cid)
	if unique == "" {
		sess.client = client

	} else {
		if sess.clientMap == nil {
			sess.clientMap = new(sync.Map)
		}

		sess.clientMap.Store(unique, client)
	}

	sess.clientNum++
	return true
}

func (that MsgGrp) Close(cid int64, unique string) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
	client := that.getClient(unique)
	if client != nil {
		return that.closeOld(client, cid, unique)
	}

	return false
}

type EPush int

const (
	EP_DIRECT EPush = 0
	EP_QUEUE  EPush = 1
	EP_LAST   EPush = 2
)

func (that MsgSess) OnResult(ret gw.Result_, err error, push EPush, client *MsgClient, unique string) bool {
	if ret == gw.Result__Succ {
		return true
	}

	// 消息发送失败
	return false
}

func (that MsgSess) queuingGet(queuing int64) Msg {
	that.grp.locker.RLocker()
	defer that.grp.locker.RUnlock()
	if that.queuing != queuing {
		return nil
	}

	msg, _ := that.queue.Get(0)
	if msg == nil {
		return nil
	}

	return msg.(Msg)
}

func (that MsgSess) queuingRemove(queuing int64, msg Msg) {
	that.grp.locker.Lock()
	defer that.grp.locker.Unlock()
	that.queue.Remove(msg)
}

func (that MsgSess) queuingStart() {
	if that.queuing == 0 || that.queuing == 1 {
		go that.queuingRun(time.Now().UnixNano())
	}
}

func (that MsgSess) queuingEnd(queuing int64) {
	if that.queuing == queuing {
		that.queuing = 0
	}
}

func (that MsgSess) queuingRun(queuing int64) {
	client := that.client
	if client == nil {
		return
	}

	that.queuing = queuing
	defer that.queuingEnd(queuing)
	for {
		if that.queuing != queuing {
			break
		}

		msg := that.queuingGet(queuing)
		if msg == nil {
			break
		}

		msgD := msg.Get()
		ret, err := client.gatewayI.Push(Server.Context, client.cid, msgD.Uri, msgD.Data, msg.Isolate())
		if !that.OnResult(ret, err, EP_QUEUE, client, "") {
			break
		}

		that.queuingRemove(queuing, msg)
	}
}

func (that MsgGrp) addMsg(msg Msg) {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.sess == nil {
		that.sess = that.newMsgSess()
	}

	unique := msg.Unique()
	if unique != "" {
		for i := that.sess.Size() - 1; i >= 0; i-- {
			g, _ := that.sess.Get(i)
			if g != nil && g.(gateway.MsgD).Unique() == unique {
				that.sess.Set(i, nil)
				break
			}
		}
	}

	that.sess.Push(msgG, true)
}
