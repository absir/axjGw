package gateway

import "sync"

type MsgMng struct {
	mq      MsgQueue
	lastMax int64
}

var Mng *MsgMng = nil

type MsgConn struct {
	cid    int64
	lastId int64
	sync.Pool
}

func (m *MsgConn) lastLoop(lastId int64) {

}
