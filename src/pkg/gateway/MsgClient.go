package gateway

import (
	"axj/ANet"
	"axjGW/gen/gw"
	"sync"
)

type MsgClient struct {
	grp           *MsgGrp
	unique        string
	locker        sync.Locker
	cid           int64
	gatewayI      gw.GatewayIClient
	connVer       int32
	idleTime      int64
	lasting       bool
	lastTime      int64
	subLastTime   int64
	subLastId     int64
	subContinuous int32
	checking      bool //checkClient中
	cidReq        *gw.CidReq
	unreadsBuff   []interface{}
	unreadVer     int32
}

func (that *MsgClient) GetConnVer() int32 {
	return that.connVer
}

func (that *MsgClient) GetSubLastId() int64 {
	return that.subLastId
}

func (that *MsgClient) getCidReq() *gw.CidReq {
	if that.cidReq == nil {
		that.cidReq = &gw.CidReq{
			Cid: that.cid,
		}
	}

	return that.cidReq
}

func (that *MsgClient) unreadPush() {
	sess := that.grp.sess
	if sess == nil {
		return
	}

	// 未读消息推送
	unreads := sess.unreads
	if unreads != nil {
		unreads.RangeBuff(that.unreadPushRange, &that.unreadsBuff, 1024)
	}
}

func (that *MsgClient) unreadPushRange(key, val interface{}) bool {
	unread, _ := val.(*SessUnread)
	if unread == nil || unread.num <= 0 {
		return true
	}

	gid, _ := key.(string)
	if that.unreadVer == unread.ver {
		return true
	}

	rep, err := Server.GetProdCid(that.cid).GetGWIClient().Rep(Server.Context, &gw.RepReq{
		Cid:  that.cid,
		Req:  ANet.REQ_READ,
		Uri:  gid,
		UriI: unread.num,
	})

	if that.grp.sess.OnResult(rep, err, ER_PUSH, that) {
		that.unreadVer = unread.ver
	}

	return true
}
