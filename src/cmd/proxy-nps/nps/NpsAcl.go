package nps

import (
	"axj/Thrd/AZap"
	"axjGW/gen/gw"
	"axjGW/pkg/gws"
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strconv"
	"strings"
)

type NpsAcl struct {
}

func (that *NpsAcl) Login(ctx context.Context, in *gw.LoginReq, opts ...grpc.CallOption) (*gw.LoginRep, error) {
	var strs []string
	json.Unmarshal(in.Data, &strs)
	secret := strs[0]
	client := ClientSecret(secret)
	if client != nil {
		// 登录成功
		client.Cid = in.Cid
		return &gw.LoginRep{
			Uid:  int64(client.Id),
			Succ: true,
		}, nil
	}

	return nil, nil
}

func (that *NpsAcl) LoginBack(ctx context.Context, in *gw.LoginBack, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return gws.Result_Succ_Rep, nil
}

func (that *NpsAcl) DiscBack(ctx context.Context, in *gw.LoginBack, opts ...grpc.CallOption) (*gw.Id32Rep, error) {
	return gws.Result_Succ_Rep, nil
}

func (that *NpsAcl) Team(ctx context.Context, in *gw.GidReq, opts ...grpc.CallOption) (*gw.TeamRep, error) {
	return nil, nil
}

func (that *NpsAcl) Addr(ctx context.Context, in *gw.AddrReq, opts ...grpc.CallOption) (*gw.AddrRep, error) {
	name := in.Name

	if name != "" {
		// host代理
		i := strings.LastIndexByte(name, ':')
		if i > 0 {
			name = name[:i]
		}

		addrRep := HostAddrRep(name)
		return addrRep, nil
	}

	// tcp代理
	id, err := strconv.Atoi(in.SName)
	if err != nil {
		AZap.Error("unknown sName "+in.SName, zap.Error(err))
		return nil, nil
	}

	val, _ := TcpMap.Load(id)
	npcTcp, _ := val.(*NpsTcp)
	if npcTcp != nil {
		return npcTcp.AddrRep(), nil
	}

	return nil, nil
}

func (that *NpsAcl) GwReg(ctx context.Context, in *gw.GwRegReq, opts ...grpc.CallOption) (*gw.BoolRep, error) {
	return nil, nil
}

func (that *NpsAcl) Traffic(ctx context.Context, in *gw.TrafficReq, opts ...grpc.CallOption) (*gw.BoolRep, error) {
	return nil, nil
}
