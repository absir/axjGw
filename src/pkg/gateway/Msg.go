package gateway

import (
	"gorm.io/gorm"
	"gw"
)

type Msg interface {
	Get() *MsgD
	Unique() string
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

type MsgU struct {
	MsgD
	unique string
}

func (m MsgU) Unique() string {
	return m.unique
}

func NewMsg(uri string, data []byte, unique string) Msg {
	if unique == "" {
		return &MsgD{
			Uri:  uri,
			Data: data,
		}

	} else {
		msg := &MsgU{
			unique: unique,
		}

		msg.Uri = uri
		msg.Data = data
		return msg
	}
}

type MsgTeam struct {
	Id      int64        `gorm:"primary_key"`                          // 消息编号
	Tid     string       `gorm:"type:varchar(255);not null;index:Gid"` // 消息分组
	Members []*gw.Member `gorm:"type:json"`                            // 消息会员
	Index   int          `gorm:""`                                     // 发送进度
	Uri     string       `gorm:"type:varchar(255);"`
	Data    []byte       `gorm:""`
}

type MsgDb interface {
	Insert(msg *MsgD) error                                                            // 插入消息
	Next(gid string, lastId int64, limit int) []MsgD                                   // 遍历消息
	LastId(gid string, limit int) int64                                                // 获取最近多少条起始Id
	Last(gid string, limit int) []MsgD                                                 // 初始消息缓存
	Delete(id int64) error                                                             // 删除消息
	DeleteF(fid int64) error                                                           // 删除来源消息
	Clear(oId int64) error                                                             // 清理过期消息
	UpdateF(id int64, fid int64) error                                                 // 更新Fid，发送成功处理
	FidGet(fid int64, gid string) int64                                                // 有关联状态
	FidRange(fid int64, step int, idMax int64, idMin int64, fun func(msgD *MsgD) bool) // 遍历超时状态Msg，Fid=F_Fail, 发送失败 超时处理
	TeamInsert(msgTeam *MsgTeam) error                                                 // 群组消息插入
	TeamUpdate(msgTeam *MsgTeam, index int) error                                      // 群组消息更新 index >= mLen || index < 0 TeamDelete
	TeamList(tid string, limit int) []MsgTeam                                          // 群组消息列表
	TeamStarts(workId int32, limit int) []string                                       // 群组消息发送管道,冷启动tid列表
}

type MsgGorm struct {
	db *gorm.DB
}

func (that MsgGorm) AutoMigrate() {
	migrator := that.db.Migrator()
	if (!migrator.HasTable(&MsgD{})) {
		migrator.AutoMigrate(&MsgD{})
	}

	if (!migrator.HasTable(&MsgTeam{})) {
		migrator.AutoMigrate(&MsgTeam{})
	}
}

func (that MsgGorm) Insert(msg *MsgD) error {
	return that.db.Create(msg).Error
}

func (that MsgGorm) Next(gid string, lastId int64, limit int) []MsgD {
	var msgDS []MsgD = nil
	that.db.Where("Gid = ? AND Id > ?", gid, lastId).Order("Id").Limit(limit).Find(&msgDS)
	return msgDS
}

func (that MsgGorm) LastId(gid string, limit int) int64 {
	var id int64 = 0
	that.db.Exec("SELECT Id FROM MsgD WHERE gid = ? ORDER BY Id DESC LIMIT ?, 1", gid, limit).First(&id)
	return id
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

func (that MsgGorm) Delete(id int64) error {
	return that.db.Exec("DELETE FROM MsgD WHERE id = ?", id).Error
}

func (that MsgGorm) DeleteF(fid int64) error {
	return that.db.Exec("DELETE FROM MsgD WHERE Fid = ?", fid).Error
}

func (that MsgGorm) Clear(oId int64) error {
	return that.db.Exec("DELETE FROM MsgD WHERE Id <= ?", oId).Error
}

func (that MsgGorm) UpdateF(id int64, fid int64) error {
	return that.db.Exec("UPDATE MsgD SET Fid = ? WHERE Id <= ?", fid, id).Error
}

func (that MsgGorm) FidGet(fid int64, gid string) int64 {
	var id int64 = 0
	that.db.Exec("SELECT Id FROM MsgD WHERE fid = ? AND gid = ?", fid, gid).First(&id)
	return id
}

func (that MsgGorm) FidRange(fid int64, step int, idMax int64, idMin int64, fun func(msgD *MsgD) bool) {
	id := int64(0)
	var msgDS []MsgD = nil
	var msgD *MsgD
	for {
		that.db.Where("Fid = ? AND Id > ? AND Id < ?", fid, id, idMax).Order("Id").Limit(step).Find(&msgDS)
		mLen := len(msgDS)
		if mLen == 0 {
			break
		}

		for i := 0; i < mLen; i++ {
			msgD = &msgDS[i]
			if !fun(msgD) && msgD.Id <= idMin {
				that.Delete(msgD.Id)
			}
		}

		id = msgD.Id
	}
}

func (that MsgGorm) TeamInsert(msgTeam *MsgTeam) error {
	return that.db.Create(msgTeam).Error
}

func (that MsgGorm) TeamUpdate(msgTeam *MsgTeam, index int) error {
	tLen := 0
	if index < 0 {
		tLen = index

	} else {
		if index == msgTeam.Index {
			return nil
		}

		if msgTeam.Members != nil {
			tLen = len(msgTeam.Members)
		}
	}

	if index >= tLen {
		return that.db.Exec("DELETE FROM MsgTeam WHERE Id <= ?", msgTeam.Id).Error

	} else {
		return that.db.Exec("UPDATE MsgTeam SET Index = ? WHERE Id <= ?", msgTeam.Id, index).Error
	}
}

func (that MsgGorm) TeamList(tid string, limit int) []MsgTeam {
	var MsgTeams []MsgTeam = nil
	that.db.Where("Tid = ?", tid).Order("Id").Limit(limit).Find(&MsgTeams)
	return MsgTeams
}

func (that MsgGorm) TeamStarts(workId int32, limit int) []string {
	var tIds []string = nil
	that.db.Exec("SELECT Tid FROM MsgTeam GROUP BY Tid").Limit(limit).Find(&tIds)
	return tIds
}
