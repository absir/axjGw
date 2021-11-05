package gws

import (
	"axjGW/gen/gw"
	"axjGW/pkg/gateway"
	"context"
)

type GatewayS struct {
}

func (g GatewayS) B(ctx context.Context) (_r bool, _err error) {
	return true, nil
}

func (g GatewayS) Close(ctx context.Context, cid int64, reason string) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().Close(ctx, cid, reason)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) Kick(ctx context.Context, cid int64, bytes []byte) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().Kick(ctx, cid, bytes)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) Rid(ctx context.Context, cid int64, name string, rid int32) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().Rid(ctx, cid, name, rid)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) Rids(ctx context.Context, cid int64, rids map[string]int32) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().Rids(ctx, cid, rids)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) Push(ctx context.Context, cid int64, uri string, bytes []byte) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().Push(ctx, cid, uri, bytes, false, 0)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) GLast(ctx context.Context, gid string) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdGid(gid).GetGWIClient().GLast(ctx, gid)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) GPush(ctx context.Context, gid string, uri string, bytes []byte, qs int32, unique string, queue bool) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdGid(gid).GetGWIClient().GPush(ctx, gid, uri, bytes, false, qs, queue, unique, 0)
	return ret >= gateway.R_SUCC_MIN, err
}

func (g GatewayS) GConn(ctx context.Context, cid int64, gid string, unique string) (_r int32, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().Conn(ctx, cid, gid, unique, false)
	return ret, err
}

func (g GatewayS) GDisc(ctx context.Context, cid int64, gid string, unique string, connVer int32) (_r bool, _err error) {
	err := gateway.Server.GetProdCid(cid).GetGWIClient().Disc(ctx, cid, gid, unique, connVer, false)
	return err == nil, err
}

func (g GatewayS) GLasts(ctx context.Context, gid string, cid int64, unique string, lastId int64, continuous int32) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdCid(cid).GetGWIClient().GLasts(ctx, gid, cid, unique, lastId, continuous)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) Send(ctx context.Context, fromId string, toId string, uri string, bytes []byte, db bool) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdGid(fromId).GetGWIClient().Send(ctx, fromId, toId, uri, bytes, db)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) TPush(ctx context.Context, fromId string, tid string, readfeed bool, uri string, bytes []byte, db bool, queue bool) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdGid(tid).GetGWIClient().TPush(ctx, fromId, tid, readfeed, uri, bytes, db, queue)
	return ret == gw.Result__Succ, err
}

func (g GatewayS) TDirty(ctx context.Context, tid string) (_r bool, _err error) {
	ret, err := gateway.Server.GetProdGid(tid).GetGWIClient().TDirty(ctx, tid)
	return ret == gw.Result__Succ, err
}
