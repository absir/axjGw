package gateway

import (
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"strconv"
)

type zDevAcl struct {
}

var ZDevAcl = &zDevAcl{}

func (Z zDevAcl) Login(ctx context.Context, in *gw.LoginReq, opts ...grpc.CallOption) (*gw.LoginRep, error) {
	return &gw.LoginRep{
		Back: true,
	}, nil
}

func (Z zDevAcl) LoginBack(ctx context.Context, in *gw.LoginBack, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	// 加载ZG组
	Util.GoSubmit(func() {
		Server.gatewayISC.GLasts(ctx, &gw.GLastsReq{
			Cid:        in.Cid,
			Unique:     strconv.FormatInt(in.Cid, 10),
			Gid:        "ZG",
			Continuous: 1,
		}, opts...)
	})
	return &gw.Id32Rep{
		Id: int32(gw.Result_Succ),
	}, nil
}

func (Z zDevAcl) DiscBack(ctx context.Context, in *gw.LoginBack, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return nil, nil
}

func (Z zDevAcl) Team(ctx context.Context, in *gw.GidReq, opts ...grpc.CallOption) (*gw.TeamRep, error) {
	return nil, nil
}

func (Z zDevAcl) Req(ctx context.Context, in *gw.PassReq, opts ...grpc.CallOption) (*gw.DataRep, error) {
	if in.Uri == "test/sendU" {
		// 向ZG组发送消息
		var strs []string
		json.Unmarshal(in.Data, &strs)
		Util.GoSubmit(func() {
			Server.gatewayISC.GPush(ctx, &gw.GPushReq{
				Gid:  "ZG",
				Qs:   3,
				Uri:  strs[0],
				Data: KtUnsafe.StringToBytes(strs[1]),
			}, opts...)
		})
		return &gw.DataRep{}, nil
	}

	return nil, ERR_FAIL
}

func (Z zDevAcl) Send(ctx context.Context, in *gw.PassReq, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	Z.Req(ctx, in, opts...)
	return &gw.Id32Rep{
		Id: int32(gw.Result_Succ),
	}, nil
}

func (Z zDevAcl) Addr(ctx context.Context, in *gw.AddrReq, opts ...grpc.CallOption) (*gw.AddrRep, error) {
	return nil, nil
}
