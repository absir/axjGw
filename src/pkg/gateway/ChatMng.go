package gateway

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"context"
	"math/rand"
	"sync"
	"time"
)

type chatMng struct {
	FDrt         time.Duration // fid特殊状态消息检查间隔(比如消息发送失败)
	FStep        int           // fid特殊状态消息检查, 单次获取消息列表数
	FTimeout     int64         // fid特殊状态， 检查开始超时时间
	FTimeoutD    int64         // fid特殊状态， 检查最大超时时间， 超过删除特殊状态
	TStartsDrt   int64         // 群消息发送管道检查间隔
	TStartsLimit int           // 群消息检查单次获取管道数
	TStartLimit  int           // 群消息检查单次获取列表数
	TIdleLive    int64         // 群消息发送管道，调用空闲存活时间
	TPushQueue   int           // 群消息发送管道, 内存管道最大值
	checkLoop    int64
	checkTime    int64
	tStartTime   int64
	locker       sync.Locker
	teamMap      *cmap.CMap
	teamBuff     []interface{}
}

var _chatMng *chatMng

func ChatMng() *chatMng {
	if _chatMng == nil {
		Server.Locker.Lock()
		defer Server.Locker.Unlock()
		if _chatMng == nil {
			initChatMng()
			go _chatMng.CheckLoop()
		}
	}

	return _chatMng
}

func initChatMng() {
	that := &chatMng{
		FDrt:         3,
		FStep:        20,
		FTimeout:     9,
		FTimeoutD:    60,
		TStartsDrt:   3,
		TStartsLimit: 3000,
		TStartLimit:  30,
		TIdleLive:    30,
		TPushQueue:   30,
	}

	// 配置处理
	APro.SubCfgBind("chat", _chatMng)
	that.FDrt = that.FDrt * time.Second
	that.FTimeout = that.FTimeout * int64(time.Second)
	that.FTimeoutD = that.FTimeoutD * int64(time.Second)
	that.TStartsDrt = that.TStartsDrt * int64(time.Second)
	that.TIdleLive = that.TIdleLive * int64(time.Second)
	that.locker = new(sync.Mutex)
	that.teamMap = cmap.NewCMapInit()
	_chatMng = that
}

// 空闲检测
func (that *chatMng) CheckStop() {
	that.checkLoop = -1
}

func (that *chatMng) CheckLoop() {
	if MsgMng().Db == nil {
		return
	}

	checkLoop := time.Now().UnixNano()
	that.checkLoop = checkLoop
	for Kt.Active && checkLoop == that.checkLoop {
		time.Sleep(that.FDrt)
		checkTime := time.Now().UnixNano()
		that.checkTime = checkTime
		MsgMng().Db.FidRange(F_SENDING, that.FStep, MsgMng().idWorker.Timestamp(checkTime-that.FTimeout), MsgMng().idWorker.Timestamp(checkTime-that.FTimeoutD), that.checkMsgD)
		if that.tStartTime < checkTime {
			that.tStartTime = checkTime + that.TStartsDrt
			tIds := MsgMng().Db.TeamStarts(Config.WorkId, that.TStartsLimit)
			if tIds != nil {
				tLen := len(tIds)
				for i := 0; i < tLen; i++ {
					// 启动组管道
					tId := tIds[i]
					Server.GetProdGid(tId).GetGWIClient().TStarts(Server.Context, &gw.GidReq{
						Gid: tId,
					})
				}
			}

			// RangeBuff内存复用、快速、安全
			that.teamMap.RangeBuff(that.checkChatTeam, &that.teamBuff, that.TStartsLimit)
		}
	}
}

func (that *chatMng) checkMsgD(msgD *MsgD) bool {
	return that.MsgFail(msgD.Id, msgD.Gid) == nil
}

func (that *chatMng) checkChatTeam(key, val interface{}) bool {
	chatTeam, _ := val.(*ChatTeam)
	if chatTeam == nil {
		that.teamMap.Delete(key)

	} else if chatTeam.starting != 1 && chatTeam.starting < that.checkTime {
		that.teamMap.Delete(key)
	}

	return true
}

const (
	F_SUCC     = 0
	F_SENDING  = 1
	R_SUCC_MIN = 16
)

func (that *chatMng) MsgSucc(id int64) error {
	if MsgMng().Db == nil {
		return ERR_NOWAY
	}

	return MsgMng().Db.UpdateF(id, F_SUCC)
}

