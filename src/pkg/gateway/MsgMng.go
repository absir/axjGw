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
	msgUser.queuing = 0
	return msgUser
}

func (m *MsgUser) Init(sid string) {
	m.sid = sid
	m.locker.Lock()
	defer m.locker.Unlock()
	if m.queue != nil {
		m.queue.Clear()
	}

	if m.lastQueue != nil {
		m.lastQueue.Clear()
	}

	m.lastLoaded = false
	m.conn = nil
	m.connMap = nil
	m.queuing = 0
}

func (m *MsgUser) lastLoad() {
	if m.lastLoaded && MsgMng.Last == nil && !MsgMng.LastLoad {
		return
	}

	m.locker.Lock()
	defer m.locker.Unlock()
	if m.lastQueue.IsEmpty() {
		msgs := MsgMng.Last.Last(m.sid, MsgMng.LastMax)
		mLen := len(msgs)
		for i := 0; i < mLen; i++ {
			m.lastQueue.Push(msgs[i], true)
		}
	}

	m.lastLoaded = true
}

func (m *MsgUser) Clear() {
	m.locker.Lock()
	defer m.locker.Unlock()
	if m.queue != nil {
		m.queue.Clear()
	}
}

func (m *MsgUser) Push(uri string, data []byte, last int, unique string, isolate bool) (bool, error) {
	if last > 0 {
		if m.lastQueue == nil {
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
	msg.Sid = m.sid
	msg.Uri = uri
	msg.Data = data
	if last > 0 {
		// 添加到队列，持久化消息
		m.lastMsgG(msgG)
		if last > 1 {
			MsgMng.Last.Insert(*msgG.Get())
		}

		// 消息更新通知
		m.lastStart()

	} else {
		if m.queue == nil {
			if m.conn != nil {
				ret, err := m.conn.Prop().GetGWIClient().Push(MsgMng.Context, m.conn.cid, msg.Uri, msg.Data, msgG.Isolate())
				return m.conn.OnResult(ret, err, EP_DIRECT, m, ""), err
			}

			return false, NOWAY
		}

		// 添加到队列，触发队列发送
		m.addMsgG(msgG)
		m.queuingStart()
	}

	return true, nil
}

func (m *MsgUser) lastMsgG(msgG MsgG) {
	m.locker.Lock()
	defer m.locker.Unlock()
	msgG.Get().Id = MsgMng.IdWorker.Generate()
	// 预加载
	m.lastLoad()
	m.lastQueue.Push(msgG, true)
}

func (m *MsgUser) addMsgG(msgG MsgG) {
	m.locker.Lock()
	defer m.locker.Unlock()
	unique := msgG.Unique()
	if unique != "" {
		for i := m.queue.Size() - 1; i >= 0; i-- {
			g, _ := m.queue.Get(i)
			if g != nil && g.(MsgG).Unique() == unique {
				m.queue.Set(i, nil)
				break
			}
		}
	}

	m.queue.Push(msgG, true)
}

func (m *MsgUser) queuingStart() {
	if m.queuing == 0 || m.queuing == 1 {
		go m.queuingRun(time.Now().UnixNano())
	}
}

func (m *MsgUser) queuingEnd(queuing int64) {
	if m.queuing == queuing {
		m.queuing = 0
	}
}

func (m *MsgUser) queuingRun(queuing int64) {
	m.queuing = queuing
	defer m.queuingEnd(queuing)
	for {
		msgG := m.queuingGet(queuing)
		if msgG == nil {
			break
		}

		msg := msgG.(MsgG).Get()
		ret, err := m.conn.Prop().GetGWIClient().Push(MsgMng.Context, m.conn.cid, msg.Uri, msg.Data, msgG.Isolate())
		if !m.conn.OnResult(ret, err, EP_QUEUE, m, "") {
			break
		}

		m.queuingRemove(queuing, msgG)
	}
}

func (m *MsgUser) queuingGet(queuing int64) MsgG {
	m.locker.RLocker()
	defer m.locker.RUnlock()
	if m.queuing != queuing {
		return nil
	}

	msgG, _ := m.queue.Get(0)
	if msgG == nil {
		return nil
	}

	return msgG.(MsgG)
}

func (m *MsgUser) queuingRemove(queuing int64, msgG MsgG) {
	m.locker.Lock()
	defer m.locker.Unlock()
	m.queue.Remove(msgG)
}

func (m *MsgUser) lastStart() {
	m.locker.Lock()
	defer m.locker.Unlock()
	if m.conn != nil {
		m.conn.lastStart(m, "")
	}

	if m.connMap != nil {
		for unique, conn := range m.connMap {
			conn.lastStart(m, unique)
		}
	}
}

func (m *MsgUser) idleCheck() {
	if m.queuing == 1 {
		m.queuingStart()
	}

	m.lastStart()
}

type MsgConn struct {
	cid      int64
	prod     *Prod
	lasting  int64
	lastTime int64
}

func (m *MsgConn) Init(cid int64) {
	m.cid = cid
	m.prod = nil
	m.lasting = 0
	m.lastTime = 0
}

func (m *MsgConn) Prop() *Prod {
	return m.prod
}

type EPush int

const (
	EP_DIRECT EPush = 0
	EP_QUEUE  EPush = 1
	EP_LAST   EPush = 2
)

func (m *MsgConn) OnResult(ret gw.Result_, err error, push EPush, user *MsgUser, unique string) bool {
	if ret == gw.Result__Succuess {
		return true
	}

	// 消息发送失败
	return false
}

func (m *MsgConn) lastStart(user *MsgUser, unique string) {
	if m.lasting >= 0 {
		if m.lasting <= 1 {
			go m.lastRun(time.Now().UnixNano(), user, unique)

		} else {
			m.lasting = time.Now().UnixNano()
		}
	}
}

func (m *MsgConn) lastEnd(lasting int64, user *MsgUser, unique string) {
	if user != nil {
		user.locker.Lock()
		defer user.locker.Unlock()
	}

	if m.lasting == lasting {
		m.lasting = 0
	}
}

func (m *MsgConn) lastRun(lasting int64, user *MsgUser, unique string) {
	m.lasting = lasting
	defer m.lastEnd(lasting, user, unique)
	for {
		ret, err := m.Prop().GetGWIClient().Last(MsgMng.Context, m.cid)
		m.OnResult(ret, err, EP_LAST, user, unique)
		if m.lastDone(lasting, user, unique) {
			break
		}
	}
}

func (m *MsgConn) lastDone(lasting int64, user *MsgUser, unique string) bool {
	if user != nil {
		user.locker.Lock()
		defer user.locker.Unlock()
	}

	return m.lasting == lasting || m.lasting == 1
}

func (m *MsgConn) lastLoop(lastId int64, user *MsgUser, unique string) {
	lastTime := time.Now().UnixNano()
	m.lastTime = lastTime
	for i := 0; i < MsgMng.LastLoop; i++ {
		msgG, lastIn := m.lastGet(lastId, user, unique)
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
					if lastTime != m.lastTime {
						return
					}

					ret, err := m.Prop().GetGWIClient().Push(MsgMng.Context, m.cid, msg.Uri, msg.Data, false)
					if !m.OnResult(ret, err, EP_DIRECT, user, unique) {
						return
					}
				}

				break
			}

		} else {
			msg := msgG.Get()
			if lastTime != m.lastTime {
				return
			}

			ret, err := m.Prop().GetGWIClient().Push(MsgMng.Context, m.cid, msg.Uri, msg.Data, msgG.Isolate())
			if !m.OnResult(ret, err, EP_DIRECT, user, unique) {
				return
			}

			// 遍历Next
			lastId = msg.Id
		}
	}

	// 下一轮消息通知
	m.lastStart(user, unique)
}

func (m *MsgConn) lastGet(lastId int64, user *MsgUser, unique string) (MsgG, bool) {
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
