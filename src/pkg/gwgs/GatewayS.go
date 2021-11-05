package gwgs

import (
	"axjGW/gen/gw"
	"context"
)

type GatewayS struct {
}

func (g GatewayS) Close(ctx context.Context, req *gw.CloseReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) Kick(ctx context.Context, req *gw.KickReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) Rid(ctx context.Context, req *gw.RidReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) Rids(ctx context.Context, req *gw.RidsReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) Push(ctx context.Context, req *gw.PushReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) GLast(ctx context.Context, req *gw.GidReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) GPush(ctx context.Context, req *gw.GPushReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) GConn(ctx context.Context, req *gw.GConnReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) GDisc(ctx context.Context, req *gw.GDiscReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) GLasts(ctx context.Context, req *gw.GLastsReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) Send(ctx context.Context, req *gw.SendReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) TPush(ctx context.Context, req *gw.TPushReq) (*gw.SuccRep, error) {
	panic("implement me")
}

func (g GatewayS) TDirty(ctx context.Context, req *gw.GidReq) (*gw.SuccRep, error) {
	panic("implement me")
}