func (that *chatMng) MsgFail(id int64, gid string) error {
	if MsgMng().Db == nil {
		return ERR_NOWAY
	}

	// 未发送
	rep, err := Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
		Gid: gid,
		//Uri: "",
		//Data: nil,
		Qs: 3,
		//Unique: "",
		//Queue: false,
		Fid: id,
		//Isolate: false,
	})

	if err != nil {
		return err
	}

	var fid = Server.Id64(rep)
	if !Server.Id64Succ(fid, true) {
		return ERR_FAIL
	}

	return MsgMng().Db.Delete(id)
}

type ChatTeam struct {
	tid      string
	starting int64
	msgQueue *Util.CircleQueue
}

func (that *chatMng) TeamStart(tid string, msgTeam *MsgTeam) {
	val, _ := that.teamMap.Load(tid)
	chatTeam, _ := val.(*ChatTeam)
	if chatTeam == nil {
		that.locker.Lock()
		val, _ = that.teamMap.Load(tid)
		chatTeam, _ = val.(*ChatTeam)
		if chatTeam == nil {
			chatTeam = &ChatTeam{
				tid: tid,
			}

			that.teamMap.Store(tid, chatTeam)
		}

		that.locker.Unlock()
	}

	chatTeam.Start(msgTeam)
}

func (that *ChatTeam) addMsgTeam(msgTeam *MsgTeam) {
	if msgTeam == nil || _chatMng.TPushQueue <= 0 {
		return
	}

	MsgMng().locker.Lock()
	if that.msgQueue == nil {
		that.msgQueue = Util.NewCircleQueue(_chatMng.TPushQueue)
	}

	that.msgQueue.Push(msgTeam, true)
	MsgMng().locker.Unlock()
}

func (that *ChatTeam) getMsgTeam() *MsgTeam {
	if that.msgQueue == nil {
		return nil
	}

	MsgMng().locker.Lock()
	if that.msgQueue == nil {
		MsgMng().locker.Unlock()
		return nil
	}

	val, _ := that.msgQueue.Get(0)
	MsgMng().locker.Unlock()
	msgTeam, _ := val.(*MsgTeam)
	return msgTeam
}

func (that *ChatTeam) removeMsgTeam(msgTeam *MsgTeam) {
	if that.msgQueue == nil {
		return
	}

	MsgMng().locker.Lock()
	if that.msgQueue == nil {
		MsgMng().locker.Unlock()
		return
	}

	that.msgQueue.Remove(msgTeam)
	MsgMng().locker.Unlock()
}

func (that *ChatTeam) Start(msgTeam *MsgTeam) {
	that.addMsgTeam(msgTeam)
	if that.starting != 0 {
		return
	}

	Util.GoSubmit(that.startRun)
}

func (that *ChatTeam) startIn() bool {
	_chatMng.locker.Lock()
	if that.starting != 1 {
		that.starting = 1
		_chatMng.locker.Unlock()
		return true
	}

	_chatMng.locker.Unlock()
	return false
}

func (that *ChatTeam) startOut() {
	_chatMng.locker.Lock()
	that.starting = time.Now().UnixNano() + _chatMng.TIdleLive
	_chatMng.locker.Unlock()
}

func (that *ChatTeam) startRun() {
	// 协程进入保护
	if !that.startIn() {
		return
	}

	defer that.startOut()
	for {
		for {
			msgTeam := that.getMsgTeam()
			if msgTeam == nil {
				break
			}

			if !that.msgTeamPush(msgTeam, false) {
				return
			}

			that.removeMsgTeam(msgTeam)
		}

		msgTeams := MsgMng().Db.TeamList(that.tid, _chatMng.TStartLimit)
		if msgTeams == nil {
			return
		}

		mLen := len(msgTeams)
		if mLen <= 0 {
			return
		}

		for i := 0; i < mLen; i++ {
			if !that.msgTeamPush(msgTeams[i], true) {
				return
			}
		}
	}
}

