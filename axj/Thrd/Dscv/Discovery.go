package Dscv

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCfg"
	"axj/Kt/KtCvt"
	"axj/Kt/KtJson"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"errors"
	"go.uber.org/zap"
	"strings"
	"sync"
	"time"
)

type Prod struct {
	Id    string            // 服务编号
	Addr  string            // 服务地址
	Port  int               // 服务端口
	Metas map[string]string // 服务额外数据
}

func (that *Prod) GetPort(portKey string, portDef int) int {
	port := 0
	if portKey == "" || portKey == "@" {
		port = that.Port

	} else if that.Metas != nil {
		port = int(KtCvt.ToInt32(that.Metas[portKey]))
	}

	if port <= 0 {
		port = portDef
	}

	return port
}

type Discovery interface {
	Unique(paras string) string                                                                                          // 服务唯一标识
	Cfg(unique string, paras string) interface{}                                                                         // 服务配置
	IdleTime(cfg interface{}) int                                                                                        // 服务检测间隔
	CtxProds(cfg interface{}) interface{}                                                                                // 服务发现Context
	ListProds(cfg interface{}, ctx interface{}, name string) ([]*Prod, error)                                            // 获取服务列表
	WatcherProds(cfg interface{}, ctx interface{}, name string, idleTime *int64, fun func([]*Prod, error)) (bool, error) // 监听服务地址
}

type discoveryC struct {
	mng       *DiscoveryMng
	dscv      Discovery
	cfg       interface{}
	regs      []*prodReg
	regsTime  int64
	regsAsync *Util.NotifierAsync
}

func (that *discoveryC) CtxProds() interface{} {
	return that.dscv.CtxProds(that.cfg)
}

func (that *discoveryC) ListProds(ctx interface{}, name string) ([]*Prod, error) {
	return that.dscv.ListProds(that.cfg, ctx, name)
}

func (that *discoveryC) WatcherProds(ctx interface{}, name string, idleTime *int64, fun func([]*Prod, error)) (bool, error) {
	return that.dscv.WatcherProds(that.cfg, ctx, name, idleTime, fun)
}

func (that *discoveryC) addRegs(reg *prodReg) {
	if that.regs == nil {
		that.regs = make([]*prodReg, 1)[:0]
	}

	if that.regsAsync == nil {
		that.regsAsync = Util.NewNotifierAsync(that.doRegs, nil, nil)
	}

	that.regs = append(that.regs, reg)
}

func (that *discoveryC) doRegs() {
	that.regsTime = time.Now().UnixNano() + GetDscvCfg().RegChkDrt
	for i, reg := range that.regs {
		if i == 0 {
			if !reg.reg.RegMiss(that.cfg) {
				break
			}
		}

		for {
			err := reg.reg.RegProd(that.cfg, reg.ctx)
			if err == nil {
				break
			}

			AZap.Error("doRegs err", zap.Error(err))
			time.Sleep(GetDscvCfg().RegWait)
		}
	}
}

type discoveryS struct {
	dscvC     *discoveryC
	name      string
	setFun    func(prods []*Prod)
	setAsync  *Util.NotifierAsync
	ctx       interface{}
	idleTime  int64
	passTime  int64
	prodsHash string
}

func (that *discoveryS) ListProds() {
	that.passTime = time.Now().UnixNano() + that.idleTime
	that.SetProds(that.dscvC.ListProds(that.ctx, that.name))
}

func (that *discoveryS) SetProds(prods []*Prod, err error) {
	if err != nil {
		AZap.Logger.Error("SetProds Err "+that.name, zap.Error(err))
	}

	if prods == nil {
		return
	}

	hash := ""
	hash, err = KtJson.ToJsonStr(prods)
	if err != nil {
		AZap.Logger.Error("SetProds Json Err "+that.name, zap.Error(err))
	}

	if err != nil || that.prodsHash != hash {
		that.setFun(prods)
		that.prodsHash = hash
	}
}

type DiscoveryMng struct {
	locker    sync.Locker
	dscvs     map[string]Discovery
	facts     map[string]func(paras string) Discovery
	dscvCs    map[string]*discoveryC
	dscvSs    *sync.Map
	checkTime int64
}

func NewMng() *DiscoveryMng {
	that := new(DiscoveryMng)
	that.locker = new(sync.Mutex)
	that.dscvs = map[string]Discovery{}
	that.facts = map[string]func(paras string) Discovery{}
	that.dscvCs = map[string]*discoveryC{}
	that.regDefs()
	return that
}

var instMng *DiscoveryMng

func InstMng() *DiscoveryMng {
	if instMng == nil {
		instMng = NewMng()
	}

	return instMng
}

type dscvReg struct {
	Dscv  string
	Port  int
	Metas map[string]string
}

func InstMngStart(create bool) {
	dReg := APro.Cfg != nil && KtCvt.ToBool(KtCfg.Get(APro.Cfg, "dscv.reg"))
	if (instMng != nil || create || dReg) && GetDscvCfg().CheckDrt > 0 {
		// 服务注册
		if dReg {
			if mp, _ := KtCfg.Get(APro.Cfg, "dscv.regs").(*Kt.LinkedMap); mp != nil {
				reg := &dscvReg{}
				mp.Range(func(key interface{}, val interface{}) bool {
					reg.Dscv = ""
					KtCvt.BindInterface(reg, val)
					if reg.Dscv != "" {
						err := InstMng().RegProd(reg.Dscv, KtCvt.ToString(key), reg.Port, reg.Metas)
						Kt.Panic(err)
					}

					return true
				})
			}
		}

		go InstMng().CheckLoop(GetDscvCfg().CheckDrt)
	}
}

