package gateway

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"github.com/apache/thrift/lib/go/thrift"
	"go.uber.org/zap"
	"gw"
	"math/rand"
	"strings"
	"sync"
)

var locker = new(sync.Mutex)

type Prod struct {
	// 编号
	id int32
	// 转发地址
	url string
	// 客户端
	client thrift.TClient
	// 控制客户端
	aclClient *gw.AclClient
	// 转发客户端
	passClient *gw.PassClient
	// 网关客户端
	gwIClient *gw.GatewayIClient
}

func NewProd(id int32, url string) (*Prod, error) {
	thrift.NewTTransportFactory()
	var transport thrift.TTransport
	var err error

	if strings.HasPrefix(url, "http") {
		transport, err = thrift.NewTHttpClient(url)

	} else {
		transport, err = thrift.NewTSocketConf(url, Config.TConfig)
	}

	if err != nil {
		return nil, err
	}

	prod := new(Prod)
	prod.id = id
	prod.url = url
	proto := thrift.NewTCompactProtocolConf(transport, Config.TConfig)
	prod.client = thrift.NewTStandardClient(proto, proto)
	prod.passClient = nil
	prod.gwIClient = nil
	return prod, nil
}

func (that Prod) GetAclClient() *gw.AclClient {
	if that.aclClient == nil {
		locker.Lock()
		defer locker.Unlock()
		if that.aclClient == nil {
			that.aclClient = gw.NewAclClient(that.client)
		}
	}

	return that.aclClient
}

func (that Prod) GetPassClient() *gw.PassClient {
	if that.passClient == nil {
		locker.Lock()
		defer locker.Unlock()
		if that.passClient == nil {
			that.passClient = gw.NewPassClient(that.client)
		}
	}

	return that.passClient
}

func (that Prod) GetGWIClient() *gw.GatewayIClient {
	if that.gwIClient == nil {
		locker.Lock()
		defer locker.Unlock()
		if that.gwIClient == nil {
			that.gwIClient = gw.NewGatewayIClient(that.client)
		}
	}

	return that.gwIClient
}

type Prods struct {
	// 服务列表
	prods map[int32]*Prod
	ids   *Util.ArrayList
}

func BuildProds(prods *Prods, id int32, url string) *Prods {
	if prods == nil {
		prods = new(Prods)
		prods.prods = map[int32]*Prod{}
		prods.ids = new(Util.ArrayList)
	}

	prod, err := NewProd(id, url)
	if prod != nil {
		if prods.prods[id] == nil {
			prods.ids.Add(id)
		}

		prods.prods[id] = prod

	} else if err != nil {
		AZap.Logger.Info("NewProd err", zap.Error(err))
	}

	return prods
}

func (that Prods) GetProd(id int32) *Prod {
	return that.prods[id]
}

func (that Prods) GetProdHash(hash int) *Prod {
	size := that.ids.Size()
	if size < 1 {
		return nil

	} else if size == 1 {
		return that.prods[that.ids.Get(0).(int32)]
	}

	id := that.ids.Get(hash % size)
	return that.prods[id.(int32)]
}

func (that Prods) GetProdRand() *Prod {
	size := that.ids.Size()
	if size < 1 {
		return nil

	} else if size == 1 {
		return that.prods[that.ids.Get(0).(int32)]
	}

	id := that.ids.Get(rand.Intn(size))
	return that.prods[id.(int32)]
}

var prodsMap = new(sync.Map)

func RegProds(name string, prods *Prods) {
	if prods == nil {
		prodsMap.Delete(name)

	} else {
		prodsMap.Store(name, prods)
	}
}

func GetProds(name string) *Prods {
	val, _ := prodsMap.Load(name)
	if val == nil {
		return nil
	}

	return val.(*Prods)
}
