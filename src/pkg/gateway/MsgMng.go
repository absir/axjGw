package gateway

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Thrd/Util"
	"context"
	"errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gw"
	"sync"
	"time"
)

var NOWAY = errors.New("NOWAY")

type msgMng struct {
	QueueMax  int
	NextLimit int
	LastLimit int
	LastMax   int
	LastLoad  bool
	LastLoop  int
	LastUrl   string
	Context   context.Context
	Last      MsgLast
	IdWorker  *Util.IdWorker
	Locker    sync.Locker
}

var MsgMng *msgMng

func init() {
	MsgMng = &msgMng{
		QueueMax:  20,
		NextLimit: 10,
		LastMax:   20,
		LastLoad:  false,
		LastLoop:  10,
		LastUrl:   "",
	}

	APro.SubCfgBind("msgMng", MsgMng)
	MsgMng.Context = context.Background()
	MsgMng.Last = nil
	if MsgMng.LastUrl != "" {
		db, err := gorm.Open(mysql.Open(MsgMng.LastUrl), &gorm.Config{
		})
		if err != nil {
			panic(err)
		}

		MsgMng.Last = MsgLastDb{
			db: db,
		}
	}

	idWorker, err := Util.NewIdWorker(APro.WorkId())
	Kt.Panic(err)
	MsgMng.IdWorker = idWorker
	MsgMng.Locker = new(sync.Mutex)
}

type MsgU struct {
	Msg
	unique  string
	isolate bool
}

func (m MsgU) Unique() string {
	return m.unique
}

func (m MsgU) Isolate() bool {
	return m.isolate
}

type MsgUser struct {
	sid        string
	locker     *sync.RWMutex
	queue      *Util.CircleQueue
	lastQueue  *Util.CircleQueue
	lastLoaded bool
	conn       *MsgConn
	connMap    map[string]*MsgConn
	connNum    int
	queuing    int64
}

func NewMsgUser(sid string) *MsgUser {
	msgUser := new(MsgUser)
	msgUser.locker = new(sync.RWMutex)
	if MsgMng.QueueMax > 0 {
		msgUser.queue = Util.NewCircleQueue(MsgMng.QueueMax)

	} else {
		msgUser.queue = nil
	}

	if MsgMng.LastMax > 0 {
		msgUser.lastQueue = Util.NewCircleQueue(MsgMng.LastMax)

	} else {
		msgUser.lastQueue = nil
	}

	msgUser.lastLoaded = false
	msgUser.conn = nil
	msgUser.connMap = nil
	msgUser.connNum = 0
	msgUser.queuing = 0
	return msgUser
}

func (that MsgUser) Conn(cid int64, unique string, kick []byte) {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.close(false, cid, unique, true, kick) {
		return
	}

	conn := new(MsgConn)
	conn.Init(cid)
	if unique == "" {
		that.conn = conn

	} else {
		if that.connMap == nil {
			that.connMap = map[string]*MsgConn{}
		}

		that.connMap[unique] = conn
	}

	that.connNum++
}

func (that MsgUser) Close(cid int64, unique string, close bool, kick []byte) bool {
	return that.close(true, cid, unique, close, kick)
}

func (that MsgUser) close(locker bool, cid int64, unique string, close bool, kick []byte) bool {
	if locker {
		that.locker.Lock()
		defer that.locker.Unlock()
	}

	var conn *MsgConn = nil
	if unique == "" {
		conn = that.conn
		if conn == nil {
			return true

		} else if cid > 0 && cid <= conn.cid {
			return false
		}

		that.conn = nil

	} else {
		if that.connMap == nil {
			return true
		}

		conn = that.connMap[unique]
		if conn == nil {
			return true

		} else if cid > 0 && cid <= conn.cid {
			return false
		}

		delete(that.connMap, unique)
	}

	// Kick onyway
	conn.prod.GetGWIClient().KickO(MsgMng.Context, conn.cid)
	that.connNum--
	return true
}

func (that MsgUser) Init(sid string) {
	that.sid = sid
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.queue != nil {
		that.queue.Clear()
	}

	if that.lastQueue != nil {
		that.lastQueue.Clear()
	}

	that.lastLoaded = false
	that.conn = nil
	that.connMap = nil
	that.connNum = 0
	that.queuing = 0
}

