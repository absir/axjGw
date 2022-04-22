package gateway

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

type msgMng struct {
	QueueMax      int           // 主client消息队列大小
	NextLimit     int           // last消息，单次读取列表数
	LastLimit     int           // last消息队类，初始化载入列表数
	LastMax       int           // last消息队列大小
	LastMaxAll    int           // 所有消息队列对打
	LastLoad      bool          // 是否执行 last消息队类，初始化载入列表数
	LastUri       string        // 消息持久化，数据库连接
	ClearCron     string        // 消息清理，执行周期
	ClearDay      int64         // 清理消息过期天数
	CheckDrt      time.Duration // 执行检查逻辑，间隔
	LiveDrt       int64         // 连接断开，存活时间
	IdleDrt       int64         // 连接检查，间隔
	ORwLocker     bool          // 独立Queue读写锁
	OClientLocker bool          // 独立客户端锁
	pushDrTest    bool          // 消息直写测试
	LastRangeWait bool          // Last通知RangeWait遍历
	checkLoop     int64
	checkTime     int64
	checkBuff     []interface{}
	Db            MsgDb
	idWorkder     *Util.IdWorker
	locker        sync.Locker
	connVer       int32
	grpMap        *cmap.CMap
	readVer       int32
}

// 初始变量
var _msgMng *msgMng

const connVerMax = KtBytes.VINT_3_MAX - 1

func MsgMng() *msgMng {
	if _msgMng == nil {
		Server.Locker.Lock()
		defer Server.Locker.Unlock()
		if _msgMng == nil {
			initMsgMng()
			go _msgMng.CheckLoop()
		}
	}

	return _msgMng
}

func initMsgMng() {
	// 消息管理配置
	that := &msgMng{
		QueueMax:      32,
		NextLimit:     100,
		LastMax:       33, // over load cover msgs [QueueMax]
		LastLoad:      true,
		LastUri:       "",
		ClearCron:     "0 0 3 * * *",
		ClearDay:      30,
		CheckDrt:      5,
		LiveDrt:       15,
		IdleDrt:       30,
		ORwLocker:     true,
		OClientLocker: true,
		pushDrTest:    false,
		LastRangeWait: false,
	}

	// 配置处理
	APro.SubCfgBind("msg", that)
	that.LastLoad = that.LastLoad && that.LastMax > 0

	// 最长消息队列
	that.LastMaxAll = that.LastMax

	// 属性初始化
	that.idWorkder = Util.NewIdWorkerPanic(Config.WorkId)
	that.locker = new(sync.Mutex)
	that.grpMap = cmap.NewCMapInit()

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
			Server.Cron(false).AddFunc(that.ClearCron, that.ClearPass)
			that.ClearPass()
		}
	}

	_msgMng = that
}

// 清理过期消息
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
	checkLoop := time.Now().UnixNano()
	that.checkLoop = checkLoop
	checkDrt := that.CheckDrt * time.Second
	for Kt.Active && checkLoop == that.checkLoop {
		time.Sleep(checkDrt)
		that.checkTime = time.Now().Unix()
		that.grpMap.RangeBuff(that.checkRange, &that.checkBuff, 1024)
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
		if sess != nil {
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
		grp.locker.Lock()
		if grp.passTime > time {
			grp.locker.Unlock()
			that.locker.Unlock()
			return
		}

		that.grpMap.Delete(key)
		grp.locker.Unlock()
		that.locker.Unlock()
		AZap.Debug("Grp Pass Del %s", grp.gid)
	}
}

// 获取创建管理组
func (that *msgMng) GetMsgGrp(gid string) *MsgGrp {
	val, _ := that.grpMap.Load(gid)
	msgGrp, _ := val.(*MsgGrp)
	return msgGrp
}

// 获取或创建管理组
func (that *msgMng) GetOrNewMsgGrp(gid string) *MsgGrp {
	that.locker.Lock()
	val, _ := that.grpMap.Load(gid)
	grp, _ := val.(*MsgGrp)
	if grp != nil {
		if Server.IsProdHash(grp.ghash) {
			grp.retain()
		}

	} else {
		grp = new(MsgGrp)
		grp.gid = gid
		grp.ghash = Kt.HashCode(KtUnsafe.StringToBytes(gid))
		grp.rwLocker = new(sync.RWMutex)
		if _msgMng.ORwLocker {
			grp.locker = new(sync.Mutex)

		} else {
			grp.locker = grp.rwLocker
		}

		grp.retain()
		that.grpMap.Store(gid, grp)
	}

	that.locker.Unlock()
	return grp
}

// 新连接版本
func (that *msgMng) newConnVer() int32 {
	that.locker.Lock()
	connVer := that.connVer
	if connVer < R_SUCC_MIN || connVer >= connVerMax {
		connVer = R_SUCC_MIN

	} else {
		connVer++
	}

	that.connVer = connVer
	that.locker.Unlock()
	return connVer
}

func (that *msgMng) GidForTid(gid string, tid string) string {
	return gid + "/" + tid
}

func (that *msgMng) TidFromGidForTid(gidForTid string) string {
	idx := strings.LastIndexByte(gidForTid, '/')
	if idx > 0 {
		return gidForTid[idx+1:]
	}

	return gidForTid
}

// 新未读版本
func (that *msgMng) newUnreadVer() int32 {
	that.locker.Lock()
	readVer := that.readVer
	if readVer < R_SUCC_MIN || readVer >= connVerMax {
		readVer = R_SUCC_MIN

	} else {
		readVer++
	}

	that.readVer = readVer
	that.locker.Unlock()
	return readVer
}

func (that *msgMng) UnreadTids(gid string, tids []string) {
	if tids == nil || that.Db == nil {
		return
	}

	for i, tid := range tids {
		tids[i] = that.GidForTid(gid, tid)
	}

	nums := that.Db.UnReads(tids)
	if nums == nil {
		return
	}

	grp := that.GetMsgGrp(gid)
	sess := grp.GetOrNewSess(true)
	for _, num := range nums {
		// 未读消息数设置
		sess.UnreadRecv(that.TidFromGidForTid(num.Gid), num.Num, 0)
	}
}
