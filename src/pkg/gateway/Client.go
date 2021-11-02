package gateway

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
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
	hash     int       // hash值
	rid      int32     // 请求编号
	ridMap   *sync.Map // 请求字典
	connTime int64     // 最后连接时间
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

func (that *ClientG) SetId(uid int64, sid string, unique string) {
	that.uid = uid
	that.sid = sid
	if uid > 0 {
		that.gid = strconv.FormatInt(uid, 10)

	} else {
		that.gid = sid
	}

	that.unique = unique
	that.hash = -1
}

func (that *ClientG) Hash() int {
	if that.hash < 0 {
		clientC := that.Get()
		clientC.Locker().Lock()
		defer clientC.Locker().Unlock()
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
		defer clientC.Locker().Unlock()
		if that.ridMap == nil {
			that.ridMap = new(sync.Map)
		}
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

func (that *ClientG) GetProd(name string, rand bool) *Prod {
	prods := Server.GetProds(name)
	if prods == nil {
		return nil
	}

	id := that.GetId(name)
	if id > 0 {
		return prods.GetProd(id)
	}

	if rand {
		return prods.GetProdRand()
	}

	return prods.GetProdHash(that.Hash())
}

func (that *ClientG) ConnKeep() {
	that.connTime = time.Now().Unix() + Config.ConnDrt
}

func (that *ClientG) ConnCheck() {
	result, err := Server.GetProdClient(that).GetGWIClient().Conn(Server.Context, that.Id(), that.sid, that.unique)
	if result != gw.Result__Succ {
		// 用户注册失败
		that.Close(err, result)
	}
}

func (that *ClientG) ConnStart(clear bool, lasts bool) {
}
