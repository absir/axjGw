package Dscv

import (
	"axj/APro"
	"axj/Kt/Kt"
	"time"
)

type Register interface {
	RegMiss(cfg interface{}) bool                                                                // 注册中心过期，重启等
	RegCtx(cfg interface{}, name string, port int, metas map[string]string) (interface{}, error) // 注册Context
	RegProd(cfg interface{}, ctx interface{}) error                                              // 注册服务
}

type DscvCfg struct {
	Group     string
	Ip        string
	CheckDrt  time.Duration
	MissWait  time.Duration
	RegChkDrt int64
}

var dscvCfg *DscvCfg

func GetDscvCfg() *DscvCfg {
	if dscvCfg == nil {
		cfg := &DscvCfg{
			CheckDrt:  30,
			RegChkDrt: 600,
		}
		if APro.Cfg != nil {
			APro.SubCfgBind("dscv", cfg)
		}

		if cfg.Ip == "" {
			cfg.Ip = APro.GetLocalIp()
		}

		if cfg.MissWait <= 0 {
			cfg.MissWait = 10
		}

		cfg.CheckDrt *= time.Second
		cfg.MissWait *= time.Second
		cfg.RegChkDrt *= int64(time.Second)
		dscvCfg = cfg
	}

	return dscvCfg
}

type prodReg struct {
	dscvC *discoveryC
	reg   Register
	ctx   interface{}
}

func (that *DiscoveryMng) RegProd(dscv string, name string, port int, metas map[string]string) error {
	var reg Register = nil
	dscvC := that.GetDiscovery(dscv)
	if dscvC != nil {
		reg, _ = dscvC.dscv.(Register)
	}

	if reg == nil {
		return Kt.NewErrReason("Dscv Reg Not Found " + dscv)
	}

	_, err := reg.RegCtx(dscvC.cfg, name, port, metas)
	if err != nil {
		return err
	}

	//regProd := &regProd{
	//	dscvC: dscvC,
	//	reg:   reg,
	//	ctx:   ctx,
	//}

	return nil
}
