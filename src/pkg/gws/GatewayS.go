package gws

import (
	"axjGW/gen/gw"
	"axjGW/pkg/gateway"
	"context"
)

type GatewayS struct {
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
