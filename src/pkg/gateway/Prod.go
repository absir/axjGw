package gateway

import (
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"github.com/apache/thrift/lib/go/thrift"
	"go.uber.org/zap"
	"math/rand"
	"strings"
	"sync"
)

var locker = new(sync.Mutex)

type Prod struct {
	// 编号
	id int32
	// 服务地址
	url string
	// 客户端
	proto thrift.TProtocol
	// 控制客户端
	aclClient *gw.AclClient
	// 转发客户端
	passClient *gw.PassClient
	// 网关客户端
	gwIClient gw.GatewayI
}

func NewProd(id int32, url string) (*Prod, error) {
	thrift.NewTTransportFactory()
	var transport thrift.TTransport = nil
	var mName = ""
	if url != "" {
		mNameI := strings.Index(url, "//")
		if mNameI > 0 {
			mName = url[mNameI+2:]
			url = url[0:mNameI]
		}

		var err error = nil
		if strings.HasPrefix(url, "http") {
			transport, err = thrift.NewTHttpClient(url)

		} else {
			transport = thrift.NewTSocketConf(url, Config.TConfig)
		}

		if err != nil {
			return nil, err
		}
	}

	prod := new(Prod)
	prod.id = id
	prod.url = url
	if transport != nil {
		var proto thrift.TProtocol = thrift.NewTCompactProtocolConf(transport, Config.TConfig)
		if mName != "" {
			proto = thrift.NewTMultiplexedProtocol(proto, mName)
		}

		prod.proto = proto
	}

	return prod, nil
}

func (that *Prod) GetAclClient() *gw.AclClient {
	if that.aclClient == nil {
		locker.Lock()
		defer locker.Unlock()
		if that.aclClient == nil {
			that.aclClient = gw.NewAclClient(thrift.NewTStandardClient(that.proto, that.proto))
		}
	}

	return that.aclClient
}

func (that *Prod) GetPassClient() *gw.PassClient {
	if that.passClient == nil {
		locker.Lock()
		defer locker.Unlock()
		if that.passClient == nil {
			that.passClient = gw.NewPassClient(thrift.NewTStandardClient(that.proto, that.proto))
		}
	}

	return that.passClient
}

func (that *Prod) GetGWIClient() gw.GatewayI {
	if that.gwIClient == nil {
		locker.Lock()
		defer locker.Unlock()
		if that.gwIClient == nil {
			if Server.gatewayI != nil {
				if that.id == Config.WorkId || that.url == "" {
					that.gwIClient = Server.gatewayI
				}
			}

			if that.gwIClient == nil {
				proto := thrift.NewTMultiplexedProtocol(that.proto, "i")
				that.gwIClient = gw.NewGatewayIClient(thrift.NewTStandardClient(proto, proto))
			}
		}
	}

	return that.gwIClient
}

type Prods struct {
	// 服务列表
	prods map[int32]*Prod
	ids   *Util.ArrayList
}

func (that *Prods) Add(id int32, url string) *Prods {
	prods := that
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

func (that *Prods) Size() int {
	return that.ids.Size()
}

func (that *Prods) GetProd(id int32) *Prod {
	return that.prods[id]
}

func (that *Prods) GetProdHash(hash int) *Prod {
	size := that.ids.Size()
	if size < 1 {
		return nil

	} else if size == 1 {
		return that.prods[that.ids.Get(0).(int32)]
	}

	id := that.ids.Get(hash % size)
	return that.prods[id.(int32)]
}

func (that *Prods) GetProdHashS(hash string) *Prod {
	if that.ids.Size() == 1 {
		return that.prods[that.ids.Get(0).(int32)]
	}

	return that.GetProdHash(Kt.HashCode(KtUnsafe.StringToBytes(hash)))
}

func (that *Prods) GetProdRand() *Prod {
	size := that.ids.Size()
	if size < 1 {
		return nil

	} else if size == 1 {
		return that.prods[that.ids.Get(0).(int32)]
	}

	id := that.ids.Get(rand.Intn(size))
	return that.prods[id.(int32)]
}
