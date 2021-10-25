package gws

import (
	"axj/ANet"
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

func (g GatewayIs) Conn(ctx context.Context, cid int64, sid string, unique string) (_r gw.Result_, _err error) {
	panic("implement me")
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

func (g GatewayIs) Last(ctx context.Context, cid int64) (_r gw.Result_, _err error) {
	if !gateway.Server.IsProdCid(cid) {
		return gw.Result__ProdErr, nil
	}

	client := gateway.Server.Manager.Client(cid)
	if client == nil {
		return gw.Result__IdNone, nil
	}

	err := client.Get().Rep(true, ANet.REQ_LAST, "", 0, nil, false, false)
	if err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIs) Push(ctx context.Context, cid int64, uri string, bytes []byte, isolate bool) (_r gw.Result_, _err error) {
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

	err := client.Get().Rep(true, ANet.REQ_PUSH, "", 0, bytes, false, isolate)
	if err != nil {
		return gw.Result__Fail, err
	}

	return gw.Result__Succ, nil
}

func (g GatewayIs) Dirty(ctx context.Context, sid string) (_r gw.Result_, _err error) {
	panic("implement me")
}
