package gateway

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"strconv"
	"sync"
	"time"
)

type MsgGrp struct {
	// 消息组编号
	gid string
	// 消息组hash
	ghash int
	// 管理锁
	locker sync.Locker
	// 消息组读写锁
	rwLocker *sync.RWMutex
	// 消息组db锁
	dbLocker sync.Locker
	// 过期时间
	passTime int64
	// 消息组场
	sess *MsgSess
}

// 延长过期时间
func (that *MsgGrp) retain() {
	that.passTime = time.Now().UnixNano() + _msgMng.LiveDrt
}

// 获取消息管理场客户端数量
func (that *MsgGrp) ClientNum() int {
	sess := that.sess
	if sess == nil {
		return 0
	}

	return sess.clientNum
}

// 获取连接消息管理场
func (that *MsgGrp) GetSess() *MsgSess {
	return that.sess
}

// 获取或创建消息管理场
func (that *MsgGrp) getOrNewSess(force bool) *MsgSess {
	if that.sess == nil && force {
		that.locker.Lock()
		if that.sess == nil {
			sess := new(MsgSess)
			sess.grp = that
			that.sess = sess
		}

		that.locker.Unlock()
	}

	return that.sess
}

// 获取或创建db锁
func (that *MsgGrp) getOrNewDbLocker() sync.Locker {
	if that.dbLocker == nil {
		that.locker.Lock()
		if that.dbLocker == nil {
			that.dbLocker = new(sync.Mutex)
		}

		that.locker.Unlock()
	}

	return that.dbLocker
}

// 获取客户端
func (that *MsgGrp) getClient(unique string) *MsgClient {
	sess := that.sess
	if sess == nil {
		return nil
	}

	if unique == "" {
		return sess.client
	}

	clientMap := sess.clientMap
	if clientMap == nil {
		return nil
	}

	client, _ := clientMap.Load(unique)
	msgClient, _ := client.(*MsgClient)
	return msgClient
}

// 创建消息客户端
func (that *MsgGrp) newMsgClient(cid int64) *MsgClient {
	client := new(MsgClient)
	if _msgMng.OClientLocker {
		client.locker = new(sync.Mutex)

	} else {
		client.locker = that.locker
	}

	client.cid = cid
	client.gatewayI = Server.GetProdCid(cid).GetGWIClient()
	client.idleTime = time.Now().UnixNano() + _msgMng.IdleDrt
	return client
}

// 关闭消息客户端
func (that *MsgGrp) closeClient(client *MsgClient, cid int64, unique string, kick bool) bool {
	sess := that.sess
	if sess == nil {
		return false
	}

	if client == nil || (cid > 0 && cid != client.cid) {
		return false
	}

	if unique == "" {
		sess.client = nil

	} else if sess.clientMap != nil {
		sess.clientMap.Delete(unique)
	}

	client.connVer = 0
	sess.dirtyClientNum()
	if kick {
		// 关闭通知
		Util.GoSubmit(func() {
			client.gatewayI.Kick(Server.Context, &gw.KickReq{Cid: client.cid})
		})

		AZap.Debug("Grp Kick %s : %d, %s = %d", that.gid, cid, unique, sess.clientNum)

	} else {
		client.gatewayI.Close(Server.Context, &gw.CloseReq{Cid: client.cid})
		AZap.Debug("Grp Close %s : %d, %s = %d", that.gid, cid, unique, sess.clientNum)
	}

	return true
}

// 消息客户端检查
func (that *MsgGrp) checkClients() {
	sess := that.sess
	if sess == nil {
		return
	}

	clientNum := 0
	if sess.client != nil {
		that.checkClient(sess.client, "")
		clientNum += 1
	}

	if sess.clientMap != nil {
		sess.clientMap.RangeBuff(that.checkRange, &sess.checkBuff, Config.ClientPMax)
		clientNum += int(sess.clientMap.CountFast())
	}

	sess.clientNum = clientNum
}

func (that *MsgGrp) checkRange(key, val interface{}) bool {
	client, _ := val.(*MsgClient)
	unique, _ := key.(string)
	that.checkClient(client, unique)
	return true
}

func (that *MsgGrp) checkClient(client *MsgClient, unique string) {
	if client == nil {
		return
	}

	if client.idleTime > _msgMng.checkTime {
		// 未空闲
		return
	}

	if client.checking {
		return
	}

	limiter := Server.getLiveLimiter()
	if limiter == nil {
		that.checkClientRun(client, unique, nil)

	} else {
		Util.GoSubmit(func() {
			that.checkClientRun(client, unique, limiter)
		})
		limiter.Add()
	}
}

func (that *MsgGrp) checkClientOut(client *MsgClient) {
	client.checking = false
}

func (that *MsgGrp) checkClientRun(client *MsgClient, unique string, limiter Util.Limiter) {
	if limiter != nil {
		defer limiter.Done()
	}

	if client.checking {
		return
	}

	client.checking = true
	defer that.checkClientOut(client)
	rep, _ := client.gatewayI.Alive(Server.Context, client.getCidReq())
	ret := Server.Id32(rep)
	if ret < R_SUCC_MIN {
		that.closeClient(client, client.cid, unique, false)
	}
}