func (that MsgUser) Release() {
	that.Init("")
}

func (that MsgUser) lastLoad() {
	if that.lastLoaded && MsgMng.Last == nil && !MsgMng.LastLoad {
		return
	}

	that.locker.Lock()
	defer that.locker.Unlock()
	if that.lastQueue.IsEmpty() {
		msgs := MsgMng.Last.Last(that.sid, MsgMng.LastMax)
		mLen := len(msgs)
		for i := 0; i < mLen; i++ {
			that.lastQueue.Push(msgs[i], true)
		}
	}

	that.lastLoaded = true
}

func (that MsgUser) Clear() {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.queue != nil {
		that.queue.Clear()
	}
}

func (that MsgUser) Push(uri string, data []byte, last int, unique string, isolate bool) (bool, error) {
	if last > 0 {
		if that.lastQueue == nil {
			last = 0

		} else if MsgMng.Last == nil {
			last = 1
		}
	}

	if last > 0 && unique != "" {
		unique = ""
	}

	var msgG MsgG = nil
	if unique == "" && !isolate {
		msgG = new(Msg)

	} else {
		msgU := new(MsgU)
		msgU.unique = unique
		msgU.isolate = isolate
		msgG = msgU
	}

	msg := msgG.Get()
	msg.Sid = that.sid
	msg.Uri = uri
	msg.Data = data
	if last > 0 {
		// 添加到队列，持久化消息
		that.lastMsgG(msgG)
		if last > 1 {
			MsgMng.Last.Insert(*msgG.Get())
		}

		// 消息更新通知
		that.lastStart()

	} else {
		if that.queue == nil {
			if that.conn != nil {
				ret, err := that.conn.Prop().GetGWIClient().Push(MsgMng.Context, that.conn.cid, msg.Uri, msg.Data, msgG.Isolate())
				return that.conn.OnResult(ret, err, EP_DIRECT, that, ""), err
			}

			return false, NOWAY
		}

		// 添加到队列，触发队列发送
		that.addMsgG(msgG)
		that.queuingStart()
	}

	return true, nil
}

func (that MsgUser) lastMsgG(msgG MsgG) {
	that.locker.Lock()
	defer that.locker.Unlock()
	msgG.Get().Id = MsgMng.IdWorker.Generate()
	// 预加载
	that.lastLoad()
	that.lastQueue.Push(msgG, true)
}

func (that MsgUser) addMsgG(msgG MsgG) {
	that.locker.Lock()
	defer that.locker.Unlock()
	unique := msgG.Unique()
	if unique != "" {
		for i := that.queue.Size() - 1; i >= 0; i-- {
			g, _ := that.queue.Get(i)
			if g != nil && g.(MsgG).Unique() == unique {
				that.queue.Set(i, nil)
				break
			}
		}
	}

	that.queue.Push(msgG, true)
}

func (that MsgUser) queuingStart() {
	if that.queuing == 0 || that.queuing == 1 {
		go that.queuingRun(time.Now().UnixNano())
	}
}

func (that MsgUser) queuingEnd(queuing int64) {
	if that.queuing == queuing {
		that.queuing = 0
	}
}

func (that *MsgUser) queuingRun(queuing int64) {
	that.queuing = queuing
	defer that.queuingEnd(queuing)
	for {
		msgG := that.queuingGet(queuing)
		if msgG == nil {
			break
		}

		msg := msgG.(MsgG).Get()
		ret, err := that.conn.Prop().GetGWIClient().Push(MsgMng.Context, that.conn.cid, msg.Uri, msg.Data, msgG.Isolate())
		if !that.conn.OnResult(ret, err, EP_QUEUE, that, "") {
			break
		}

		that.queuingRemove(queuing, msgG)
	}
}

func (that MsgUser) queuingGet(queuing int64) MsgG {
	that.locker.RLocker()
	defer that.locker.RUnlock()
	if that.queuing != queuing {
		return nil
	}

	msgG, _ := that.queue.Get(0)
	if msgG == nil {
		return nil
	}

	return msgG.(MsgG)
}

func (that MsgUser) queuingRemove(queuing int64, msgG MsgG) {
	that.locker.Lock()
	defer that.locker.Unlock()
	that.queue.Remove(msgG)
}

