package gws

import (
	"axj/ANet"
	"axjGW/gen/gw"
	"axjGW/pkg/gateway"
	"context"
	"errors"
)

var Result_ProdErr error
var Result_True *gw.BoolRep
var Result_Fasle *gw.BoolRep
var Result_Fail_Rep *gw.Id32Rep
var Result_IdNone_Rep *gw.Id32Rep
var Result_ProdErr_Rep *gw.Id32Rep
var Result_Succ_Rep *gw.Id32Rep
var Result_Fail_Rep64 *gw.Id64Rep
var Result_IdNone_Rep64 *gw.Id64Rep
var Result_ProdErr_Rep64 *gw.Id64Rep
var Result_Succ_Rep64 *gw.Id64Rep

func init() {
	Result_ProdErr = errors.New("ProdErr")
	Result_True = &gw.BoolRep{
		Val: true,
	}
	Result_Fasle = &gw.BoolRep{
		Val: false,
	}
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

func (g GatewayIs) Uid(ctx context.Context, req *gw.CidReq) (*gw.UIdRep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return nil, Result_ProdErr
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return nil, nil
	}

	clientG := gateway.Handler.ClientG(client)
	return clientG.UidRep(), nil
}

func (g GatewayIs) Online(ctx context.Context, req *gw.GidReq) (*gw.BoolRep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return nil, Result_ProdErr
	}

	msgGrp := gateway.MsgMng().GetMsgGrp(req.Gid)
	if msgGrp == nil || msgGrp.ClientNum() <= 0 {
		return Result_Fasle, nil
	}

	return Result_True, nil
}