// 消息客户端连接
func (that *MsgGrp) Conn(cid int64, unique string, kick bool, newVer bool) *MsgClient {
	client := that.getClient(unique)
	if client != nil {
		if client.cid == cid {
			if newVer {
				client.connVer = _msgMng.newConnVer()
			}

			return client

		} else if client.cid > cid {
			return nil
		}

		that.closeClient(client, client.cid, unique, kick)
	}

	sess := that.getOrNewSess(true)
	client = that.newMsgClient(cid)
	client.connVer = _msgMng.newConnVer()

	if unique == "" {
		sess.client = client

	} else {
		sess.getOrNewClientMap().Store(unique, client)
	}

	sess.dirtyClientNum()
	AZap.Debug("Grp Conn %s : %d, %s = %d", that.gid, cid, unique, sess.clientNum)
	return client
}

// 消息客户端关闭
func (that *MsgGrp) Close(cid int64, unique string, connVer int32, kick bool) bool {
	client := that.getClient(unique)
	if client != nil && (connVer == 0 || connVer == client.connVer) {
		return that.closeClient(client, cid, unique, kick)
	}

	return false
}

// 消息客户端清理
func (that *MsgGrp) Clear(queue bool, last bool) {
	sess := that.sess
	if sess != nil {
		locked := false
		if queue && sess.queue != nil {
			if !locked {
				locked = true
				that.rwLocker.Lock()
			}

			sess.queue.Clear()
		}

		if last && sess.lastQueue != nil {
			if !locked {
				locked = true
				that.rwLocker.Lock()
			}

			sess.lastQueue.Clear()
			sess.lastQueueLoaded = false
		}

		if locked {
			that.rwLocker.Unlock()
		}
	}
}

// 消息客户端推送消息
func (that *MsgGrp) Push(uri string, data []byte, isolate bool, qs int32, queue bool, unique string, fid int64) (int64, bool, error) {
	AZap.Debug("Msg Push %s %s,%d", that.gid, uri, qs)
	if qs >= 2 {
		if _msgMng.LastMax <= 0 {
			qs = 1
		}
	}

	sess := that.getOrNewSess(queue)
	if qs <= 1 {
		if sess == nil {
			return 0, false, ERR_NOWAY
		}

		msg := NewMsg(uri, data, unique)
		succ, err := sess.QueuePush(msg)
		return msg.Get().Id, succ, err

	} else {
		if qs >= 0 && fid > 0 && _msgMng.Db != nil {
			// 唯一性校验
			id := _msgMng.Db.FidGet(fid, that.gid)
			if id > 0 {
				return id, true, nil
			}
		}

		var err error = nil
		msg := NewMsg(uri, data, unique)
		that.lastQueuePush(sess, msg, fid)
		msgD := msg.Get()
		if qs >= 3 && _msgMng.Db != nil {
			err = that.lastDbInsert(msgD)
		}

		// 消息直写测试
		if _msgMng.pushDrTest && sess != nil && sess.clientMap != nil {
			sTime := time.Now().UnixNano() / 1000000
			sess.clientMap.Range(func(key, value interface{}) bool {
				client := value.(*MsgClient)
				Util.GoSubmit(func() {
					sess.Push(msgD, client, key.(string), true)
				})
				return true
			})

			AZap.Logger.Debug("Msg PushDrTest span " + strconv.FormatInt(time.Now().UnixNano()/1000000-sTime, 10) + "ms")

		} else if sess != nil && msg.Get().Id > 0 {
			sess.LastStart()
		}

		return msg.Get().Id, true, err
	}
}

// 消息添加到last队列
func (that *MsgGrp) lastQueuePush(sess *MsgSess, msg Msg, fid int64) {
	msgD := msg.Get()
	msgD.Gid = that.gid
	msgD.Fid = fid

	// 顺序队列
	if sess == nil || sess.getOrNewLastQueue() == nil {
		return
	}

	// last队列载入
	sess.lastQueueLoad()
	// 锁加入队列
	that.rwLocker.Lock()
	msgD.Id = _msgMng.idWorkder.Generate()
	sess.lastQueue.Push(msg, true)
	that.rwLocker.Unlock()
}

// 消息插入到DB
func (that *MsgGrp) lastDbInsert(msgD *MsgD) error {
	if msgD.Id > 0 {
		// sess.lastQueue加强不漏消息
		return _msgMng.Db.Insert(msgD)

	} else {
		// db锁加强消息顺序写入
		dbLocker := that.getOrNewDbLocker()
		dbLocker.Lock()
		if that.sess != nil && that.sess.lastQueue != nil {
			dbLocker.Unlock()
			// 插入到队列
			that.lastQueuePush(that.sess, msgD, msgD.Id)
			return _msgMng.Db.Insert(msgD)
		}

		msgD.Id = _msgMng.idWorkder.Generate()
		err := _msgMng.Db.Insert(msgD)
		dbLocker.Unlock()
		return err
	}
}
