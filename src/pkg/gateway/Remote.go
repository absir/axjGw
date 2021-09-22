package gateway

import (
	"context"
	"gw"
)

type Remote struct {
}

func (r Remote) Reg(ctx context.Context, serv *gw.Serv) (_err error) {
	panic("implement me")
}

func (r Remote) Dirty(ctx context.Context, sid string) (_err error) {
	panic("implement me")
}

func (r Remote) Push(ctx context.Context, cid int64, uid int64, sid string, msg *gw.Msg) (_r bool, _err error) {
	panic("implement me")
}



