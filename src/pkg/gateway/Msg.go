package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtJson"
	"axjGW/gen/gw"
	"gorm.io/gorm"
	"strings"
)

type Msg interface {
	Get() *MsgD
	Unique() string
}

type MsgD struct {
	Id   int64  `gorm:"primary_key"`                                    // 消息编号
	Gid  string `gorm:"type:varchar(255);not null;index:Gid,type:hash"` // 消息分组
	Fid  int64  `gorm:"index:Fid,type:hash"`                            // 消息来源编号, 从哪条发送消息编号生成的
	Uri  string `gorm:"type:varchar(255);"`
	Data []byte `gorm:""`
	// 压缩后Data不映射字段
	cData []byte `sql:"-"`
	cDid  bool   `sql:"-"`
}

func (that *MsgD) Get() *MsgD {
	return that
}

func (that *MsgD) Unique() string {
	return ""
}

func (that *MsgD) CData() ([]byte, bool) {
	if that.cData != nil {
		return that.cData, that.cDid
	}

	CompressorCData(that.Data, &that.cData, &that.cDid)
	return that.cData, that.cDid
}

type MsgU struct {
	MsgD
	unique string
}

func (m *MsgU) Unique() string {
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
	Id         int64        `gorm:"primary_key"`                                    // 消息编号
	Sid        string       `gorm:""`                                               // 发送者编号
	Tid        string       `gorm:"type:varchar(255);not null;index:Tid,type:hash"` // 消息分组
	Members    []*gw.Member `gorm:"-"`                                              // 消息会员
	MembersS   string       `gorm:"column:members;type:json"`                       // 消息会员存储
	Index      int          `gorm:""`                                               // 发送进度
	Rand       int          `gorm:""`                                               // 发送顺序随机
	Uri        string       `gorm:"type:varchar(255);"`                             // 消息路由
	Data       []byte       `gorm:""`                                               // 消息内容
	Unique     string       `gorm:""`                                               // 唯一消息
	UnreadFeed int          `gorm:""`                                               // 未读消息扩散类型 0 不扩散 1 读扩散 2 写扩散
}

type MsgRead struct {
	Gid    string `gorm:"type:varchar(255);not null;primary_key"` // 消息分组
	LastId int64  `gorm:""`                                       // 最后消息编号
}

type MsgReadNum struct {
	Gid  string
	Num  int32
	Id   int64
	Uri  string
	Data []byte
}

type MsgDb interface {
	Insert(msg *MsgD) error                                                            // 插入消息
	Next(gid string, lastId int64, limit int) []*MsgD                                  // 遍历消息
	LastId(gid string, limit int) int64                                                // 获取最近多少条起始Id
	Last(gid string, limit int) []*MsgD                                                // 初始消息缓存
	Delete(id int64) error                                                             // 删除消息
	DeleteF(fid int64) error                                                           // 删除来源消息
	Clear(oId int64) error                                                             // 清理过期消息
	UpdateF(id int64, fid int64) error                                                 // 更新Fid，发送成功处理
	FidGet(fid int64, gid string) int64                                                // 有关联状态
	FidRange(fid int64, step int, idMax int64, idMin int64, fun func(msgD *MsgD) bool) // 遍历超时状态Msg，Fid=F_Fail, 发送失败 超时处理
	TeamInsert(msgTeam *MsgTeam) error                                                 // 群组消息插入
	TeamUpdate(msgTeam *MsgTeam, index int) error                                      // 群组消息更新 index >= mLen || index < 0 TeamDelete
	TeamList(tid string, limit int) []*MsgTeam                                         // 群组消息列表
	TeamStarts(workId int32, limit int) []string                                       // 群组消息发送管道,冷启动tid列表
	Revoke(delete bool, id int64, gid string, push func() error) error                 // 撤销消息
	Read(gid string, lastId int64) error                                               // 消息已读
	UnReads(gid string, tids []string) []MsgReadNum                                    // 未读消息列表
}

type MsgGorm struct {
	db *gorm.DB
}

func (that *MsgGorm) AutoMigrate() {
	migrator := that.db.Migrator()
	if (!migrator.HasTable(&MsgD{})) {
		migrator.AutoMigrate(&MsgD{})
	}

	if (!migrator.HasTable(&MsgTeam{})) {
		migrator.AutoMigrate(&MsgTeam{})
	}

	if (!migrator.HasTable(&MsgRead{})) {
		migrator.AutoMigrate(&MsgRead{})
	}
}