func (that *ChatTeam) msgTeamPush(msgTeam *MsgTeam, db bool) bool {
	// 执行hash校验
	if !Server.IsProdHashS(msgTeam.Tid) {
		return false
	}

	members := msgTeam.Members
	if members == nil {
		if db {
			MsgMng().Db.TeamUpdate(msgTeam, 1)
		}

		return true
	}

	mLen := len(members)
	index := msgTeam.Index
	if index < 0 {
		index = 0
	}

	if index >= mLen {
		if db {
			MsgMng().Db.TeamUpdate(msgTeam, index)
		}

		return true
	}

	// 群消息持久化
	var qs int32 = 2
	if db {
		qs = 3
	}

	// 已执行索引
	indexDid := index
	for ; index < mLen; index++ {
		// Rand随机发送顺序
		member := members[(index+msgTeam.Rand)%mLen]
		gid := member.Gid
		// 群消息不需要再发送给自己
		if gid != msgTeam.Sid {
			if msgTeam.UnreadFeed != 2 {
				if member.Nofeed || msgTeam.UnreadFeed > 0 {
					// 不推送到gid， 需要主动拉取GidForTid
					gid = MsgMng().GidForTid(gid, that.tid)
				}

				rep, _ := Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
					Gid:    gid,
					Uri:    msgTeam.Uri,
					Data:   msgTeam.Data,
					Qs:     qs,
					Fid:    msgTeam.Id,
					Unique: msgTeam.Unique,
				})

				var fid = Server.Id64(rep)
				if !Server.Id64Succ(fid, true) {
					// 已执行索引
					if db {
						MsgMng().Db.TeamUpdate(msgTeam, indexDid)
					}

					return false
				}
			}

			if msgTeam.UnreadFeed > 0 {
				// 未读消息发送
				gid = member.Gid
				rep, _ := Server.GetProdGid(gid).GetGWIClient().Unread(Server.Context, &gw.UnreadReq{
					Gid:    gid,
					Tid:    msgTeam.Tid,
					LastId: msgTeam.Id,
					Uri:    msgTeam.Uri,
					Data:   msgTeam.Data,
					Entry:  true,
				})

				if rep == nil || !Server.Id32Succ(rep.Id) {
					// 已执行索引
					if db {
						MsgMng().Db.TeamUpdate(msgTeam, indexDid)
					}

					return false
				}
			}
		}

		indexDid = index
		if db {
			msgTeam.Index = indexDid
		}
	}

	if db {
		return MsgMng().Db.TeamUpdate(msgTeam, -1) == nil
	}

	return true
}

// 点对点发送聊天 调用注意分布一致hash 入口
func (that *chatMng) Send(req *gw.SendReq) (bool, error) {
	var qs int32 = 2
	if req.Db {
		qs = 3
	}

	var fidStatus int64 = 0
	if req.Db && MsgMng().Db != nil {
		fidStatus = F_SENDING
	}

	fClient := Server.GetProdGid(req.FromId).GetGWIClient()
	rep, err := fClient.GPush(Server.Context, &gw.GPushReq{
		Gid:     req.FromId,
		Uri:     req.Uri,
		Data:    req.Data,
		Qs:      qs,
		Fid:     fidStatus,
		Isolate: true,
	})

	fid := Server.Id64(rep)
	if !Server.Id64Succ(fid, true) {
		return false, err
	}

	rep, err = Server.GetProdGid(req.ToId).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
		Gid:  req.ToId,
		Uri:  req.Uri,
		Data: req.Data,
		Qs:   qs,
		Fid:  fid,
	})

	tid := Server.Id64(rep)
	if !Server.Id64Succ(tid, true) {
		if fidStatus > 0 {
			fClient.GPushA(Server.Context, &gw.IGPushAReq{
				Gid:  req.FromId,
				Id:   fid,
				Succ: false,
			})
		}

		return false, err
	}

	if fidStatus > 0 {
		fClient.GPushA(Server.Context, &gw.IGPushAReq{
			Gid:  req.FromId,
			Id:   fid,
			Succ: true,
		})
	}

	return true, err
}

