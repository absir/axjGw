package gws

import (
	"axj/ANet"
	"axjGW/gen/gw"
	"axjGW/pkg/gateway"
	"context"
)

var Result_Fail_Rep *gw.Id32Rep
var Result_IdNone_Rep *gw.Id32Rep
var Result_ProdErr_Rep *gw.Id32Rep
var Result_Succ_Rep *gw.Id32Rep
var Result_Fail_Rep64 *gw.Id64Rep
var Result_IdNone_Rep64 *gw.Id64Rep
var Result_ProdErr_Rep64 *gw.Id64Rep
var Result_Succ_Rep64 *gw.Id64Rep

func init() {
	Result_Fail_Rep = &gw.Id32Rep{
		Id: int32(gw.Result_Fail),
	}
	Result_IdNone_Rep = &gw.Id32Rep{
		Id: int32(gw.Result_IdNone),
	}
	Result_ProdErr_Rep = &gw.Id32Rep{
		Id: int32(gw.Result_ProdErr),
	}
	Result_Succ_Rep = &gw.Id32Rep{
		Id: int32(gw.Result_Succ),
	}
	Result_Fail_Rep64 = &gw.Id64Rep{
		Id: int64(gw.Result_Fail),
	}
	Result_IdNone_Rep64 = &gw.Id64Rep{
		Id: int64(gw.Result_IdNone),
	}
	Result_ProdErr_Rep64 = &gw.Id64Rep{
		Id: int64(gw.Result_ProdErr),
	}
	Result_Succ_Rep64 = &gw.Id64Rep{
		Id: int64(gw.Result_Succ),
	}
}

type GatewayIs struct {
}

func (g GatewayIs) Close(ctx context.Context, req *gw.CloseReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	client.Get().Close(nil, req.Reason)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) Kick(ctx context.Context, req *gw.KickReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	client.Get().Kick(req.Data, false, 0)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) Alive(ctx context.Context, req *gw.CidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Rid(ctx context.Context, req *gw.RidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	gateway.Handler.ClientG(client).PutRId(req.Name, req.Rid)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) Rids(ctx context.Context, req *gw.RidsReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	gateway.Handler.ClientG(client).PutRIds(req.Rids)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) Conn(ctx context.Context, req *gw.GConnReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.MsgMng.GetMsgGrp(req.Gid).Conn(req.Cid, req.Unique, req.Kick, req.NewVer)
	if client == nil {
		return Result_Fail_Rep, nil
	}

	return &gw.Id32Rep{
		Id: client.ConnVer(),
	}, nil
}

func (g GatewayIs) Disc(ctx context.Context, req *gw.GDiscReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	if !gateway.MsgMng.GetMsgGrp(req.Gid).Close(req.Cid, req.Unique, req.ConnVer, req.Kick) {
		return Result_Fail_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Last(ctx context.Context, req *gw.ILastReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	if req.Gid == gateway.Handler.ClientG(client).Gid() {
		req.Gid = ""
	}

	var err error
	if req.Continuous {
		err = client.Get().Rep(true, ANet.REQ_LASTC, req.Gid, req.ConnVer, nil, false, false, 0)

	} else {
		err = client.Get().Rep(true, ANet.REQ_LAST, req.Gid, req.ConnVer, nil, false, false, 0)
	}

	if err != nil {
		return Result_Fail_Rep, err
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Push(ctx context.Context, req *gw.PushReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	reqI := ANet.REQ_PUSH
	if req.Id > 0 {
		reqI = ANet.REQ_PUSHI
	}

	if ctx != gateway.Server.Context {
		req.Isolate = false
	}

	err := client.Get().Rep(true, reqI, req.Uri, 0, req.Data, req.Isolate, true, req.Id)
	if err != nil {
		return Result_Fail_Rep, err
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) GQueue(ctx context.Context, req *gw.IGQueueReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng.GetMsgGrp(req.Gid)
	client := grp.Conn(req.Cid, req.Unique, false, false)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	if req.Clear {
		grp.Clear(true, false)

	} else {
		grp.Sess().QueueStart()
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) GClear(ctx context.Context, req *gw.IGClearReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng.MsgGrp(req.Gid)
	if grp != nil {
		grp.Clear(req.Queue, req.Last)
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) GLasts(ctx context.Context, req *gw.GLastsReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng.GetMsgGrp(req.Gid)
	client := grp.Conn(req.Cid, req.Unique, false, false)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	// 开始拉取
	grp.Sess().Lasts(req.LastId, client, req.Unique, req.Continuous)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) GLast(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng.MsgGrp(req.Gid)
	if grp != nil {
		sess := grp.Sess()
		if sess != nil {
			sess.LastsStart()
		}
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) GPush(ctx context.Context, req *gw.GPushReq) (*gw.Id64Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep64, nil
	}

	if ctx != gateway.Server.Context {
		req.Isolate = false
	}

	grp := gateway.MsgMng.GetMsgGrp(req.Gid)
	id, succ, err := grp.Push(req.Uri, req.Data, req.Isolate, req.Qs, req.Queue, req.Unique, req.Fid)
	if !succ {
		return Result_Fail_Rep64, err
	}

	return &gw.Id64Rep{Id: id}, nil
}

func (g GatewayIs) GPushA(ctx context.Context, req *gw.IGPushAReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	var err error = nil
	if req.Succ {
		err = gateway.ChatMng.MsgSucc(req.Id)

	} else {
		err = gateway.ChatMng.MsgFail(req.Id, req.Gid)
	}

	if err != nil {
		return Result_Fail_Rep, err
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Send(ctx context.Context, req *gw.SendReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.FromId) {
		return Result_ProdErr_Rep, nil
	}

	succ, err := gateway.ChatMng.Send(req.FromId, req.ToId, req.Uri, req.Data, req.Db)
	if !succ || err != nil {
		return Result_Fail_Rep, err
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) TPush(ctx context.Context, req *gw.TPushReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.FromId) {
		return Result_ProdErr_Rep, nil
	}

	succ, err := gateway.ChatMng.TeamPush(req.FromId, req.Tid, req.ReadFeed, req.Uri, req.Data, req.Queue, req.Db)
	if !succ || err != nil {
		return Result_Fail_Rep, err
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) TDirty(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	gateway.TeamMng.Dirty(req.Gid)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) TStarts(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	gateway.ChatMng.TeamStart(req.Gid, nil)
	return Result_Succ_Rep, nil
}
