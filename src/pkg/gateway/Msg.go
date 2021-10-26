package gateway

import (
	"gorm.io/gorm"
)

type Msg interface {
	Get() *MsgD
	Unique() string
	Isolate() bool
}

type MsgD struct {
	Id   int64  `gorm:"primary_key"`                          // 消息编号
	Gid  string `gorm:"type:varchar(255);not null;index:Gid"` // 消息分组
	Fid  int64  `gorm:"index:Fid"`                            // 消息来源编号
	Uri  string `gorm:"type:varchar(255);"`
	Data []byte `gorm:""`
}

func (that *MsgD) Get() *MsgD {
	return that
}

func (that MsgD) Unique() string {
	return ""
}

func (that MsgD) Isolate() bool {
	return false
}

type MsgU struct {
	MsgD
	unique  string
	isolate bool
}

func (m MsgU) Unique() string {
	return m.unique
}

func (m MsgU) Isolate() bool {
	return m.isolate
}

func NewMsg(uri string, data []byte, unique string, isolate bool) Msg {
	if unique == "" && !isolate {
		return &MsgD{
			Uri:  uri,
			Data: data,
		}

	} else {
		msg := &MsgU{
			unique:  unique,
			isolate: isolate,
		}

		msg.Uri = uri
		msg.Data = data
		return msg
	}
}

type MsgDb interface {
	Insert(msg MsgD) int64                           // 插入消息
	Next(gid string, lastId int64, limit int) []MsgD // 遍历消息
	Last(gid string, limit int) []MsgD               // 初始消息缓存
	Delete(fid int64)                                // 来源删除消息
	Clear(oId int64)                                 // 清理过期消息
}

type MsgGorm struct {
	db *gorm.DB
}

func (that MsgGorm) AutoMigrate() {
	migrator := that.db.Migrator()
	if (!migrator.HasTable(&MsgD{})) {
		migrator.AutoMigrate(&MsgD{})
	}
}

func (that MsgGorm) Insert(msg MsgD) int64 {
	that.db.Create(msg)
	return msg.Id
}

func (that MsgGorm) Next(gid string, lastId int64, limit int) []MsgD {
	var msgDS []MsgD = nil
	that.db.Where("Gid = ? AND Id > ?", gid, lastId).Order("Id").Limit(limit).Find(&msgDS)
	return msgDS
}

func (that MsgGorm) Last(gid string, limit int) []MsgD {
	var msgDS []MsgD = nil
	that.db.Where("Gid = ?", gid).Order("Id DESC").Limit(limit).Find(&msgDS)
	if msgDS != nil {
		// 倒序
		mLen := len(msgDS)
		last := mLen - 1
		mLen = mLen / 2
		for i := 0; i < mLen; i++ {
			msg := msgDS[i]
			j := last - i
			msgDS[i] = msgDS[j]
			msgDS[j] = msg
		}
	}

	return msgDS
}

func (that MsgGorm) Delete(fid int64) {
	that.db.Exec("DELETE FROM MsgD WHERE Fid = ?", fid)
}

func (that MsgGorm) Clear(oId int64) {
	that.db.Exec("DELETE FROM MsgD WHERE Id <= ?", oId)
}