func (that MsgUser) lastStart() {
	that.locker.Lock()
	defer that.locker.Unlock()
	if that.conn != nil {
		that.conn.lastStart(that, "")
	}

	if that.connMap != nil {
		for unique, conn := range that.connMap {
			conn.lastStart(that, unique)
		}
	}
}

func (that MsgUser) idleCheck() {
	if that.queuing == 1 {
		that.queuingStart()
	}

	that.lastStart()
}

type MsgConn struct {
	cid      int64
	prod     *Prod
	lasting  int64
	lastTime int64
}

func (that MsgConn) Init(cid int64) {
	that.cid = cid
	that.prod = nil
	that.lasting = 0
	that.lastTime = 0
}

func (that MsgConn) Prop() *Prod {
	return that.prod
}

type EPush int

const (
	EP_DIRECT EPush = 0
	EP_QUEUE  EPush = 1
	EP_LAST   EPush = 2
)

func (that MsgConn) OnResult(ret gw.Result_, err error, push EPush, user *MsgUser, unique string) bool {
	if ret == gw.Result__Succuess {
		return true
	}

	// 消息发送失败
	return false
}

func (that MsgConn) lastStart(user *MsgUser, unique string) {
	if that.lasting >= 0 {
		if that.lasting <= 1 {
			go that.lastRun(time.Now().UnixNano(), user, unique)

		} else {
			that.lasting = time.Now().UnixNano()
		}
	}
}

func (that MsgConn) lastEnd(lasting int64, user *MsgUser, unique string) {
	if user != nil {
		user.locker.Lock()
		defer user.locker.Unlock()
	}

	if that.lasting == lasting {
		that.lasting = 0
	}
}

func (that MsgConn) lastRun(lasting int64, user *MsgUser, unique string) {
	that.lasting = lasting
	defer that.lastEnd(lasting, user, unique)
	for {
		ret, err := that.Prop().GetGWIClient().Last(MsgMng.Context, that.cid)
		that.OnResult(ret, err, EP_LAST, user, unique)
		if that.lastDone(lasting, user, unique) {
			break
		}
	}
}

func (that MsgConn) lastDone(lasting int64, user *MsgUser, unique string) bool {
	if user != nil {
		user.locker.Lock()
		defer user.locker.Unlock()
	}

	return that.lasting == lasting || that.lasting == 1
}

func (that MsgConn) lastLoop(lastId int64, user *MsgUser, unique string) {
	lastTime := time.Now().UnixNano()
	that.lastTime = lastTime
	for i := 0; i < MsgMng.LastLoop; i++ {
		msgG, lastIn := that.lastGet(lastId, user, unique)
		if msgG == nil {
			if lastIn {
				// 消息已读取完毕
				return

			} else {
				// 缓冲消息
				msgs := MsgMng.Last.Next(user.sid, lastId, MsgMng.NextLimit)
				mLen := len(msgs)
				if mLen <= 0 {
					return
				}

				for j := 0; j < mLen; j++ {
					msg := msgs[j]
					if lastTime != that.lastTime {
						return
					}

					ret, err := that.Prop().GetGWIClient().Push(MsgMng.Context, that.cid, msg.Uri, msg.Data, false)
					if !that.OnResult(ret, err, EP_DIRECT, user, unique) {
						return
					}
				}

				break
			}

		} else {
			msg := msgG.Get()
			if lastTime != that.lastTime {
				return
			}

			ret, err := that.Prop().GetGWIClient().Push(MsgMng.Context, that.cid, msg.Uri, msg.Data, msgG.Isolate())
			if !that.OnResult(ret, err, EP_DIRECT, user, unique) {
				return
			}

			// 遍历Next
			lastId = msg.Id
		}
	}

	// 下一轮消息通知
	that.lastStart(user, unique)
}

func (that MsgConn) lastGet(lastId int64, user *MsgUser, unique string) (MsgG, bool) {
	user.locker.RLocker()
	defer user.locker.RUnlock()
	// 预加载
	user.lastLoad()
	size := user.lastQueue.Size()
	i := 0
	lastIn := false
	for ; i < size; i++ {
		msgG, _ := user.lastQueue.Get(i)
		msg := msgG.(MsgG).Get()
		if msg.Id > lastId {
			return msgG.(MsgG), lastIn

		} else {
			lastIn = true
		}
	}

	return nil, lastIn
}