func (g GatewayIs) Onlines(ctx context.Context, req *gw.GidsReq) (*gw.BoolsRep, error) {
	gLen := 0
	if req.Gids != nil {
		gLen = len(req.Gids)
	}

	if gLen <= 0 {
		return nil, nil
	}

	boolsRep := &gw.BoolsRep{}
	boolsRep.Vals = make([]bool, gLen)
	for i := 0; i < gLen; i++ {
		gid := req.Gids[i]
		if !gateway.Server.IsProdHashS(gid) {
			return nil, Result_ProdErr
		}

		msgGrp := gateway.MsgMng().GetMsgGrp(gid)
		if msgGrp == nil || msgGrp.ClientNum() <= 0 {
			continue
		}

		boolsRep.Vals[i] = true
	}

	return boolsRep, nil
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

func (g GatewayIs) CidGid(ctx context.Context, req *gw.CidGidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	gateway.Handler.ClientG(client).CidGid(req)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) GidCid(ctx context.Context, req *gw.CidGidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng().GetMsgGrp(req.Gid)
	if grp == nil || !grp.CheckConn(req.Cid, req.Unique, req.State == gw.GidState_GLast) {
		return Result_Fail_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Conn(ctx context.Context, req *gw.GConnReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.MsgMng().GetOrNewMsgGrp(req.Gid).Conn(req.Cid, req.Unique, req.Kick, req.NewVer, true)
	if client == nil {
		return Result_Fail_Rep, nil
	}

	return &gw.Id32Rep{
		Id: client.GetConnVer(),
	}, nil
}

func (g GatewayIs) Disc(ctx context.Context, req *gw.GDiscReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	if !gateway.MsgMng().GetOrNewMsgGrp(req.Gid).Close(req.Cid, req.Unique, req.ConnVer, true) {
		return Result_Fail_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Rep(ctx context.Context, req *gw.RepReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdCid(req.Cid) {
		return Result_ProdErr_Rep, nil
	}

	client := gateway.Server.Manager.Client(req.Cid)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	if ctx != gateway.Server.Context {
		req.Isolate = false
	}

	err := client.Get().RepCData(true, req.Req, req.Uri, req.UriI, req.Data, req.CDid, req.Isolate, req.Encry, 0)
	if err != nil {
		return Result_Fail_Rep, err
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

	err := client.Get().RepCData(true, reqI, req.Uri, 0, req.Data, req.CData, req.Isolate, true, req.Id)
	if err != nil {
		return Result_Fail_Rep, err
	}

	return Result_Succ_Rep, nil
}

/**
 * 单用户登录 client 建立连接 gid, 消息队列清理开启
 */
func (g GatewayIs) GQueue(ctx context.Context, req *gw.IGQueueReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng().GetOrNewMsgGrp(req.Gid)
	client := grp.Conn(req.Cid, req.Unique, false, false, true)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	if req.Clear {
		grp.Clear(true, false)

	} else {
		grp.GetSess().QueueStart()
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) GClear(ctx context.Context, req *gw.IGClearReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng().GetMsgGrp(req.Gid)
	if grp != nil {
		grp.Clear(req.Queue, req.Last)
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) GLasts(ctx context.Context, req *gw.GLastsReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng().GetOrNewMsgGrp(req.Gid)
	client := grp.GetClient(req.Unique)
	if client == nil || client.GetSubLastId() == 0 {
		rep, err := gateway.Server.GetProdCid(req.Cid).GetGWIClient().CidGid(gateway.Server.Context, &gw.CidGidReq{Cid: req.Cid, Gid: req.Gid, Unique: req.Unique, State: gw.GidState_GLast})
		if !gateway.Server.Id32Succ(gateway.Server.Id32(rep)) {
			return rep, err
		}
	}

	client = grp.Conn(req.Cid, req.Unique, false, true, false)
	if client == nil {
		return Result_IdNone_Rep, nil
	}

	// 开始拉取
	grp.GetSess().SubLast(req.LastId, client, req.Continuous)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) GLast(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng().GetMsgGrp(req.Gid)
	if grp != nil {
		sess := grp.GetSess()
		if sess != nil {
			sess.LastStart()
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

	grp := gateway.MsgMng().GetOrNewMsgGrp(req.Gid)
	id, succ, _ := grp.Push(req.Uri, req.Data, req.Isolate, req.Qs, req.Queue, req.Unique, req.Fid)
	if !succ {
		return Result_Fail_Rep64, nil
	}

	if id <= Result_Succ_Rep64.Id {
		return Result_Succ_Rep64, nil
	}

	return &gw.Id64Rep{Id: id}, nil
}

func (g GatewayIs) GPushA(ctx context.Context, req *gw.IGPushAReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	var err error = nil
	if req.Succ {
		err = gateway.ChatMng().MsgSucc(req.Id)

	} else {
		err = gateway.ChatMng().MsgFail(req.Id, req.Gid)
	}

	if err != nil {
		return Result_Fail_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Send(ctx context.Context, req *gw.SendReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.FromId) {
		return Result_ProdErr_Rep, nil
	}

	succ, err := gateway.ChatMng().Send(req)
	if !succ || err != nil {
		return Result_Fail_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) TPush(ctx context.Context, req *gw.TPushReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.FromId) {
		return Result_ProdErr_Rep, nil
	}

	succ, err := gateway.ChatMng().TeamPush(req)
	if !succ || err != nil {
		return Result_Fail_Rep, nil
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) TDirty(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	gateway.TeamMng().Dirty(req.Gid)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) TStarts(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	gateway.ChatMng().TeamStart(req.Gid, nil)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) SetProds(ctx context.Context, rep *gw.ProdsRep) (*gw.BoolRep, error) {
	gateway.Server.SetProdsRep(rep)
	return Result_True, nil
}

func (g GatewayIs) Read(ctx context.Context, req *gw.ReadReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	var grp *gateway.MsgGrp = nil
	grp = gateway.MsgMng().GetMsgGrp(req.Gid)
	if grp != nil && grp.GetSess() != nil {
		grp.GetSess().ReadLastId(req.Tid, req.LastId)

	} else {
		// 数据库保存
		if gateway.MsgMng().Db != nil {
			gateway.MsgMng().Db.Read(gateway.MsgMng().GidForTid(req.Gid, req.Tid), req.LastId)
		}
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) Unread(ctx context.Context, req *gw.UnreadReq) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(req.Gid) {
		return Result_ProdErr_Rep, nil
	}

	var grp *gateway.MsgGrp = nil
	if req.Num > 0 {
		grp = gateway.MsgMng().GetOrNewMsgGrp(req.Gid)

	} else {
		grp = gateway.MsgMng().GetMsgGrp(req.Gid)
	}

	if grp == nil {
		return Result_Fail_Rep, nil
	}

	// 未读消息
	grp.GetOrNewSess(true).UnreadRecv(req.Tid, req.Num, req.LastId, req.Uri, req.Data, req.Entry)
	return Result_Succ_Rep, nil
}

func (g GatewayIs) Unreads(ctx context.Context, reqs *gw.UnreadReqs) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(reqs.Gid) {
		return Result_ProdErr_Rep, nil
	}

	grp := gateway.MsgMng().GetOrNewMsgGrp(reqs.Gid)
	if grp == nil {
		return Result_Fail_Rep, nil
	}

	if reqs.Reqs != nil {
		sess := grp.GetOrNewSess(true)
		for _, req := range reqs.Reqs {
			// 未读消息
			sess.UnreadRecv(req.Tid, req.Num, req.LastId, req.Uri, req.Data, req.Entry)
		}
	}

	return Result_Succ_Rep, nil
}

func (g GatewayIs) UnreadTids(ctx context.Context, tids *gw.UnreadTids) (*gw.Id32Rep, error) {
	if !gateway.Server.IsProdHashS(tids.Gid) {
		return Result_ProdErr_Rep, nil
	}

	gateway.MsgMng().UnreadTids(tids.Gid, tids.Tids)
	return Result_Succ_Rep, nil
}
