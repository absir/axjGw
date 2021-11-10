package gws

import (
	"axjGW/gen/gw"
	"axjGW/pkg/gateway"
	"context"
)

type GatewayS struct {
}

func (g GatewayS) Uid(ctx context.Context, req *gw.CidReq) (*gw.UIdRep, error) {
	return gateway.Server.GetProdCid(req.Cid).GetGWIClient().Uid(ctx, req)
}

func (g GatewayS) Online(ctx context.Context, req *gw.GidReq) (*gw.BoolRep, error) {
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().Online(ctx, req)
}

func (g GatewayS) Onlines(ctx context.Context, req *gw.GidsReq) (*gw.BoolsRep, error) {
	gids := req.Gids
	gLen := 0
	if gids != nil {
		gLen = len(gids)
	}

	if gLen <= 0 {
		return nil, nil
	}

	nVals := make([]bool, gLen)
	nIdxs := make([]int, gLen)
	nGids := make([]string, gLen)
	pId := gateway.Server.GetProdGid(req.Gids[0]).Id()
	var pIdIdx int
	var pIdNext int32
	for {
		pIdIdx = 0
		pIdNext = 0
		for i := 0; i < gLen; i++ {
			gid := gids[i]
			id := gateway.Server.GetProdGid(gid).Id()
			if id == pId {
				// 同prod 合并请求
				nGids[pIdIdx] = gid
				nIdxs[pIdIdx] = i
				pIdIdx++

			} else if pIdIdx > 0 && pIdNext == 0 {
				// 下一个遍历prodId
				pIdNext = id
			}
		}

		if pIdIdx <= 0 || pIdNext == 0 {
			break
		}

		req.Gids = nGids[0:pIdIdx]
		boolsRep, err := gateway.Server.GetProdId(pId).GetGWIClient().Onlines(ctx, req)
		if err != nil {
			return nil, err
		}

		// 开始赋值
		for i := 0; i < pIdIdx; i++ {
			nVals[nIdxs[i]] = boolsRep.Vals[i]
		}

		pId = pIdNext
	}

	return &gw.BoolsRep{
		Vals: nVals,
	}, nil
}

func (g GatewayS) Close(ctx context.Context, req *gw.CloseReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdCid(req.Cid).GetGWIClient().Close(ctx, req)
}

func (g GatewayS) Kick(ctx context.Context, req *gw.KickReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdCid(req.Cid).GetGWIClient().Kick(ctx, req)
}

func (g GatewayS) Rid(ctx context.Context, req *gw.RidReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdCid(req.Cid).GetGWIClient().Rid(ctx, req)
}

func (g GatewayS) Rids(ctx context.Context, req *gw.RidsReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdCid(req.Cid).GetGWIClient().Rids(ctx, req)
}

func (g GatewayS) Push(ctx context.Context, req *gw.PushReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdCid(req.Cid).GetGWIClient().Push(ctx, req)
}

func (g GatewayS) GConn(ctx context.Context, req *gw.GConnReq) (*gw.Id32Rep, error) {
	// req.Kick = false
	req.NewVer = true
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().Conn(ctx, req)
}

func (g GatewayS) GDisc(ctx context.Context, req *gw.GDiscReq) (*gw.Id32Rep, error) {
	// req.Kick = false
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().Disc(ctx, req)
}

func (g GatewayS) GLast(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().GLast(ctx, req)
}

func (g GatewayS) GPush(ctx context.Context, req *gw.GPushReq) (*gw.Id64Rep, error) {
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().GPush(ctx, req)
}

func (g GatewayS) GLasts(ctx context.Context, req *gw.GLastsReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().GLasts(ctx, req)
}

func (g GatewayS) Send(ctx context.Context, req *gw.SendReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdGid(req.FromId).GetGWIClient().Send(ctx, req)
}

func (g GatewayS) TPush(ctx context.Context, req *gw.TPushReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdGid(req.FromId).GetGWIClient().TPush(ctx, req)
}

func (g GatewayS) TDirty(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	return gateway.Server.GetProdGid(req.Gid).GetGWIClient().TDirty(ctx, req)
}

func (g GatewayS) Revoke(ctx context.Context, req *gw.RevokeReq) (*gw.BoolRep, error) {
	if gateway.MsgMng.Db == nil {
		return Result_Fasle, nil
	}

	err := gateway.MsgMng.Db.Revoke(req.Id, req.Gid)
	if err != nil {
		return Result_Fasle, err
	}

	return Result_True, nil
}