// 发送群聊天 调用注意分布一致hash 入口
func (that *chatMng) TeamPush(req *gw.TPushReq) (bool, error) {
	// 群成员
	var team *gw.TeamRep = nil
	mLen := 0
	unreadFeed := false
	if !req.ReadFeed || req.UnreadFeed {
		team = TeamMng().GetTeam(req.Tid)
		if team == nil {
			return false, ANet.ERR_NOWAY
		}

		unreadFeed = team.UnreadFeed || req.UnreadFeed
	}

	var qs int32 = 2
	if req.Db {
		qs = 3
	}

	if !req.ReadFeed {
		// 写扩散
		if MsgMng().Db == nil {
			qs = 2
		}

		var fidStatus int64 = 0
		if req.Db && MsgMng().Db != nil {
			fidStatus = F_SENDING
		}

		// 写扩散
		fClient := Server.GetProdGid(req.FromId).GetGWIClient()
		var fid int64 = 0
		if req.FromId != "" {
			rep, err := fClient.GPush(Server.Context, &gw.GPushReq{
				Gid:   req.FromId,
				Uri:   req.Uri,
				Data:  req.Data,
				Qs:    qs,
				Queue: req.Queue,
				Fid:   fidStatus,
			})

			fid = Server.Id64(rep)
			if !Server.Id64Succ(fid, true) {
				return false, err
			}
		}

		msgTeam := &MsgTeam{
			Id:      fid,
			Sid:     req.FromId,
			Tid:     req.Tid,
			Members: team.Members,
			Index:   0,
			Rand:    int(rand.Int31n(int32(mLen))),
			Uri:     req.Uri,
			Data:    req.Data,
			Unique:  req.Unique,
		}

		if unreadFeed {
			// 写扩散，未读消息扩散
			msgTeam.UnreadFeed = 1
		}

		msgDb := MsgMng().Db != nil && qs == 3

		if msgTeam.Id <= 0 {
			msgTeam.Id = MsgMng().idWorker.Generate()
		}

		if msgDb {
			err := MsgMng().Db.TeamInsert(msgTeam)
			if err != nil {
				if fidStatus > 0 {
					fClient.GPushA(Server.Context, &gw.IGPushAReq{
						Gid:  req.FromId,
						Id:   fid,
						Succ: false,
					})
				}

				return false, err
			}
		}

		if fidStatus > 0 {
			fClient.GPushA(Server.Context, &gw.IGPushAReq{
				Gid:  req.FromId,
				Id:   fid,
				Succ: true,
			})
		}

		if msgDb {
			that.TeamStart(req.Tid, nil)

		} else {
			that.TeamStart(req.Tid, msgTeam)
		}

		return true, nil
	}

	// 读扩散
	unique := req.Unique
	rep, err := Server.GetProdGid(req.Tid).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
		Gid:    req.Tid,
		Uri:    req.Uri,
		Data:   req.Data,
		Qs:     qs,
		Unique: unique,
		Queue:  req.Queue,
	})

	var fid = Server.Id64(rep)
	if !Server.Id64Succ(fid, true) {
		return false, err
	}

	if unreadFeed && mLen > 0 {
		// 读扩散，未读消息扩散
		msgTeam := &MsgTeam{
			Id:         fid,
			Sid:        req.FromId,
			Tid:        req.Tid,
			Members:    team.Members,
			Index:      0,
			Rand:       int(rand.Int31n(int32(mLen))),
			Uri:        req.Uri,
			Data:       req.Data,
			Unique:     req.Unique,
			UnreadFeed: 2,
		}

		// 未读扩散持久化
		msgDb := MsgMng().Db != nil && qs == 3
		if msgTeam.Id <= 0 {
			msgTeam.Id = MsgMng().idWorker.Generate()
		}

		if msgDb {
			err = MsgMng().Db.TeamInsert(msgTeam)
			if err != nil {
				return false, err
			}
		}

		// 启动扩散任务
		if msgDb {
			that.TeamStart(req.Tid, nil)

		} else {
			that.TeamStart(req.Tid, msgTeam)
		}
	}

	return true, nil
}

func (that *chatMng) Revoke(ctx context.Context, req *gw.RevokeReq) (bool, error) {
	if MsgMng().Db == nil {
		return false, nil
	}

	var push func() error = nil
	if req.Push != nil {
		push = func() error {
			rep, err := Server.GetProdGid(req.Push.Gid).GetGWIClient().GPush(ctx, req.Push)
			if err != nil {
				return err
			}

			if rep == nil || !Server.Id64Succ(rep.Id, true) {
				return ERR_FAIL
			}

			return nil
		}

	} else if req.TPush != nil {
		push = func() error {
			rep, err := Server.GetProdGid(req.TPush.GetTid()).GetGWIClient().TPush(ctx, req.TPush)
			if err != nil {
				return err
			}

			if rep == nil || rep.Id < R_SUCC_MIN {
				return ERR_FAIL
			}

			return nil
		}
	}

	err := MsgMng().Db.Revoke(req.Delete, req.Id, req.Gid, push)
	if err != nil {
		return false, err
	}

	return true, nil
}
