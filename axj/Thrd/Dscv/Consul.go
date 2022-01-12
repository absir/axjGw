package Dscv

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"github.com/hashicorp/consul/api"
	"go.uber.org/zap"
	"net/http"
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
	Check      *api.AgentServiceCheck
	CheckPort  int
	client     *api.Client
	regCheck   string
}

const REG_CHECK_KEY = "AXJ_CHECK_KEY"

type consulCtx struct {
	hash     string
	idleTime int64
	passTime int64
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
		WatcherDrt: 3,
		Check: &api.AgentServiceCheck{
			//HTTP:                           "http://127.0.0.1:8682/",
			Status:                         "passing",
			Interval:                       "30s",
			DeregisterCriticalServiceAfter: "600s",
		},
		CheckPort: 8682,
	}

	APro.SubCfgBind("consule-"+paras, cCfg)
	if cCfg.CheckPort > 0 {
		http.Handle("/consul/check", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			id := request.URL.Query().Get("id")
			if KtCvt.ToInt32(id) == APro.WorkId() {
				writer.Write(KtUnsafe.StringToBytes("ok"))
				return
			}

			// WorkId校验失败
			writer.WriteHeader(400)
			writer.Write(KtUnsafe.StringToBytes("fail"))
		}))
		cCfg.Check.HTTP = "http://" + GetDscvCfg().Ip + ":8682/consul/check?id=" + strconv.Itoa(int(APro.WorkId()))
	}

	cCfg.WaitTime *= time.Second
	cCfg.WatcherDrt *= time.Second

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
	return c.ReqProds(cCfg, cCtx, name, false)
}

func (c consul) ReqProds(cCfg *consulCfg, cCtx *consulCtx, name string, wait bool) ([]*Prod, error) {
	var hash = ""
	var infos []api.AgentServiceChecksInfo = nil
	var err error = nil
	if wait && cCfg.WaitTime > 0 {
		q := &api.QueryOptions{
			WaitHash: cCtx.hash,
			WaitTime: cCfg.WaitTime,
		}

		hash, infos, err = cCfg.client.Agent().AgentHealthServiceByNameOpts(name, q)

	} else {
		hash, infos, err = cCfg.client.Agent().AgentHealthServiceByName(name)
	}

	if err != nil {
		return nil, err
	}

	// hash 未改变
	if hash == cCtx.hash {
		return nil, nil
	}

	if hash != "" {
		cCtx.hash = hash
		if cCtx.idleTime > 0 {
			cCtx.passTime = time.Now().UnixNano() + cCtx.idleTime
		}
	}

	if infos == nil {
		return nil, nil
	}

	prods := make([]*Prod, len(infos))
	for idx, info := range infos {
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

		prods[idx] = prod
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
			for {
				time.Sleep(cCfg.WatcherDrt)
				if cCtx.idleTime > 0 && cCtx.passTime <= time.Now().UnixNano() {
					fun(c.ReqProds(cCfg, cCtx, name, false))

				} else {
					fun(c.ReqProds(cCfg, cCtx, name, true))
				}
			}
		})

		if idleTime != nil {
			*idleTime = 0
		}
	}

	return true, nil
}

func (c consul) RegMiss(cfg interface{}) bool {
	cCfg := cfg.(*consulCfg)
	for {
		pair, _, err := cCfg.client.KV().Get(REG_CHECK_KEY, &api.QueryOptions{})
		if err != nil {
			AZap.Error("REG_CHECK_KEY Err", zap.Error(err))
			return false
		}

		if pair == nil {

			time.Sleep(time.Second)
			continue

		} else {
			regCheck := KtUnsafe.BytesToString(pair.Value)
			if cCfg.regCheck != regCheck {
				cCfg.regCheck = regCheck
				return true
			}
		}
	}

	return false
}

func (c consul) RegCtx(cfg interface{}, name string, port int, metas map[string]string) (interface{}, error) {
	cCfg := cfg.(*consulCfg)
	service := &api.AgentServiceRegistration{}
	service.Name = GetDscvCfg().Group + name
	service.ID = service.Name + "-" + strconv.Itoa(int(APro.WorkId()))
	service.Address = GetDscvCfg().Ip
	service.Port = port
	if metas != nil {
		service.Meta = metas
	}

	service.Check = cCfg.Check
	return service, nil
}

func (c consul) RegProd(cfg interface{}, ctx interface{}) error {
	cCfg := cfg.(*consulCfg)
	service := ctx.(*api.AgentServiceRegistration)
	return cCfg.client.Agent().ServiceRegister(service)
}
