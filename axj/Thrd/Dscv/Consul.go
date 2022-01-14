package Dscv

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Thrd/Util"
	"github.com/hashicorp/consul/api"
	"strconv"
	"strings"
	"time"
)

type consul struct {
}

type consulCfg struct {
	Config     *api.Config
	IdleTime   int
	WaitTime   time.Duration
	WatcherDrt time.Duration
	CheckHttp  bool
	Check      *api.AgentServiceCheck
	client     *api.Client
	RegChkDrt  int64
	regChkTime int64
}

type consulCtx struct {
	lastIndex uint64
	idleTime  int64
	passTime  int64
}

func (c consul) Unique(paras string) string {
	return paras
}

func (c consul) Cfg(unique string, paras string) interface{} {
	// 配置初始化
	cCfg := &consulCfg{
		Config: &api.Config{
			Address: "127.0.0.1:8500",
		},
		IdleTime:   600,
		WaitTime:   30,
		WatcherDrt: 1,
		CheckHttp:  true,
		Check: &api.AgentServiceCheck{
			Status:                         "passing",
			DeregisterCriticalServiceAfter: "5s",
			TTL:                            "45s",
		},
		RegChkDrt: 30,
	}

	APro.SubCfgBind("consul."+paras, cCfg)

	cCfg.WaitTime *= time.Second
	cCfg.WatcherDrt *= time.Second
	cCfg.RegChkDrt *= int64(time.Second)

	// 初始化client
	client, err := api.NewClient(cCfg.Config)
	Kt.Panic(err)
	cCfg.client = client
	return cCfg
}

func (c consul) IdleTime(cfg interface{}) int {
	cCfg := cfg.(*consulCfg)
	return cCfg.IdleTime
}

func (c consul) CtxProds(cfg interface{}) interface{} {
	return &consulCtx{}
}

func (c consul) ListProds(cfg interface{}, ctx interface{}, name string) ([]*Prod, error) {
	cCfg := cfg.(*consulCfg)
	cCtx := ctx.(*consulCtx)
	return c.ReqProds(cCfg, cCtx, name, nil)
}

func (c consul) ReqProds(cCfg *consulCfg, cCtx *consulCtx, name string, wait *api.QueryOptions) ([]*Prod, error) {
	if wait != nil {
		q := wait
		now := time.Now().UnixNano()
		if cCfg.WaitTime > 0 && cCtx.passTime > now {
			q.WaitIndex = cCtx.lastIndex
			q.WaitTime = cCfg.WaitTime

		} else {
			q.WaitIndex = 0
			q.WaitTime = 0
		}

		_, meta, err := cCfg.client.Catalog().Service(name, "", q)
		if err != nil {
			return nil, err
		}

		if meta != nil {
			if cCtx.lastIndex == meta.LastIndex && q.WaitTime > 0 {
				return nil, nil
			}

			cCtx.lastIndex = meta.LastIndex
			if cCtx.idleTime > 0 {
				cCtx.passTime = now + cCtx.idleTime
			}
		}
	}

	_, infos, err := cCfg.client.Agent().AgentHealthServiceByName(name)
	if err != nil {
		return nil, err
	}

	if infos == nil {
		return nil, nil
	}

	prods := make([]*Prod, len(infos))
	prods = prods[:0]
	for _, info := range infos {
		service := info.Service
		id := service.ID
		i := strings.LastIndexByte(id, '-')
		if i >= 0 {
			id = id[i+1:]
		}

		prod := &Prod{
			Id:    id,
			Addr:  service.Address,
			Port:  service.Port,
			Metas: service.Meta,
		}

		prods = append(prods, prod)
	}

	return prods, nil
}

func (c consul) WatcherProds(cfg interface{}, ctx interface{}, name string, idleTime *int64, fun func([]*Prod, error)) (bool, error) {
	cCfg := cfg.(*consulCfg)
	if cCfg.WatcherDrt > 0 && fun != nil {
		cCtx := ctx.(*consulCtx)
		if idleTime != nil {
			cCtx.idleTime = *idleTime
			cCtx.passTime = time.Now().UnixNano() + cCtx.idleTime
		}

		Util.GoSubmit(func() {
			wait := &api.QueryOptions{}
			for Kt.Active {
				time.Sleep(cCfg.WatcherDrt)
				fun(c.ReqProds(cCfg, cCtx, name, wait))
			}
		})

		if idleTime != nil {
			*idleTime = 0
		}
	}

	return true, nil
}

func (c consul) RegCheck(cfg interface{}, now int64) bool {
	cCfg := cfg.(*consulCfg)
	if cCfg.RegChkDrt > 0 && cCfg.regChkTime <= now {
		cCfg.regChkTime = now + cCfg.RegChkDrt
		return true
	}

	return false
}

func (c consul) RegMiss(cfg interface{}) bool {
	return false
}

type consulCtxReg struct {
	service *api.AgentServiceRegistration
	checkId string
}

func (c consul) RegCtx(cfg interface{}, name string, port int, metas map[string]string) (interface{}, error) {
	cCfg := cfg.(*consulCfg)
	service := &api.AgentServiceRegistration{}
	service.Name = name
	service.ID = service.Name + "-" + strconv.Itoa(int(APro.WorkId()))
	service.Address = GetDscvCfg().Ip
	service.Port = port
	if metas != nil {
		service.Meta = metas
	}

	service.Check = cCfg.Check
	APro.StopAdd(func() {
		cCfg.client.Agent().ServiceDeregister(service.ID)
	})

	return &consulCtxReg{
		service: service,
		checkId: "service:" + service.ID,
	}, nil
}

func (c consul) RegProd(cfg interface{}, ctx interface{}, reged bool) error {
	cCfg := cfg.(*consulCfg)
	reg := ctx.(*consulCtxReg)
	var err error = nil
	if reged {
		err = cCfg.client.Agent().PassTTL(reg.checkId, "")
		if err == nil {
			return nil
		}
	}

	err = cCfg.client.Agent().ServiceRegister(reg.service)
	if err != nil {
		return err
	}

	return cCfg.client.Agent().PassTTL(reg.checkId, "")
}
