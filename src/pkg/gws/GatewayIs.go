package gws

import (
	"axj/ANet"
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axjGW/pkg/gateway"
	"context"
	"gw"
)

type GatewayIs struct {
}

func (g GatewayIs) Close(ctx context.Context, cid int64, reason string) (_r gw.Result_, _err error) {
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

func (g GatewayIs) Kick(ctx context.Context, cid int64, bytes []byte) (_r gw.Result_, _err error) {
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

func (g GatewayIs) Conn(ctx context.Context, cid int64, gid string, unique string) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHash(Kt.HashCode(KtUnsafe.StringToBytes(gid))) {
		return gw.Result__ProdErr, nil
	}

	if gateway.MsgMng.GetMsgGrp(gid).Conn(cid, unique) != nil {
		return gw.Result__Succ, nil
	}

	return gw.Result__Fail, nil
}

func (g GatewayIs) Disc(ctx context.Context, cid int64, gid string, unique string, connVer int32) (_err error) {
	if !gateway.Server.IsProdHash(Kt.HashCode(KtUnsafe.StringToBytes(gid))) {
		return nil
	}

	if gateway.MsgMng.GetMsgGrp(gid).Close(cid, unique, connVer) {
		return nil
	}

	return nil
}

func (g GatewayIs) Alive(ctx context.Context, cid int64) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	return gw.Result__Succ, nil
}

func (g GatewayIs) Rid(ctx context.Context, cid int64, name string, rid int32) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	gateway.Server.Handler.ClientG(client).PutRId(name, rid)
	return gw.Result__Succ, nil
}

func (g GatewayIs) Rids(ctx context.Context, cid int64, rids map[string]int32) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	gateway.Server.Handler.ClientG(client).PutRIds(rids)
	return gw.Result__Succ, nil
}

func (g GatewayIs) Last(ctx context.Context, cid int64, gid string, connVer int32) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	if gid == gateway.Server.Handler.ClientG(client).Gid() {
		gid = ""
	}

	err := client.Get().Rep(true, ANet.REQ_LAST, gid, connVer, nil, false, false, 0)
	if err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIs) Push(ctx context.Context, cid int64, uri string, bytes []byte, isolate bool, id int64) (_r gw.Result_, _err error) {
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

	err := client.Get().Rep(true, ANet.REQ_PUSH, uri, 0, bytes, false, isolate, id)
	if err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIs) GQueue(ctx context.Context, gid string, cid int64, unique string, clear bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHash(Kt.HashCode(KtUnsafe.StringToBytes(gid))) {
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

func (g GatewayIs) GClear(ctx context.Context, gid string, queue bool, last bool) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHash(Kt.HashCode(KtUnsafe.StringToBytes(gid))) {
		return gw.Result__ProdErr, nil
	}

	grp := gateway.MsgMng.MsgGrp(gid)
	if grp != nil {
		grp.Clear(queue, last)
	}

	return gw.Result__Succ, nil
}

func (g GatewayIs) GLasts(ctx context.Context, gid string, cid int64, unique string, lastId int64) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHash(Kt.HashCode(KtUnsafe.StringToBytes(gid))) {
		return gw.Result__ProdErr, nil
	}

	grp := gateway.MsgMng.GetMsgGrp(gid)
	client := grp.Conn(cid, unique)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	// 开始拉取
	go grp.Sess().Lasts(lastId, client, unique)
	return gw.Result__Succ, nil
}

func (g GatewayIs) GPush(ctx context.Context, gid string, uri string, bytes []byte, isolate bool, qs int32, queue bool, unique string, fid int64) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdHash(Kt.HashCode(KtUnsafe.StringToBytes(gid))) {
		return gw.Result__ProdErr, nil
	}

	grp := gateway.MsgMng.GetMsgGrp(gid)
	_, succ, err := grp.Push(uri, bytes, isolate, qs, queue, unique, fid)
	if succ {
		return gw.Result__Succ, nil
	}

	return gw.Result__Fail, err
}

func (g GatewayIs) GDirty(ctx context.Context, gid string) (_r gw.Result_, _err error) {
	panic("implement me")
}
