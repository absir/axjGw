package gateway

import (
	"context"
	"fmt"
	"gw"
)

type Remote struct {
}

func (r Remote) Req(ctx context.Context, uid int64, sid string, uri string, bytes []byte) (_r []byte, _err error) {
	//helper, _ := thrift.GetResponseHelper(ctx)
	//helper.THeaderResponseHelper
	fmt.Println(sid)
	fmt.Println(uri)
	return nil, nil
}

func (r Remote) Send(ctx context.Context, uid int64, sid string, uri string, bytes []byte) (_err error) {
	panic("implement me")
}

func (r Remote) Reg(ctx context.Context, serv *gw.Serv) (_err error) {
	panic("implement me")
}

func (r Remote) Beat(ctx context.Context) (_err error) {
	panic("implement me")
}

func (r Remote) Push(ctx context.Context, uid int64, sid string, msg *gw.Msg) (_err error) {
	panic("implement me")
}

func (r Remote) Group(ctx context.Context, sid string, group *gw.Group, deleted bool) (_err error) {
	panic("implement me")
}
