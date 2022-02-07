package gateway

import (
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Dscv"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"context"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Prod struct {
	// 编号
	id int32
	// 服务地址
	url string
	// 锁
	locker sync.Locker
	// 客户端
	client *grpc.ClientConn
	// 网关客户端
	gwIClient gw.GatewayIClient
	// 控制客户端
	aclClient gw.AclClient
	// 转发客户端
	passClient gw.PassClient
}

func NewProd(id int32, url string) (*Prod, error) {
	prod := new(Prod)
	prod.id = id
	prod.url = url
	prod.initClient(false, true)
	prod.locker = new(sync.Mutex)
	return prod, nil
}

func (that *Prod) Id() int32 {
	return that.id
}

func (that *Prod) initClient(locker bool, new bool) {
	if that.client != nil {
		return
	}

	if that.url == "" {
		if !new {
			AZap.Logger.Warn("InitClient Err nil")
		}

		return
	}

	if locker {
		that.locker.Lock()
		defer that.locker.Unlock()
	}

	if that.client != nil {
		return
	}

	client, err := grpc.Dial(that.url, grpc.WithInsecure())
	if err != nil {
		AZap.Logger.Warn("InitClient Err " + err.Error())
		return
	}

	that.client = client
}

func (that *Prod) GetGWIClient() gw.GatewayIClient {
	if that.gwIClient != nil {
		return that.gwIClient
	}

	if Server.gatewayISC != nil {
		if that.id == Config.WorkId {
			that.gwIClient = Server.gatewayISC
			return that.gwIClient
		}
	}

	that.initClient(true, false)
	if that.client == nil {
		return nil
	}

	if that.gwIClient == nil {
		that.locker.Lock()
		if that.gwIClient == nil {
			that.gwIClient = gw.NewGatewayIClient(that.client)
		}

		that.locker.Unlock()
	}

	return that.gwIClient
}

func (that *Prod) GetAclClient() gw.AclClient {
	if that.aclClient != nil {
		return that.aclClient
	}

	if Config.zDevAcl {
		return ZDevAcl
	}

	that.initClient(true, false)
	if that.client == nil {
		return nil
	}

	if that.aclClient == nil {
		that.locker.Lock()
		if that.aclClient == nil {
			that.aclClient = gw.NewAclClient(that.client)
		}

		that.locker.Unlock()
	}

	return that.aclClient
}

func (that *Prod) GetPassClient() gw.PassClient {
	if that.passClient != nil {
		return that.passClient
	}

	if Config.zDevAcl {
		return ZDevAcl
	}

	that.initClient(true, false)
	if that.client == nil {
		return nil
	}

	if that.passClient == nil {
		that.locker.Lock()
		if that.passClient == nil {
			that.passClient = gw.NewPassClient(that.client)
		}

		that.locker.Unlock()
	}

	return that.passClient
}

type ProdsIn struct {
	// 服务列表
	prods map[int32]*Prod
	ids   *Util.ArrayList
}

func (that *ProdsIn) addIn(id int32, url string) {
	if that.ids == nil {
		that.prods = map[int32]*Prod{}
		that.ids = Util.NewArrayList()
	}

	if _, has := that.prods[id]; has {
		AZap.Logger.Warn("AddProd Has " + strconv.Itoa(int(id)))
		return
	}

	prod, err := NewProd(id, url)
	if prod != nil {
		if that.prods[id] == nil {
			that.ids.Add(id)
		}

		that.prods[id] = prod

	} else if err != nil {
		AZap.Logger.Warn("NewProd Err", zap.Error(err))
	}
}

type Prods struct {
	*ProdsIn
	// 超时时间
	Timeout time.Duration
	// 服务发现
	Dscv string
	// 服务发现名
	DscvName string
	// 服务发现间隔
	DscvIdle int
	// 服务发现成功
	inited bool
	// 模块服务查询
	PassUrl string
	// 模块编号地址
	IdUrl string
	// 模块编号数量
	IdCount int
}

func (that *Prods) TimeoutCtx() context.Context {
	if that.Timeout <= 0 {
		return Server.Context
	}

	ctx, _ := context.WithTimeout(Server.Context, that.Timeout)
	return ctx
}

func (that *Prods) Set(prods []*Dscv.Prod) {
	if prods == nil {
		return
	}

	prodsIn := new(ProdsIn)
	for _, prod := range prods {
		prodsIn.addIn(KtCvt.ToInt32(prod.Id), prod.Addr+":"+strconv.Itoa(prod.GetPort(Config.ProdPortKey, Config.ProdPort)))
	}

	that.ProdsIn = prodsIn
}

func (that *Prods) Add(id int32, url string) *Prods {
	prods := that
	if prods == nil {
		prods = new(Prods)
	}

	if prods.ProdsIn == nil {
		prods.ProdsIn = new(ProdsIn)
	}

	prods.ProdsIn.addIn(id, url)
	return prods
}

func (that *Prods) init() {
	if that.inited {
		return
	}

	if that.PassUrl != "" {
		// 查询服务地址
		prod, err := NewProd(0, that.PassUrl)
		Kt.Panic(err)
		var rep *gw.ProdsRep
		rep, err = prod.GetPassClient().Prods(that.TimeoutCtx(), &gw.Void{})
		Kt.Panic(err)
		that.SetProdsRep(rep)

	} else if that.IdUrl != "" && that.IdCount > 0 {
		// 有状态服务地址
		prodsIn := new(ProdsIn)
		for id := 0; id < that.IdCount; id++ {
			prodsIn.addIn(int32(id), strings.ReplaceAll(that.IdUrl, "$id", strconv.Itoa(id)))
		}

		that.ProdsIn = prodsIn
	}

	that.inited = true
}

func (that *Prods) SetProdsRep(prods *gw.ProdsRep) {
	prodsIn := new(ProdsIn)
	for _, prod := range prods.Prods {
		prodsIn.addIn(prod.Id, prod.Uri)
	}

	that.ProdsIn = prodsIn
}

func (that *Prods) Size() int {
	that.init()
	if that.ProdsIn == nil {
		return 0
	}

	return that.ids.Size()
}

func (that *Prods) GetProd(id int32) *Prod {
	that.init()
	prodsIn := that.ProdsIn
	if prodsIn == nil {
		return nil
	}

	return prodsIn.prods[id]
}

func (that *Prods) GetProdHash(hash int) *Prod {
	that.init()
	prodsIn := that.ProdsIn
	if prodsIn == nil {
		return nil
	}

	size := prodsIn.ids.Size()
	if size < 1 {
		return nil

	} else if size == 1 {
		return prodsIn.prods[prodsIn.ids.Get(0).(int32)]
	}

	id := prodsIn.ids.Get(hash % size)
	return prodsIn.prods[id.(int32)]
}

func (that *Prods) GetProdHashS(hash string) *Prod {
	that.init()
	prodsIn := that.ProdsIn
	if prodsIn == nil {
		return nil
	}

	if prodsIn.ids.Size() == 1 {
		return prodsIn.prods[prodsIn.ids.Get(0).(int32)]
	}

	return that.GetProdHash(Kt.HashCode(KtUnsafe.StringToBytes(hash)))
}

func (that *Prods) GetProdRand() *Prod {
	that.init()
	prodsIn := that.ProdsIn
	if prodsIn == nil {
		return nil
	}

	size := prodsIn.ids.Size()
	if size < 1 {
		return nil

	} else if size == 1 {
		return prodsIn.prods[prodsIn.ids.Get(0).(int32)]
	}

	id := prodsIn.ids.Get(rand.Intn(size))
	return prodsIn.prods[id.(int32)]
}
