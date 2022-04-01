package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"strconv"
	"sync"
	"time"
)

type ClientG struct {
	ANet.ClientMng
	uid      int64     // 用户编号int64
	sid      string    // 用户编号string
	gid      string    // 消息组编号
	unique   string    // 唯一标识
	discBack bool      // 断线回调
	hash     int       // hash值
	rid      int32     // 请求编号
	ridMap   *sync.Map // 请求字典
	connTime int64     // 最后连接时间
	conning  bool      // 连接中
	connReq  *gw.GConnReq
	uidRep   *gw.UIdRep
	gidConn  *GidConn
	gidMap   *cmap.CMap // gid状态
}

type GidConn struct {
	connTime int64         // 最后连接时间 update set delete locker
	req      *gw.CidGidReq // 连接请求
}

func (that *ClientG) Uid() int64 {
	return that.uid
}

func (that *ClientG) Sid() string {
	return that.sid
}

func (that *ClientG) Gid() string {
	return that.gid
}

func (that *ClientG) Unique() string {
	return that.unique
}

func (that *ClientG) SetId(uid int64, sid string, unique string, discBack bool) {
	that.uid = uid
	that.sid = sid
	if uid > 0 {
		that.gid = strconv.FormatInt(uid, 10)

	} else {
		that.gid = sid
	}

	that.unique = unique
	that.discBack = discBack
	that.hash = -1
}

func (that *ClientG) Hash() int {
	if that.hash < 0 {
		var hash int
		if that.gid != "" {
			hash = Kt.HashCode(KtUnsafe.StringToBytes(that.gid))

		} else if that.sid != "" {
			hash = Kt.HashCode(KtUnsafe.StringToBytes(that.sid))

		} else if that.uid > 0 {
			hash = int(that.uid)

		} else {
			hash = int(that.Id())
		}

		if hash < 0 {
			hash = -hash
		}

		that.hash = hash
	}

	return that.hash
}

func (that *ClientG) GetId(name string) int32 {
	if name == "" || name == Config.AclProd {
		return that.rid
	}

	if that.ridMap == nil {
		return 0
	}

	id, _ := that.ridMap.Load(name)
	if id == nil {
		return 0
	}

	return id.(int32)
}

func (that *ClientG) initRidMap() {
	if that.ridMap == nil {
		clientC := that.Get()
		clientC.Locker().Lock()
		if that.ridMap == nil {
			that.ridMap = new(sync.Map)
		}

		clientC.Locker().Unlock()
	}
}

func (that *ClientG) PutRId(name string, id int32) {
	if name == "" || name == Config.AclProd {
		that.rid = id
		return
	}

	if that.ridMap == nil {
		if id <= 0 {
			return
		}
	}

	that.initRidMap()
	if id <= 0 {
		that.ridMap.Delete(name)

	} else {
		that.ridMap.Store(name, id)
	}
}

func (that *ClientG) PutRIds(ids map[string]int32) {
	if ids == nil {
		return
	}

	for name, id := range ids {
		that.PutRId(name, id)
	}
}

func (that *ClientG) GetProd(name string, rand bool) (*Prod, *Prods) {
	prods := Server.GetProds(name)
	if prods == nil {
		return nil, prods
	}

	id := that.GetId(name)
	if id > 0 {
		return prods.GetProd(id), prods
	}

	if rand {
		return prods.GetProdRand(), prods
	}

	return prods.GetProdHash(that.Hash()), prods
}

func (that *ClientG) UidRep() *gw.UIdRep {
	if that.uidRep == nil {
		uidRep := &gw.UIdRep{}
		if that.uid > 0 {
			uidRep.Uid = that.uid
		}

		if that.sid != "" {
			uidRep.Sid = that.sid
		}
	}

	return that.uidRep
}

func (that *GidConn) connKeep() {
	that.connTime = time.Now().UnixNano() + Config.ConnDrt
}

func (that *ClientG) CidGid(req *gw.CidGidReq) {
	if req.Gid == that.gid {
		if req.State == gw.GidState_Disc {
			that.gidConn = nil
			return
		}

		conn := &GidConn{
			req: req,
		}

		conn.connKeep()
		that.gidConn = conn
		return
	}

	if that.gidMap == nil {
		if req.State == gw.GidState_Disc {
			return
		}
	}

	that.Locker().Lock()
	if that.gidMap == nil {
		that.gidMap = cmap.NewCMapInit()
	}

	that.Locker().Unlock()
	if req.State == gw.GidState_Disc {
		that.gidMap.Delete(that.gid)

	} else {
		conn := &GidConn{
			req: req,
		}

		conn.connKeep()
		that.gidMap.Store(that.gid, conn)
	}
}

func (that *ClientG) GidConnRange(key, val interface{}) bool {
	conn, _ := val.(*GidConn)
	if conn == nil {
		that.gidMap.Delete(key)

	} else {
		gid, _ := key.(string)
		that.GidConn(Server.Manager.CheckTime(), gid, conn)
	}

	return true
}

func (that *ClientG) GidConn(time int64, gid string, conn *GidConn) {
	if conn == nil || conn.connTime > time {
		return
	}

	conn.connKeep()
	limiter := Server.getConnLimiter()
	if limiter == nil {
		that.gidConnRun(limiter, gid, conn)

	} else {
		Util.GoSubmit(func() {
			that.gidConnRun(limiter, gid, conn)
		})

		limiter.Add()
	}
}

func (that *ClientG) getGidConn(gid string) *GidConn {
	if gid == that.gid {
		return that.gidConn
	}

	if that.gidMap == nil {
		return nil
	}

	conn, _ := that.gidMap.Load(gid)
	if conn == nil {
		return nil
	}

	return conn.(*GidConn)
}

func (that *ClientG) gidConnRun(limiter Util.Limiter, gid string, conn *GidConn) {
	if limiter != nil {
		defer limiter.Done()
	}

	rep, err := Server.GetProdGid(gid).GetGWIClient().GidCid(Server.Context, conn.req)
	if !Server.Id32Succ(Server.Id32(rep)) && conn == that.getGidConn(gid) {
		that.Close(err, rep)
	}
}
