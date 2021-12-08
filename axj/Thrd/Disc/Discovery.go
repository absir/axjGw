package Disc

import (
	"axj/Kt/Kt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"errors"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type Prod struct {
	Id      string                      // 服务编号
	Addr    string                      // 服务地址
	AddrOut string                      // 服务地址外网
	Weight  float32                     // 服务权重
	Metas   map[interface{}]interface{} // 服务额外数据
}

type Discovery interface {
	Unique(paras string) string                                        // 服务唯一标识
	Cfg(unique string, paras string) interface{}                       // 服务配置
	IdleTime(cfg interface{}) int                                      // 服务检测间隔
	ListProds(cfg interface{}, name string) ([]Prod, bool)             // 获取服务列表
	WatcherProds(cfg interface{}, name string, fun func([]Prod)) error // 监听服务地址
}

type discoveryC struct {
	disc Discovery
	cfg  interface{}
}

func (that *discoveryC) ListProds(name string) ([]Prod, bool) {
	return that.disc.ListProds(that.cfg, name)
}

func (that *discoveryC) WatcherProds(name string, fun func([]Prod)) error {
	return that.disc.WatcherProds(that.cfg, name, fun)
}

type discoveryS struct {
	discC    *discoveryC
	name     string
	setFun   func(prods []Prod)
	setAsync *Util.NotifierAsync
	idleTime int64
	passTime int64
}

func (that *discoveryS) ListProds() {
	prods, ok := that.discC.ListProds(that.name)
	if prods == nil {
		if !ok {
			that.passTime = 0
		}

	} else {
		that.setFun(prods)
	}
}

type DiscoveryMng struct {
	facts     map[string]func(paras string) Discovery
	discs     map[string]Discovery
	discCs    map[string]*discoveryC
	discSs    *sync.Map
	checkTime int64
}

func (that *DiscoveryMng) Init() *DiscoveryMng {
	that.facts = map[string]func(paras string) Discovery{}
	that.discs = map[string]Discovery{}
	that.discCs = map[string]*discoveryC{}
	return that
}

func (that *DiscoveryMng) Reg(disc string, fun func(paras string) Discovery) {
	that.facts[disc] = fun
}

func (that *DiscoveryMng) GetDiscovery(disc string) *discoveryC {
	discC := that.discCs[disc]
	if discC != nil {
		return discC
	}

	idx := strings.IndexByte(disc, ':')
	if idx > 0 {
		return that.GetDiscoveryD(disc[0:idx], disc[idx+1:])
	}

	return that.GetDiscoveryD(disc, "")
}

func (that *DiscoveryMng) GetDiscoveryD(prt string, paras string) *discoveryC {
	key := prt + ":" + paras
	discC := that.discCs[key]
	if discC != nil {
		return discC
	}

	fun := that.facts[prt]
	if fun == nil {
		return nil
	}

	pDisc := that.discs[prt]
	unique := ""
	puKey := ""
	if pDisc != nil {
		unique = pDisc.Unique(paras)
		puKey = prt + ":" + unique
		disc := that.discs[puKey]
		if disc != nil {
			discC = &discoveryC{
				disc: disc,
				cfg:  disc.Cfg(unique, paras),
			}

			that.discCs[key] = discC
			return discC
		}
	}

	disc := fun(paras)
	if disc == nil {
		return nil
	}

	if pDisc == nil {
		unique = disc.Unique(paras)
		puKey = prt + ":" + unique
	}

	discC = &discoveryC{
		disc: disc,
		cfg:  disc.Cfg(unique, paras),
	}

	// 缓存协议解析
	that.discs[prt] = disc
	that.discs[puKey] = disc
	that.discCs[key] = discC
	return discC
}

func (that *DiscoveryMng) SetDiscoveryS(disc string, name string, setFun func(prods []Prod), idleTime int, panic bool) *discoveryS {
	// map初始化
	if that.discSs == nil {
		that.discSs = new(sync.Map)
	}

	// 已载入检测
	val, has := that.discSs.Load(name)
	discS, _ := val.(*discoveryS)
	if has {
		if panic {
			Kt.Panic(errors.New("DiscSs Has " + name))

		} else {
			AZap.Logger.Error("DiscSs Has " + name)
		}

		return discS
	}

	discC := that.GetDiscovery(disc)
	if discC == nil {
		return nil
	}

	discS = &discoveryS{
		discC:  discC,
		name:   name,
		setFun: setFun,
	}

	// setAsync idleTime
	discS.setAsync = Util.NewNotifierAsync(discS.ListProds, nil, nil)
	if idleTime <= 0 {
		idleTime = discC.disc.IdleTime(discC.cfg)
	}

	discS.idleTime = int64(idleTime) * int64(time.Second)

	err := discS.discC.WatcherProds(name, setFun)
	if err != nil {
		if panic {
			Kt.Panic(err)

		} else {
			AZap.Logger.Error("WatcherProds Err "+name, zap.Error(err))
		}
	}

	that.discSs.Store(name, discS)
	return discS
}

func (that *DiscoveryMng) CheckEmpty() bool {
	return that.discSs == nil
}

func (that *DiscoveryMng) CheckStop() {
	that.checkTime = 0
}

func (that *DiscoveryMng) CheckLoop(checkDrt time.Duration) {
	checkTime := time.Now().UnixNano()
	for checkTime == that.checkTime {
		time.Sleep(checkDrt)
		if that.discSs != nil {
			time := time.Now().UnixNano()
			that.discSs.Range(func(key, value interface{}) bool {
				discS, _ := key.(*discoveryS)
				// 定时刷新prods
				if discS.passTime <= time {
					discS.passTime = time + discS.idleTime
					discS.setAsync.Start(nil)
				}

				return true
			})
		}
	}
}

// 注册默认发现类
func (that *DiscoveryMng) RegDefs() {
}