func (that *MsgGorm) Insert(msg *MsgD) error {
	return that.db.Create(msg).Error
}

func (that *MsgGorm) Next(gid string, lastId int64, limit int) []*MsgD {
	var msgDS []*MsgD = nil
	that.db.Table("msg_ds as a").Select("a.*, b.data as cData").Joins("LEFT JOIN msg_ds as b ON b.id = a.fid and b.id != a.id").Where("a.gid = ? AND a.id > ?", gid, lastId).Order("id").Limit(limit).Find(&msgDS)
	if msgDS != nil {
		mLen := len(msgDS)
		for i := 0; i < mLen; i++ {
			var msgD = msgDS[i]
			if msgD.cData != nil && len(msgD.cData) > 0 {
				msgD.Data = msgD.cData
				msgD.cData = nil
			}
		}
	}
	return msgDS
}

func (that *MsgGorm) LastId(gid string, limit int) int64 {
	var id int64 = 0
	that.db.Raw("SELECT id FROM msg_ds WHERE gid = ? ORDER BY id DESC LIMIT ?, 1", gid, limit).Find(&id)
	return id
}

func (that *MsgGorm) Last(gid string, limit int) []*MsgD {
	var msgDS []*MsgD = nil
	that.db.Table("msg_ds as a").Select("a.*, b.data as cData").Joins("LEFT JOIN msg_ds as b ON b.id = a.fid and b.id != a.id").Where("a.gid = ?", gid).Order("a.id DESC").Limit(limit).Find(&msgDS)
	if msgDS != nil {
		// 倒序
		mLen := len(msgDS)
		last := mLen - 1
		mLen = mLen / 2
		for i := 0; i < mLen; i++ {
			msgD := msgDS[i]
			if msgD.cData != nil && len(msgD.cData) > 0 {
				msgD.Data = msgD.cData
				msgD.cData = nil
			}

			j := last - i
			msgDS[i] = msgDS[j]
			msgDS[j] = msgD
		}
	}

	return msgDS
}

func (that *MsgGorm) Delete(id int64) error {
	return that.db.Exec("DELETE FROM msg_ds WHERE id = ?", id).Error
}

func (that *MsgGorm) DeleteF(fid int64) error {
	return that.db.Exec("DELETE FROM msg_ds WHERE fid = ?", fid).Error
}

func (that *MsgGorm) Clear(oId int64) error {
	return that.db.Exec("DELETE FROM msg_ds WHERE id <= ?", oId).Error
}

func (that *MsgGorm) UpdateF(id int64, fid int64) error {
	return that.db.Exec("UPDATE msg_ds SET fid = ? WHERE id = ?", fid, id).Error
}

func (that *MsgGorm) FidGet(fid int64, gid string) int64 {
	var id int64 = 0
	that.db.Raw("SELECT id FROM msg_ds WHERE fid = ? AND gid = ?", fid, gid).Find(&id)
	return id
}

