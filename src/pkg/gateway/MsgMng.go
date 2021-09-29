package gateway

import (
	"axj/APro"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"context"
	"go.uber.org/zap"
	"sync"
)

type msgMng struct {
	QueueMax  int
	NextLimit int
	LastLimit int
	Context   context.Context
	Mq        MsgQueue
	IdWorker  *Util.IdWorker
}

var MsgMng *msgMng

func init() {
	MsgMng = &msgMng{
		QueueMax:  20,
		NextLimit: 20,
		LastLimit: 20,
	}

	APro.SubCfgBind("msgMng", MsgMng)
	MsgMng.Context = context.Background()
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
	sid       string
	locker    *sync.RWMutex
	queue     *Util.CircleQueue
	lastQueue *Util.CircleQueue
	conn      *MsgConn
	connMap   map[string]*MsgConn
}

func NewMsgUser(sid string) *MsgUser {
	msgUser := new(MsgUser)
	msgUser.locker = new(sync.RWMutex)
	if MsgMng.QueueMax > 0 {
		msgUser.queue = Util.NewCircleQueue(MsgMng.QueueMax)

	} else {
		msgUser.queue = nil
	}

	if MsgMng.LastLimit > 0 {
		msgUser.queue = Util.NewCircleQueue(MsgMng.QueueMax)

	} else {
		msgUser.queue = Util.NewCircleQueue(MsgMng.QueueMax)
	}

	return msgUser
}

func (m *MsgUser) Push(uri string, data []byte, qs int, unique string, isolate bool) {
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
	m.queueMsgG(msgG)
	if qs >= 2 && MsgMng.Mq != nil {
		MsgMng.Mq.Insert(*msg)
	}
}

func (m *MsgUser) queueMsgG(msgG MsgG) {
	m.locker.Lock()
	defer m.locker.Unlock()
	msgG.Get().Id = MsgMng.IdWorker.Generate()
	m.queue.Push(msgG, true)
}

type MsgConn struct {
	cid    int64
	prod   *Prod
	lastId int64
}

func (m *MsgConn) lastLoop(sid string, lastId int64) {
	if m.lastId == lastId || MsgMng.Mq == nil || MsgMng.NextLimit <= 0 {
		return
	}

	m.lastId = lastId
	for m.lastId == lastId {
		msgs := MsgMng.Mq.Next(sid, lastId, MsgMng.NextLimit)
		if msgs == nil {
			return
		}

		mLen := len(msgs)
		if mLen <= 0 {
			return
		}

		for i := 0; i < mLen; i++ {
			msg := msgs[i]
			lastId = msg.Id
			if m.prod == nil {
				m.prod = GetProds(Config.GwProd).GetProd(Util.GetWorkerId(m.cid))
			}

			succ, err := m.prod.GetGWIClient().Push(MsgMng.Context, m.cid, msg.Uri, msg.Data)
			if err != nil || !succ {
				AZap.Logger.Warn("msg lastLoop err", zap.Int64("id", msg.Id), zap.Error(err))
				return
			}

			m.lastId = lastId
		}
	}
}
