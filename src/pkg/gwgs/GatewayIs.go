package gwgs

import (
	"axjGW/gen/gw"
	"context"
)

type GatewayIs struct {
}

func (g GatewayIs) Close(ctx context.Context, req *gw.CloseReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Kick(ctx context.Context, req *gw.KickReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Conn(ctx context.Context, req *gw.IConnReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Disc(ctx context.Context, req *gw.IDiscReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Alive(ctx context.Context, req *gw.CidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Rid(ctx context.Context, req *gw.RidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Rids(ctx context.Context, req *gw.RidsReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Last(ctx context.Context, req *gw.ILastReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Push(ctx context.Context, req *gw.IPushReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) GQueue(ctx context.Context, req *gw.IGQueueReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) GClear(ctx context.Context, req *gw.IGClearReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) GLasts(ctx context.Context, req *gw.GLastsReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) GLast(ctx context.Context, req *gw.IGLastReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) GPush(ctx context.Context, req *gw.IGPushReq) (*gw.Id64Rep, error) {
	panic("implement me")
}

func (g GatewayIs) GPushA(ctx context.Context, req *gw.IGPushAReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) Send(ctx context.Context, req *gw.ISendReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) TPush(ctx context.Context, req *gw.ITPushReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) TDirty(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayIs) TStarts(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}
