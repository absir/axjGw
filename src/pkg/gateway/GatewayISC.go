package gateway

import (
	"axjGW/gen/gw"
	"context"
	"google.golang.org/grpc"
)

type gatewayISC struct {
	Server gw.GatewayIServer
}

func (that gatewayISC) Close(ctx context.Context, in *gw.CloseReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Close(ctx, in)
}

func (that gatewayISC) Kick(ctx context.Context, in *gw.KickReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Kick(ctx, in)
}

func (that gatewayISC) Conn(ctx context.Context, in *gw.GConnReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Conn(ctx, in)
}

func (that gatewayISC) Disc(ctx context.Context, in *gw.GDiscReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Disc(ctx, in)
}

func (that gatewayISC) Alive(ctx context.Context, in *gw.CidReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Alive(ctx, in)
}

func (that gatewayISC) Rid(ctx context.Context, in *gw.RidReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Rid(ctx, in)
}

func (that gatewayISC) Rids(ctx context.Context, in *gw.RidsReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Rids(ctx, in)
}

func (that gatewayISC) Last(ctx context.Context, in *gw.ILastReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Last(ctx, in)
}

func (that gatewayISC) Push(ctx context.Context, in *gw.PushReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Push(ctx, in)
}

func (that gatewayISC) GQueue(ctx context.Context, in *gw.IGQueueReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.GQueue(ctx, in)
}

func (that gatewayISC) GClear(ctx context.Context, in *gw.IGClearReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.GClear(ctx, in)
}

func (that gatewayISC) GLasts(ctx context.Context, in *gw.GLastsReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.GLasts(ctx, in)
}

func (that gatewayISC) GLast(ctx context.Context, in *gw.GidReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.GLast(ctx, in)
}

func (that gatewayISC) GPush(ctx context.Context, in *gw.GPushReq, opts ...grpc.CallOption) (*gw.Id64Rep, error) {
	return that.Server.GPush(ctx, in)
}

func (that gatewayISC) GPushA(ctx context.Context, in *gw.IGPushAReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.GPushA(ctx, in)
}

func (that gatewayISC) Send(ctx context.Context, in *gw.SendReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.Send(ctx, in)
}

func (that gatewayISC) TPush(ctx context.Context, in *gw.TPushReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.TPush(ctx, in)
}

func (that gatewayISC) TDirty(ctx context.Context, in *gw.GidReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.TDirty(ctx, in)
}

func (that gatewayISC) TStarts(ctx context.Context, in *gw.GidReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return that.Server.TStarts(ctx, in)
}
