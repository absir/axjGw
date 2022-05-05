package proxy

import (
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"axjGW/pkg/gws"
	"context"
	"time"
)

type GatewayS struct {
}

func (g GatewayS) Uid(ctx context.Context, req *gw.CidReq) (*gw.UIdRep, error) {
	client := PrxServMng.Manager.Client(req.Cid)
	if client == nil {
		return nil, nil
	}

	return &gw.UIdRep{Sid: Handler.ClientG(client).gid}, nil
}

func (g GatewayS) Online(ctx context.Context, req *gw.GidReq) (*gw.BoolRep, error) {
	_, ok := PrxMng.gidMap.Load(req.Gid)
	if ok {
		return gws.Result_True, nil
	}

	return gws.Result_Fasle, nil
}

func (g GatewayS) Onlines(ctx context.Context, req *gw.GidsReq) (*gw.BoolsRep, error) {
	vals := make([]bool, len(req.Gids))
	for i, gid := range req.Gids {
		_, vals[i] = PrxMng.gidMap.Load(gid)
	}

	return &gw.BoolsRep{Vals: vals}, nil
}

func (g GatewayS) Close(ctx context.Context, req *gw.CloseReq) (*gw.Id32Rep, error) {
	client := PrxServMng.Manager.Client(req.Cid)
	if client == nil {
		return gws.Result_Fail_Rep, nil
	}

	client.Get().Close(nil, req.Reason)
	return gws.Result_Succ_Rep, nil
}

func (g GatewayS) Kick(ctx context.Context, req *gw.KickReq) (*gw.Id32Rep, error) {
	client := PrxServMng.Manager.Client(req.Cid)
	if client == nil {
		return gws.Result_Fail_Rep, nil
	}

	client.Get().Kick(req.Data, false, 0)
	return gws.Result_Succ_Rep, nil
}

func (g GatewayS) Rid(ctx context.Context, req *gw.RidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) Rids(ctx context.Context, req *gw.RidsReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) Push(ctx context.Context, req *gw.PushReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) GConn(ctx context.Context, req *gw.GConnReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) GDisc(ctx context.Context, req *gw.GDiscReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) GLast(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) GPush(ctx context.Context, req *gw.GPushReq) (*gw.Id64Rep, error) {
	panic("implement me")
}

func (g GatewayS) GLasts(ctx context.Context, req *gw.GLastsReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) Send(ctx context.Context, req *gw.SendReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) TPush(ctx context.Context, req *gw.TPushReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) TDirty(ctx context.Context, req *gw.GidReq) (*gw.Id32Rep, error) {
	panic("implement me")
}

func (g GatewayS) Revoke(ctx context.Context, req *gw.RevokeReq) (*gw.BoolRep, error) {
	panic("implement me")
}

var sProxy *gw.ProxyReq

func (g GatewayS) SetProxy(ctx context.Context, req *gw.ProxyReq) (*gw.BoolRep, error) {
	sProxy = req
	return gws.Result_True, nil
}

func (g GatewayS) SetProds(ctx context.Context, rep *gw.ProdsRep) (*gw.BoolRep, error) {
	panic("implement me")
}

func (g GatewayS) DialProxy(ctx context.Context, req *gw.DialProxyReq) (*gw.BoolRep, error) {
	ok := PrxMng.Dial(req.Cid, req.Gid, req.Addr, time.Duration(req.Timeout))
	if ok {
		return gws.Result_True, nil
	}

	return gws.Result_Fasle, nil
}

func (g GatewayS) DialsProxy(ctx context.Context, req *gw.DialsProxyReq) (*gw.BoolsRep, error) {
	size := len(req.Dials)
	bools := make([]bool, size)
	reps := make(chan bool, 1)
	for i, dial := range req.Dials {
		if dial.Timeout <= 0 {
			dial.Timeout = int64(req.Timeout)
		}

		_i := i
		_dial := dial
		Util.GoSubmit(func() {
			bools[_i] = PrxMng.Dial(_dial.Cid, _dial.Gid, _dial.Addr, time.Duration(_dial.Timeout))
			end := false
			PrxServMng.locker.Lock()
			size -= 1
			end = size <= 0
			PrxServMng.locker.Unlock()
			if end && reps != nil {
				defer recover()
				reps <- true
			}
		})
	}

	timeout := time.Duration(req.Timeout) * time.Millisecond
	if timeout <= 0 {
		timeout = Config.DialTimeout
	}

	select {
	case <-reps:
		break
	case <-time.After(timeout):
		break
	}

	close(reps)
	reps = nil
	return &gw.BoolsRep{
		Vals: bools,
	}, nil
}

func (g GatewayS) UnreadTids(ctx context.Context, tids *gw.UnreadTids) (*gw.Id32Rep, error) {
	return nil, nil
}
