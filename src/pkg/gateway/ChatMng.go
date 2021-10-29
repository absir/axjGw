package gateway

import (
	"axj/APro"
	"sync"
	"time"
)

type chatMng struct {
	FDrt         time.Duration
	FStep        int
	FTimeout     int64
	FTimeoutD    int64
	TStartsDrt   int64
	TStartsLimit int
	TStartLimit  int
	TLive        int64
	loopTime     int64
	checkLoop    int64
	checkTime    int64
	tStartTime   int64
	locker       sync.Locker
	teamMap      *sync.Map
}

var ChatMng *chatMng

func initChatMng() {
	ChatMng = &chatMng{
		FDrt:         3000,
		FStep:        20,
		FTimeout:     9000,
		FTimeoutD:    60000,
		TStartsDrt:   3000,
		TStartsLimit: 3000,
		TStartLimit:  30,
		TLive:        30000,
	}

	// 配置处理
	APro.SubCfgBind("chat", ChatMng)
	that := ChatMng
	that.FDrt = that.FDrt * time.Millisecond
	that.FTimeout = that.FTimeout * int64(time.Millisecond)
	that.FTimeoutD = that.FTimeoutD * int64(time.Millisecond)
	that.TStartsDrt = that.TStartsDrt * int64(time.Millisecond)
	that.TLive = that.TLive * int64(time.Millisecond)
	that.locker = new(sync.Mutex)
	that.teamMap = new(sync.Map)
}

// 空闲检测
// 空闲检测
func (that chatMng) CheckStop() {
	that.checkLoop = -1
}

func (that chatMng) CheckLoop() {
	loopTime := time.Now().UnixNano()
	that.checkLoop = loopTime
	for loopTime == that.checkLoop {
		time.Sleep(that.FDrt)
		checkTime := time.Now().UnixNano()
		that.checkTime = checkTime
		MsgMng.Db.FidRange(F_FAIL, that.FStep, MsgMng.idWorkder.Timestamp(checkTime-that.FTimeout), MsgMng.idWorkder.Timestamp(checkTime-that.FTimeoutD), that.checkMsgD)
		if that.tStartTime < checkTime {
			that.tStartTime = checkTime + that.TStartsDrt
			tIds := MsgMng.Db.TeamStarts(Config.WorkId, that.TStartsLimit)
			if tIds != nil {
				tLen := len(tIds)
				for i := 0; i < tLen; i++ {
					// 启动组管道
					tId := tIds[i]
					Server.GetProdGid(tId).GetGWIClient().TeamStarts(Server.Context, tId)
				}
			}

			that.teamMap.Range(that.checkChatTeam)
		}
	}
}

func (that chatMng) checkMsgD(msgD *MsgD) bool {
	return that.MsgFail(msgD.Id, msgD.Gid) == nil
}

func (that chatMng) checkChatTeam(key, val interface{}) bool {
	chatTeam := val.(*ChatTeam)
	if chatTeam == nil {
		that.teamMap.Delete(key)

	} else if chatTeam.starting != 1 && chatTeam.starting < that.checkTime {
		that.teamMap.Delete(key)
	}

	return true
}

const (
	F_SUCC = 0
	F_FAIL = 1
)

// 点对点发送聊天
func (that chatMng) Send(fromId string, toId string, uri string, bytes []byte) (bool, error) {
	fClient := Server.GetProdGid(fromId).GetGWIClient()
	fid, err := fClient.GPush(Server.Context, fromId, uri, bytes, true, 3, false, "", F_FAIL)
	if fid < 32 {
		return false, err
	}

	tid, err := Server.GetProdGid(toId).GetGWIClient().GPush(Server.Context, toId, uri, bytes, true, 3, false, "", fid)
	if tid < 32 {
		fClient.GPushA(Server.Context, fromId, fid, false)
		return false, err
	}

	fClient.GPushA(Server.Context, fromId, fid, true)
	return true, err
}

func (that chatMng) MsgSucc(id int64) error {
	if MsgMng.Db == nil {
		return ERR_NOWAY
	}

	return MsgMng.Db.UpdateF(id, F_SUCC)
}

func (that chatMng) MsgFail(id int64, gid string) error {
	if MsgMng.Db == nil {
		return ERR_NOWAY
	}

	// 未发送
	result, err := Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, gid, "", nil, false, 3, false, "", id)
	if err != nil {
		return err
	}

	if result <= 32 {
		return ERR_FAIL
	}

	return MsgMng.Db.Delete(id)
}

