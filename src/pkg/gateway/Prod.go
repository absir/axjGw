package gateway

import (
	"axj/Kt/Kt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/gen/gw"
	"github.com/apache/thrift/lib/go/thrift"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"math/rand"
	"strconv"
	"sync"
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
	aclClient *gw.AclClient
	// 转发客户端
	passClient *gw.PassClient
}

func NewProd(id int32, url string) (*Prod, error) {
	prod := new(Prod)
	prod.id = id
	prod.url = url
	err := prod.initClient(false)
	prod.locker = new(sync.Mutex)
	if err != nil {
		AZap.Logger.Debug("NewProd init err " + strconv.Itoa(int(id)) + ", " + url + " : " + err.Error())
	}

	return prod, nil
}

func (that *Prod) initClient(locker bool) error {
	if that.url == "" {
		return nil
	}

	if that.client != nil {
		return nil
	}

	if locker {
		that.locker.Lock()
		defer that.locker.Unlock()
	}

	if that.client != nil {
		return nil
	}

	that.gwIClient = nil
	that.aclClient = nil
	that.passClient = nil
	client, err := grpc.Dial(that.url)
	if err != nil {
		return err
	}

	that.client = client
	return nil
}

func (that *Prod) GetGWIClient() gw.GatewayIClient {
	that.initClient(true)
	if that.gwIClient == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.gwIClient == nil {
			if Server.gatewayI != nil {
				if that.id == Config.WorkId || that.url == "" {
					that.gwIClient = Server.gatewayI
				}
			}

			if that.gwIClient == nil {
				gw.new
				proto := thrift.NewTMultiplexedProtocol(that.proto, "i")
				that.gwIClient = gw.NewGatewayIClient(thrift.NewTStandardClient(proto, proto))
			}
		}
	}

	return that.gwIClient
}

func (that *Prod) GetAclClient() *gw.AclClient {
	that.initClient(true)
	if that.aclClient == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.aclClient == nil {
			that.aclClient = gw.NewAclClient(thrift.NewTStandardClient(that.proto, that.proto))
		}
	}

	return that.aclClient
}

func (that *Prod) GetPassClient() *gw.PassClient {
	that.initClient(true)
	if that.passClient == nil {
		that.locker.Lock()
		defer that.locker.Unlock()
		if that.passClient == nil {
			that.passClient = gw.NewPassClient(thrift.NewTStandardClient(that.proto, that.proto))
		}
	}

	return that.passClient
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
	}

	if prods.ids == nil {
		prods.prods = map[int32]*Prod{}
		prods.ids = Util.NewArrayList()
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