func (that *MsgGorm) FidRange(fid int64, step int, idMax int64, idMin int64, fun func(msgD *MsgD) bool) {
	id := int64(0)
	var msgDS []MsgD = nil
	var msgD *MsgD
	for {
		that.db.Where("fid = ? AND id > ? AND id < ?", fid, id, idMax).Order("id").Limit(step).Find(&msgDS)
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

func (that *MsgGorm) TeamInsert(msgTeam *MsgTeam) error {
	if msgTeam.MembersS == "" && msgTeam.Members != nil {
		msgTeam.MembersS, _ = KtJson.ToJsonStr(msgTeam.Members)
	}

	err := that.db.Create(msgTeam).Error
	if err != nil {
		return err
	}

	if msgTeam.Unique != "" {
		// 删除重复消息
		that.db.Exec("DELETE FROM msg_teams WHERE tid = ? AND unique = ? AND id < ?", msgTeam.Tid, msgTeam.Unique, msgTeam.Id)
	}

	return nil
}

func (that *MsgGorm) TeamUpdate(msgTeam *MsgTeam, index int) error {
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
		return that.db.Exec("DELETE FROM msg_teams WHERE id <= ?", msgTeam.Id).Error

	} else {
		return that.db.Exec("UPDATE msg_teams SET index = ? WHERE id <= ?", msgTeam.Id, index).Error
	}
}

func (that *MsgGorm) TeamList(tid string, limit int) []*MsgTeam {
	var msgTeams []*MsgTeam = nil
	tx := that.db.Where("tid = ?", tid).Order("id").Limit(limit).Find(&msgTeams)
	Kt.Err(tx.Error, false)
	for _, msgTeam := range msgTeams {
		if msgTeam.MembersS != "" {
			KtJson.FromJsonStr(msgTeam.MembersS, &msgTeam.Members)
		}
	}

	return msgTeams
}

func (that *MsgGorm) TeamStarts(workId int32, limit int) []string {
	var tIds []string = nil
	that.db.Raw("SELECT tid FROM msg_teams GROUP BY tid").Limit(limit).Find(&tIds)
	return tIds
}

func (that *MsgGorm) Revoke(delete bool, id int64, gid string, push func() error) error {
	var tid int64 = 0
	that.db.Raw("SELECT id FROM msg_ds WHERE id = ? AND gid = ?", id, gid).Find(&tid)
	if tid <= 0 {
		return ANet.ERR_DENIED
	}

	var err error
	if !delete {
		err = that.DeleteF(id)
		if err != nil {
			return err
		}
	}

	if push != nil {
		err = push()
		if err != nil {
			return err
		}
	}

	tx := that.db.Exec("DELETE FROM msg_ds WHERE id = ? AND gid = ?", id, gid)
	return tx.Error
}

func (that *MsgGorm) Read(gid string, lastId int64) error {
	var id = ""
	that.db.Raw("SELECT gid FROM msg_reads WHERE gid = ? ", gid).First(&id)
	if id == "" {
		return that.db.Create(&MsgRead{Gid: gid, LastId: lastId}).Error

	} else {
		return that.db.Exec("UPDATE msg_reads SET lastId = ? WHERE gid <= ? AND lastId < ?", lastId, gid, lastId).Error
	}
}

func (that *MsgGorm) UnRead(gid string) int {
	var num = 0
	//that.db.Raw("SELECT COUNT(a.id) FROM msg_ds LEFT JOIN msg_reads as b ON b.gid = a.gid WHERE a.gid = ? AND a.id > b.lastId", gid).First(&num)
	var lastId int64 = 0
	that.db.Raw("SELECT lastId FROM msg_reads WHERE gid = ? ", gid).First(&lastId)
	that.db.Raw("SELECT COUNT(id) FROM msg_ds WHERE gid = ? AND id > ?", gid, lastId).First(&num)
	return num
}

func (that *MsgGorm) UnReads(gid string, tids []string) []MsgReadNum {
	sb := strings.Builder{}
	for i, tid := range tids {
		if i > 0 {
			sb.WriteByte(',')
		}

		sb.WriteByte('"')
		sb.WriteString(tid)
		sb.WriteByte('"')
	}

	var numAs []MsgReadNum = nil
	that.db.Raw("SELECT a.num as num, a.gid as gid, b.id as id, b.uri as uri, b.data as data FROM (" +
		"SELECT COUNT(a.id) as num, gid as gid, MAX(id) as id FROM msg_ds LEFT JOIN msg_reads as b ON b.gid = CONCAT(\"" + gid + "/\", a.gid) WHERE a.gid IN (" + sb.String() + ") AND a.id > b.lastId GROUP BY a.gid" +
		") as a LEFT JOIN msg_ds as b ON b.id = a.id",
	).Find(&numAs)

	sb.Reset()
	for i, tid := range tids {
		if i > 0 {
			sb.WriteByte(',')
		}

		sb.WriteString(MsgMng().GidForTid(gid, tid))
	}

	var numBs []MsgReadNum = nil
	that.db.Raw("SELECT a.num as num, a.gid as gid, b.id as id, b.uri as uri, b.data as data FROM (" +
		"SELECT COUNT(a.id) as num, gid as gid, MAX(id) as id FROM msg_ds LEFT JOIN msg_reads as b ON b.gid = a.gid WHERE a.gid IN (" + sb.String() + ") AND a.id > b.lastId GROUP BY a.gid" +
		") as a LEFT JOIN msg_ds as b ON b.id = a.id",
	).Find(&numBs)
	for _, numA := range numAs {
		numA.Gid = MsgMng().GidForTid(gid, numA.Gid)
	}

	for i, numA := range numAs {
		for _, numB := range numBs {
			if numA.Gid == numB.Gid {
				numB.Num += numA.Num
				if numB.Id < numA.Id {
					numB.Id = numA.Id
					numB.Uri = numA.Uri
					numB.Data = numA.Data
				}

				i = -1
			}
		}

		if i >= 0 {
			numBs = append(numBs, numA)
		}
	}

	return numBs
}
