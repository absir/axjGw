package gateway

import "axjGW/gen/gw"

type MsgClient struct {
	cid           int64
	gatewayI      gw.GatewayIClient
	connVer       int32
	idleTime      int64
	lasting       bool
	lastTime      int64
	subLastTime   int64
	subLastId     int64
	subContinuous int32
	cidReq        *gw.CidReq
}

func (that *MsgClient) GetConnVer() int32 {
	return that.connVer
}

func (that *MsgClient) getCidReq() *gw.CidReq {
	if that.cidReq == nil {
		that.cidReq = &gw.CidReq{
			Cid: that.cid,
		}
	}

	return that.cidReq
}
