package gateway

import (
	"axj/ANet"
	"axj/APro"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
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
		FDrt:         3000,
		FStep:        20,
		FTimeout:     9000,
		FTimeoutD:    60000,
		TStartsDrt:   3000,
		TStartsLimit: 3000,
		TStartLimit:  30,
		TIdleLive:    30000,
		TPushQueue:   30,
	}

	// 配置处理
	APro.SubCfgBind("chat", _chatMng)
	that.FDrt = that.FDrt * time.Millisecond
	that.FTimeout = that.FTimeout * int64(time.Millisecond)
	that.FTimeoutD = that.FTimeoutD * int64(time.Millisecond)
	that.TStartsDrt = that.TStartsDrt * int64(time.Millisecond)
	that.TIdleLive = that.TIdleLive * int64(time.Millisecond)
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
	for checkLoop == that.checkLoop {
		time.Sleep(that.FDrt)
		checkTime := time.Now().UnixNano()
		that.checkTime = checkTime
		MsgMng().Db.FidRange(F_FAIL, that.FStep, MsgMng().idWorkder.Timestamp(checkTime-that.FTimeout), MsgMng().idWorkder.Timestamp(checkTime-that.FTimeoutD), that.checkMsgD)
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
	F_FAIL     = 1
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
	if fid < R_SUCC_MIN {
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
			if !that.msgTeamPush(&msgTeams[i], true) {
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

	// 已执行索引
	indexDid := index
	for ; index < mLen; index++ {
		// Rand随机发送顺序
		member := members[(index+msgTeam.Rand)%mLen]
		gid := member.Gid
		if member.Nofeed {
			// 不推送到gid， 需要主动拉去tid_gid
			gid = that.tid + "_" + gid
		}

		rep, _ := Server.GetProdGid(gid).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
			Gid:  gid,
			Uri:  msgTeam.Uri,
			Data: msgTeam.Data,
			Qs:   3,
			Fid:  msgTeam.Id,
		})

		var fid = Server.Id64(rep)
		if fid < R_SUCC_MIN {
			// 已执行索引
			if db {
				MsgMng().Db.TeamUpdate(msgTeam, indexDid)
			}

			return false
		}

		indexDid = index
		if !db {
			msgTeam.Index = indexDid
		}
	}

	if db {
		return MsgMng().Db.TeamUpdate(msgTeam, -1) == nil
	}

	return true
}

// 点对点发送聊天 调用注意分布一致hash 入口
func (that *chatMng) Send(fromId string, toId string, uri string, data []byte, db bool) (bool, error) {
	var qs int32 = 3
	if !db {
		qs = 2
	}

	fClient := Server.GetProdGid(fromId).GetGWIClient()
	rep, err := fClient.GPush(Server.Context, &gw.GPushReq{
		Gid:     fromId,
		Uri:     uri,
		Data:    data,
		Qs:      qs,
		Fid:     F_FAIL,
		Isolate: true,
	})

	fid := Server.Id64(rep)
	if fid < R_SUCC_MIN {
		return false, err
	}

	rep, err = Server.GetProdGid(toId).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
		Gid:  toId,
		Uri:  uri,
		Data: data,
		Qs:   qs,
		Fid:  fid,
	})

	tid := Server.Id64(rep)
	if tid < R_SUCC_MIN {
		if db {
			fClient.GPushA(Server.Context, &gw.IGPushAReq{
				Gid:  fromId,
				Id:   fid,
				Succ: false,
			})
		}

		return false, err
	}

	if db {
		fClient.GPushA(Server.Context, &gw.IGPushAReq{
			Gid:  fromId,
			Id:   fid,
			Succ: true,
		})
	}

	return true, err
}

// 发送群聊天 调用注意分布一致hash 入口
func (that *chatMng) TeamPush(fromId string, tid string, readfeed bool, uri string, data []byte, queue bool, db bool) (bool, error) {
	var qs int32 = 3
	if !db {
		qs = 2
	}

	if !readfeed {
		if MsgMng().Db == nil {
			qs = 2
		}

		team := TeamMng().GetTeam(tid)
		if team == nil {
			return false, ANet.ERR_NOWAY
		}

		if !team.ReadFeed {
			mLen := 0
			if team.Members != nil {
				mLen = len(team.Members)
			}

			if mLen <= 0 {
				return false, ANet.ERR_NOWAY
			}

			// 写扩散
			fClient := Server.GetProdGid(fromId).GetGWIClient()
			rep, err := fClient.GPush(Server.Context, &gw.GPushReq{
				Gid:  fromId,
				Uri:  uri,
				Data: data,
				Qs:   qs,
				//Unique: "",
				Queue: queue,
				Fid:   F_FAIL,
				//Isolate: false,
			})

			var fid = Server.Id64(rep)
			if fid < R_SUCC_MIN {
				return false, err
			}

			msgTeam := &MsgTeam{
				Id:      fid,
				Tid:     tid,
				Members: team.Members,
				Index:   0,
				Rand:    int(rand.Int31n(int32(mLen))),
				Uri:     uri,
				Data:    data,
			}

			msgDb := MsgMng().Db != nil && qs == 3

			if msgDb {
				err = MsgMng().Db.TeamInsert(msgTeam)
				if err != nil {
					if db {
						fClient.GPushA(Server.Context, &gw.IGPushAReq{
							Gid:  fromId,
							Id:   fid,
							Succ: false,
						})
					}

					return false, err
				}
			}

			if db {
				fClient.GPushA(Server.Context, &gw.IGPushAReq{
					Gid:  fromId,
					Id:   fid,
					Succ: true,
				})
			}

			if msgDb {
				that.TeamStart(tid, nil)

			} else {
				that.TeamStart(tid, msgTeam)
			}

			return true, nil
		}
	}

	// 读扩散
	rep, err := Server.GetProdGid(tid).GetGWIClient().GPush(Server.Context, &gw.GPushReq{
		Gid:  tid,
		Uri:  uri,
		Data: data,
		Qs:   qs,
		//Unique: "",
		Queue: queue,
		//Isolate: false,
	})

	var fid = Server.Id64(rep)
	if fid < R_SUCC_MIN {
		return false, err
	}

	return true, nil
}