// 发送群聊天
func (that chatMng) Team(fromId string, tid string, readfeed bool, uri string, bytes []byte, qs int32, queue bool, unique string) (bool, error) {
	if !readfeed {
		if MsgMng.Db == nil {
			return false, ERR_NOWAY
		}

		team := TeamMng.GetTeam(tid)
		if !team.ReadFeed {
			// 写扩散
			fClient := Server.GetProdGid(fromId).GetGWIClient()
			fid, err := fClient.GPush(Server.Context, fromId, uri, bytes, true, 3, false, "", F_FAIL)
			if fid < 32 {
				return false, err
			}

			msgTeam := &MsgTeam{
				Id:      fid,
				Tid:     tid,
				Members: team.Members,
				Index:   0,
				Uri:     uri,
				Data:    bytes,
			}

			err = MsgMng.Db.TeamInsert(msgTeam)
			if err != nil {
				fClient.GPushA(Server.Context, fromId, fid, false)
				return false, err
			}

			fClient.GPushA(Server.Context, fromId, fid, true)
			Server.GetProdGid(tid).GetGWIClient().TeamStarts(Server.Context, tid)
			return true, nil
		}
	}

	// 读扩散
	ret, err := Server.GetProdGid(tid).GetGWIClient().GPush(Server.Context, tid, uri, bytes, false, qs, queue, unique, 0)
	if err != nil {
		return false, err
	}

	if ret <= 32 {
		return false, ERR_FAIL
	}

	return true, nil
}

func (that chatMng) TeamStarts(tid string) {

}

type ChatTeam struct {
	tid      string
	starting int64
}

func (that chatMng) teamStart(tid string) {
	val, _ := that.teamMap.Load(tid)
	chatTeam := val.(*ChatTeam)
	if chatTeam == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		val, _ = that.teamMap.Load(tid)
		chatTeam = val.(*ChatTeam)
		if chatTeam == nil {
			chatTeam = &ChatTeam{
				tid: tid,
			}

			that.teamMap.Store(tid, chatTeam)
		}
	}

	chatTeam.Start()
}

func (that ChatTeam) Start() {
	if that.starting != 0 {
		return
	}

	go that.startRun()
}

func (that ChatTeam) startIn() bool {
	ChatMng.locker.Lock()
	defer ChatMng.locker.Unlock()
	if that.starting != 1 {
		that.starting = 1
		return true
	}

	return false
}

func (that ChatTeam) startOut() {
	that.starting = time.Now().UnixNano() + ChatMng.TLive
}

func (that ChatTeam) startRun() {
	// 协程进入保护
	if !that.startIn() {
		return
	}

	defer that.startOut()
	for {
		msgTeams := MsgMng.Db.TeamList(that.tid, ChatMng.TStartLimit)
		if msgTeams == nil {
			return
		}

		mLen := len(msgTeams)
		if mLen <= 0 {
			return
		}

		for i := 0; i < mLen; i++ {
			if !that.msgTeamPush(&msgTeams[i]) {
				return
			}
		}
	}
}

func (that ChatTeam) msgTeamPush(msgTeam *MsgTeam) bool {
	// 执行hash校验
	if !Server.IsProdHashS(msgTeam.Tid) {
		return false
	}

	members := msgTeam.Members
	if members == nil {
		MsgMng.Db.TeamUpdate(msgTeam, 1)
		return true
	}

	mLen := len(members)
	index := msgTeam.Index
	if index < 0 {
		index = 0
	}

	if index >= mLen {
		MsgMng.Db.TeamUpdate(msgTeam, index)
		return true
	}

	// 已执行索引
	indexDid := index
	for ; index < mLen; index++ {
		member := members[index]
		if member.Nofeed {
			// 不推送到gid， 需要主动拉去tid_gid
			gid := that.tid + "_" + member.Gid
			ret, err := Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, gid, msgTeam.Uri, msgTeam.Data, false, 3, false, "", msgTeam.Id)
			if err != nil || ret <= 32 {
				// 已执行索引
				MsgMng.Db.TeamUpdate(msgTeam, indexDid)
				return false
			}

			indexDid = index

		} else {
			// 直接推送到gid
			gid := member.Gid
			ret, err := Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, gid, msgTeam.Uri, msgTeam.Data, false, 3, false, "", msgTeam.Id)
			if err != nil || ret <= 32 {
				// 已执行索引
				MsgMng.Db.TeamUpdate(msgTeam, indexDid)
				return false
			}

			indexDid = index
		}
	}

	return MsgMng.Db.TeamUpdate(msgTeam, -1) == nil
}
