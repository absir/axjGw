package gws

import (
	"axj/ANet"
	"axjGW/gen/gw"
	"axjGW/pkg/gateway"
	"context"
)

type GatewayIS struct {
}

func (g GatewayIS) Close(ctx context.Context, cid int64, reason string) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	client.Get().Close(nil, reason)
	return gw.Result__Succ, nil
}

func (g GatewayIS) Kick(ctx context.Context, cid int64, bytes []byte) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	client.Get().Kick(bytes, false, gateway.Config.KickDrt)
	return gw.Result__Succ, nil
}

func (g GatewayIS) Conn(ctx context.Context, cid int64, gid string, unique string) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return gw.Result__ProdErr, nil
	}

	if gateway.MsgMng.GetMsgGrp(gid).Conn(cid, unique) != nil {
		return gw.Result__Succ, nil
	}

	return gw.Result__Fail, nil
}

func (g GatewayIS) Disc(ctx context.Context, cid int64, gid string, unique string, connVer int32) (_err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return nil
	}

	if gateway.MsgMng.GetMsgGrp(gid).Close(cid, unique, connVer) {
		return nil
	}

	return nil
}

func (g GatewayIS) Alive(ctx context.Context, cid int64) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) Rid(ctx context.Context, cid int64, name string, rid int32) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	gateway.Handler.ClientG(client).PutRId(name, rid)
	return gw.Result__Succ, nil
}

func (g GatewayIS) Rids(ctx context.Context, cid int64, rids map[string]int32) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	gateway.Handler.ClientG(client).PutRIds(rids)
	return gw.Result__Succ, nil
}

func (g GatewayIS) Last(ctx context.Context, cid int64, gid string, connVer int32, continuous bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	if gid == gateway.Handler.ClientG(client).Gid() {
		gid = ""
	}

	var err error
	if continuous {
		err = client.Get().Rep(true, ANet.REQ_LASTC, gid, connVer, nil, false, false, 0)

	} else {
		err = client.Get().Rep(true, ANet.REQ_LAST, gid, connVer, nil, false, false, 0)
	}

	if err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) Push(ctx context.Context, cid int64, uri string, bytes []byte, isolate bool, id int64) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	if ctx != gateway.Server.Context {
		isolate = false
	}

	req := ANet.REQ_PUSH
	if id > 0 {
		req = ANet.REQ_PUSHI
	}

	err := client.Get().Rep(true, req, uri, 0, bytes, false, isolate, id)
	if err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) GQueue(ctx context.Context, gid string, cid int64, unique string, clear bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return gw.Result__ProdErr, nil
	}

	grp := gateway.MsgMng.GetMsgGrp(gid)
	client := grp.Conn(cid, unique)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	if clear {
		grp.Clear(true, false)

	} else {
		grp.Sess().QueueStart()
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) GClear(ctx context.Context, gid string, queue bool, last bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return gw.Result__ProdErr, nil
	}

	grp := gateway.MsgMng.MsgGrp(gid)
	if grp != nil {
		grp.Clear(queue, last)
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) GLasts(ctx context.Context, gid string, cid int64, unique string, lastId int64, continuous bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return gw.Result__ProdErr, nil
	}

	grp := gateway.MsgMng.GetMsgGrp(gid)
	client := grp.Conn(cid, unique)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	// 开始拉取
	grp.Sess().Lasts(lastId, client, unique, continuous)
	return gw.Result__Succ, nil
}

func (g GatewayIS) GPush(ctx context.Context, gid string, uri string, bytes []byte, isolate bool, qs int32, queue bool, unique string, fid int64) (_r int64, _err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return int64(gw.Result__ProdErr), nil
	}

	grp := gateway.MsgMng.GetMsgGrp(gid)
	id, succ, err := grp.Push(uri, bytes, isolate, qs, queue, unique, fid)
	if succ {
		return id, nil
	}

	return int64(gw.Result__Fail), err
}

func (g GatewayIS) GPushA(ctx context.Context, gid string, id int64, succ bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(gid) {
		return gw.Result__ProdErr, nil
	}

	var err error = nil
	if succ {
		err = gateway.ChatMng.MsgSucc(id)

	} else {
		err = gateway.ChatMng.MsgFail(id, gid)
	}
	if err == nil {
		return gw.Result__Succ, nil
	}

	return gw.Result__Fail, err
}

func (g GatewayIS) Send(ctx context.Context, fromId string, toId string, uri string, bytes []byte, db bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(fromId) {
		return gw.Result__ProdErr, nil
	}

	succ, err := gateway.ChatMng.Send(fromId, toId, uri, bytes, db)
	if !succ || err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) TPush(ctx context.Context, fromId string, tid string, readfeed bool, uri string, bytes []byte, db bool, queue bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(tid) {
		return gw.Result__ProdErr, nil
	}

	succ, err := gateway.ChatMng.TeamPush(fromId, tid, readfeed, uri, bytes, queue, db)
	if !succ || err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIS) TDirty(ctx context.Context, tid string) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(tid) {
		return gw.Result__ProdErr, nil
	}

	gateway.TeamMng.Dirty(tid)
	return gw.Result__Fail, nil
}

func (g GatewayIS) TStarts(ctx context.Context, tid string) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHashS(tid) {
		return gw.Result__ProdErr, nil
	}

	gateway.ChatMng.TeamStart(tid, nil)
	return gw.Result__Succ, nil
}
