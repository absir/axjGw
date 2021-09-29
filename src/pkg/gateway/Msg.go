package gateway

import (
	"gorm.io/gorm"
)

type MsgG interface {
	Get() *Msg
	Unique() string
	Isolate() bool
}

type Msg struct {
	Id   int64  `gorm:"primary_key"`
	Sid  string `gorm:"type:varchar(255);not null;index:Sid"`
	Uri  string `gorm:"type:varchar(255);"`
	Data []byte `gorm:""`
}

func (m *Msg) Get() *Msg {
	return m
}

func (m Msg) Unique() string {
	return ""
}

func (m Msg) Isolate() bool {
	return false
}

type MsgQueue interface {
	Insert(msg Msg) int64
	Next(sid string, id int64, limit int) []Msg
	Last(sid string, limit int) []Msg
}

type MsgQueueDb struct {
	db *gorm.DB
}

func (m MsgQueueDb) Insert(msg Msg) int64 {
	m.db.Create(msg)
	return msg.Id
}

func (m MsgQueueDb) Next(sid string, lastId int64, limit int) []Msg {
	var msgs []Msg = nil
	m.db.Where("Sid = ?", sid).Order("Id").Limit(limit).Find(&msgs)
	return msgs
}

func (m MsgQueueDb) Last(sid string, limit int) []Msg {
	var msgs []Msg = nil
	m.db.Order("Id DESC").Limit(limit).Find(&msgs)
	if msgs != nil {
		// 倒序
		mLen := len(msgs)
		last := mLen - 1
		mLen = mLen / 2
		for i := 0; i < mLen; i++ {
			msg := msgs[i]
			j := last - i
			msgs[i] = msgs[j]
			msgs[j] = msg
		}
	}

	return msgs
}
