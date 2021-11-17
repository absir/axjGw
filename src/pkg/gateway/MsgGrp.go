package gateway

import (
	"axj/Thrd/AZap"
	"axjGW/gen/gw"
	"sync"
	"time"
)

type MsgGrp struct {
	// 消息组编号
	gid string
	// 消息组hash
	ghash int
	// 消息组读写锁
	locker *sync.RWMutex
	// 消息组db锁
	dbLocker sync.Locker
	// 过期时间
	passTime int64
	// 消息组场
	sess *MsgSess
}

// 延长过期时间
func (that *MsgGrp) retain() {
	that.passTime = time.Now().UnixNano() + MsgMng.LiveDrt
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
		defer that.locker.Unlock()
		if that.sess == nil {
			sess := new(MsgSess)
			sess.grp = that
			that.sess = sess
		}
	}

	return that.sess
}

// 获取或创建db锁
func (that *MsgGrp) getOrNewDbLocker() sync.Locker {
	if that.dbLocker == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.dbLocker == nil {
			that.dbLocker = new(sync.Mutex)
		}
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
	client.cid = cid
	client.gatewayI = Server.GetProdCid(cid).GetGWIClient()
	client.idleTime = time.Now().UnixNano() + that.passTime
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

	} else {
		sess.clientMap.Delete(unique)
	}

	client.connVer = 0
	sess.clientNum--
	if kick {
		// 关闭通知
		go client.gatewayI.Kick(Server.Context, &gw.KickReq{Cid: client.cid}, nil)
	}

	AZap.Debug("Msg Close %s : %d, %s = %d", that.gid, cid, unique, sess.clientNum)
	return true
}

// 消息客户端检查
func (that *MsgGrp) checkClients() {
	sess := that.sess
	if sess == nil {
		return
	}

	that.checkClient(sess.client, "")
	if sess.clientMap != nil {
		sess.clientMap.Range(that.checkRange)
	}
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

	if client.idleTime > MsgMng.checkTime {
		// 未空闲
		return
	}

	rep, _ := client.gatewayI.Alive(Server.Context, client.getCidReq())
	ret := Server.Id32(rep)
	if ret < R_SUCC_MIN {
		that.locker.Lock()
		defer that.locker.Unlock()
		that.closeClient(client, client.cid, unique, false)
	}
}

// 消息客户端连接
func (that *MsgGrp) Conn(cid int64, unique string, kick bool, newVer bool) *MsgClient {
	client := that.getClient(unique)
	if client != nil {
		if client.cid == cid {
			if newVer {
				client.connVer = MsgMng.newConnVer()
			}

			return client

		} else if client.cid > cid {
			return nil
		}

		that.closeClient(client, client.cid, unique, kick)
	}

	sess := that.getOrNewSess(true)
	client = that.newMsgClient(cid)
	client.connVer = MsgMng.newConnVer()
	if unique == "" {
		sess.client = client

	} else {
		if sess.clientMap == nil {
			sess.clientMap = new(sync.Map)
		}

		sess.clientMap.Store(unique, client)
	}

	sess.clientNum++
	AZap.Debug("Msg Conn %s : %d, %s = %d", that.gid, cid, unique, sess.clientNum)
	return client
}

// 消息客户端关闭
func (that *MsgGrp) Close(cid int64, unique string, connVer int32, kick bool) bool {
	that.locker.Lock()
	defer that.locker.Unlock()
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
				that.locker.Lock()
				defer that.locker.Unlock()
			}

			sess.queue.Clear()
		}

		if last && sess.lastQueue != nil {
			if !locked {
				locked = true
				that.locker.Lock()
				defer that.locker.Unlock()
			}

			sess.lastQueue.Clear()
			sess.lastQueueLoaded = false
		}
	}
}

// 消息客户端推送消息
func (that *MsgGrp) Push(uri string, data []byte, isolate bool, qs int32, queue bool, unique string, fid int64) (int64, bool, error) {
	AZap.Debug("Msg Push %s %s,%d", that.gid, uri, qs)
	if qs >= 2 {
		if MsgMng.LastMax <= 0 {
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
		if qs >= 0 && fid > 0 && MsgMng.Db != nil {
			// 唯一性校验
			id := MsgMng.Db.FidGet(fid, that.gid)
			if id > 0 {
				return id, true, nil
			}
		}

		var err error = nil
		msg := NewMsg(uri, data, unique)
		that.lastQueuePush(sess, msg, fid)
		msgD := msg.Get()
		if qs >= 3 && MsgMng.Db != nil {
			err = that.lastDbInsert(msgD)
		}

		if sess != nil && msg.Get().Id > 0 {
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
	that.locker.Lock()
	defer that.locker.Unlock()
	msgD.Id = MsgMng.idWorkder.Generate()
	sess.lastQueue.Push(msg, true)
}

// 消息插入到DB
func (that *MsgGrp) lastDbInsert(msgD *MsgD) error {
	if msgD.Id > 0 {
		// sess.lastQueue加强不漏消息
		return MsgMng.Db.Insert(msgD)

	} else {
		// db锁加强消息顺序写入
		that.getOrNewDbLocker().Lock()
		defer that.getOrNewDbLocker().Unlock()
		if that.sess != nil && that.sess.lastQueue != nil {
			// 插入到队列
			that.lastQueuePush(that.sess, msgD, msgD.Id)
			return MsgMng.Db.Insert(msgD)
		}

		msgD.Id = MsgMng.idWorkder.Generate()
		return MsgMng.Db.Insert(msgD)
	}
}
