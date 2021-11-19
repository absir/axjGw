package proxy

import (
	"axjGW/gen/gw"
	"context"
	"google.golang.org/grpc"
)

var LoginRepOk = &gw.LoginRep{}

type PrxAcl struct {
}

func (p PrxAcl) Login(ctx context.Context, in *gw.LoginReq, opts ...grpc.CallOption) (*gw.LoginRep, error) {
	return LoginRepOk, nil
}

func (p PrxAcl) LoginBack(ctx context.Context, in *gw.LoginBack, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return nil, nil
}

func (p PrxAcl) DiscBack(ctx context.Context, in *gw.LoginBack, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return nil, nil
}

func (p PrxAcl) Team(ctx context.Context, in *gw.GidReq, opts ...grpc.CallOption) (*gw.TeamRep, error) {
	return nil, nil
}