func (that *DiscoveryMng) GetDiscovery(dscv string) *discoveryC {
	dscvC := that.dscvCs[dscv]
	if dscvC != nil {
		return dscvC
	}

	idx := strings.IndexByte(dscv, ':')
	if idx > 0 {
		return that.GetDiscoveryD(dscv, strings.TrimSpace(dscv[0:idx]), strings.TrimSpace(dscv[idx+1:]))
	}

	return that.GetDiscoveryD(dscv, dscv, "")
}

func (that *DiscoveryMng) GetDiscoveryD(key string, prt string, paras string) *discoveryC {
	dscvC := that.dscvCs[key]
	if dscvC != nil {
		return dscvC
	}

	dscv := that.dscvs[prt]
	if dscv == nil {
		fact := that.facts[prt]
		if fact == nil {
			return nil
		}

		dscv = fact(paras)
		if dscv == nil {
			return nil
		}
	}

	unique := dscv.Unique(paras)
	uKey := prt + ":" + unique
	dscvC = that.dscvCs[uKey]
	if dscvC != nil {
		that.dscvCs[key] = dscvC
		return dscvC
	}

	dscvC = &discoveryC{
		mng:  that,
		dscv: dscv,
		cfg:  dscv.Cfg(unique, paras),
	}

	// 缓存协议解析
	that.dscvCs[uKey] = dscvC
	that.dscvCs[key] = dscvC
	return dscvC
}

func (that *DiscoveryMng) SetDiscoveryS(dscv string, name string, setFun func(prods []*Prod), idleTime int, panic bool) *discoveryS {
	// map初始化
	if that.dscvSs == nil {
		that.dscvSs = new(sync.Map)
	}

	// 已载入检测
	val, has := that.dscvSs.Load(name)
	dscvS, _ := val.(*discoveryS)
	if has {
		if panic {
			Kt.Panic(errors.New("dscvSs Has " + name))

		} else {
			AZap.Logger.Error("dscvSs Has " + name)
		}

		return dscvS
	}

	dscvC := that.GetDiscovery(dscv)
	if dscvC == nil {
		return nil
	}

	dscvS = &discoveryS{
		dscvC:  dscvC,
		name:   name,
		setFun: setFun,
	}

	// setAsync idleTime
	dscvS.setAsync = Util.NewNotifierAsync(dscvS.ListProds, nil, nil)
	if idleTime <= 0 {
		idleTime = dscvC.dscv.IdleTime(dscvC.cfg)
	}

	dscvS.idleTime = int64(idleTime) * int64(time.Second)
	dscvS.ctx = dscvS.dscvC.CtxProds()

	// ListProds
	dscvS.ListProds()
	// WatcherProds
	_, err := dscvS.dscvC.WatcherProds(dscvS.ctx, name, &dscvS.idleTime, dscvS.SetProds)
	if err != nil {
		if panic {
			Kt.Panic(err)

		} else {
			AZap.Logger.Error("WatcherProds SetUp Err "+name, zap.Error(err))
		}
	}

	that.dscvSs.Store(name, dscvS)
	return dscvS
}

func (that *DiscoveryMng) CheckEmpty() bool {
	return that.dscvSs == nil
}

func (that *DiscoveryMng) CheckStop() {
	that.checkTime = 0
}

func (that *DiscoveryMng) CheckLoop(checkDrt time.Duration) {
	if that.checkTime != 0 {
		return
	}

	that.locker.Lock()
	if that.checkTime != 0 {
		return
		that.locker.Unlock()
	}

	checkTime := time.Now().UnixNano()
	that.checkTime = checkTime
	that.locker.Unlock()
	for checkTime == that.checkTime {
		now := time.Now().UnixNano()
		if that.dscvSs != nil {
			that.dscvSs.Range(func(key, value interface{}) bool {
				dscvS, _ := key.(*discoveryS)
				// 定时刷新prods
				if dscvS.passTime <= now && dscvS.idleTime > 0 {
					dscvS.passTime = now + dscvS.idleTime
					dscvS.setAsync.Start(nil)
				}

				return true
			})
		}

		if that.dscvCs != nil {
			for _, dscvC := range that.dscvCs {
				if dscvC.regsAsync != nil && dscvC.regsTime <= now && GetDscvCfg().RegChkDrt > 0 {
					dscvC.regsTime = now + GetDscvCfg().RegChkDrt
					dscvC.regsAsync.Start(nil)
				}
			}
		}

		time.Sleep(checkDrt)
	}
}

// 注册默认发现类
func (that *DiscoveryMng) regDefs() {
	that.dscvs["consul"] = new(consul)
}

func (that *DiscoveryMng) RegDisc(name string, dscv Discovery) {
	that.dscvs[name] = dscv
}

func (that *DiscoveryMng) RegFact(disc string, fact func(paras string) Discovery) {
	that.facts[disc] = fact
}
