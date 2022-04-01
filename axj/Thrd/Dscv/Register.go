package Dscv

import (
	"axj/APro"
	"axj/Kt/Kt"
	"time"
)

type Register interface {
	RegCheck(cfg interface{}, now int64) bool                                                    // 注册检查
	RegMiss(cfg interface{}) bool                                                                // 注册中心过期，重启等
	RegCtx(cfg interface{}, name string, port int, metas map[string]string) (interface{}, error) // 注册Context
	RegProd(cfg interface{}, ctx interface{}, reged bool) error                                  // 注册服务
}

type DscvCfg struct {
	Group     string
	Ip        string
	HttpPort  int
	CheckDrt  time.Duration
	RegWait   time.Duration
	RegChkDrt int64
}

var dscvCfg *DscvCfg

func GetDscvCfg() *DscvCfg {
	if dscvCfg == nil {
		cfg := &DscvCfg{
			Group:     "axj",
			HttpPort:  8682,
			CheckDrt:  10,
			RegChkDrt: 60,
		}

		if APro.Cfg != nil {
			APro.SubCfgBind("dscv", cfg)
		}

		if cfg.Ip == "" {
			cfg.Ip = APro.GetLocalIp()
		}

		if cfg.RegWait <= 0 {
			cfg.RegWait = 30
		}

		dscvCfg = cfg
	}

	return dscvCfg
}

type prodReg struct {
	dscvC *discoveryC
	reg   Register
	ctx   interface{}
	reged bool
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

	ctx, err := reg.RegCtx(dscvC.cfg, name, port, metas)
	if err != nil {
		return err
	}

	dscvC.addRegs(&prodReg{
		dscvC: dscvC,
		reg:   reg,
		ctx:   ctx,
	})

	return nil
}
