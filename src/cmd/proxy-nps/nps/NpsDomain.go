package nps

import (
	"axj/Kt/KtFile"
	"axj/Kt/KtRand"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"axjGW/pkg/proxy"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type NpsId interface {
	GetId() int
	SetId(id int)
}

type NpsClient struct {
	Id     int    //编号
	Name   string //名称
	Secret string //秘钥

	Cid int64 //连接编号
}

func (that *NpsClient) GetId() int {
	return that.Id
}

func (that *NpsClient) SetId(id int) {
	that.Id = id
}

type NpsHost struct {
	Id       int    // 编号
	Domains  string // 域名
	ClientId int    // 客户端编号
	PAddr    string // 代理地址

	domains     []string    `json:"-"` // 精准域名
	wildDomains []string    `json:"-"` // 泛域名
	addrRep     *gw.AddrRep `json:"-"` // 代理返回
}

func (that *NpsHost) GetId() int {
	return that.Id
}

func (that *NpsHost) SetId(id int) {
	that.Id = id
}

// npsHost缓存加速
var npsHostCurr = 0
var npsHostDirty = 1
var npsHostMap map[string]*NpsHost = nil
var npsHostWilds []*NpsHost = nil

func HostAddrRep(name string) *gw.AddrRep {
	if npsHostCurr != npsHostDirty {
		dirty := npsHostDirty
		hostMap := make(map[string]*NpsHost)
		hostWilds := make([]*NpsHost, 0)
		HostMap.Range(func(key, value interface{}) bool {
			npsHost, _ := value.(*NpsHost)
			if npsHost != nil {
				domains := npsHost.initDomains()
				for _, domain := range domains {
					hostMap[domain] = npsHost
				}

				if npsHost.wildDomains != nil {
					hostWilds = append(hostWilds, npsHost)
				}
			}

			return true
		})

		npsHostMap = hostMap
		npsHostWilds = hostWilds
		npsHostCurr = dirty
	}

	host := npsHostMap[name]
	if host != nil {
		return host.AddrRep()
	}

	wilds := npsHostWilds
	for i := len(wilds) - 1; i >= 0; i-- {
		host = npsHostWilds[i]
		wildDomains := host.wildDomains
		if wildDomains != nil {
			for j := len(wildDomains) - 1; j >= 0; j-- {
				if strings.Index(name, wildDomains[j]) >= 0 {
					return host.AddrRep()
				}
			}
		}
	}

	return nil
}

func (that *NpsHost) initDomains() []string {
	domains := that.domains
	if domains == nil {
		domains = make([]string, 0)
		var wildDomains []string = nil
		var strs = KtStr.SplitStrS(that.Domains, ",", true, 0, false)
		for i := 0; i < len(strs); i++ {
			str := strs[i].(string)
			if strings.IndexByte(str, '*') < 0 {
				domains = append(domains, str)

			} else {
				if wildDomains == nil {
					wildDomains = make([]string, 0)
				}

				wildDomains = append(wildDomains, strings.ReplaceAll(str, "*", ""))
			}
		}

		that.wildDomains = wildDomains
		that.domains = domains
	}

	return domains
}

func (that *NpsHost) Allow(name string, wild bool) bool {
	domains := that.initDomains()
	if wild {
		wildDomains := that.wildDomains
		if wildDomains != nil {
			for _, wildDomain := range wildDomains {
				if strings.Index(name, wildDomain) >= 0 {
					return true
				}
			}
		}

	} else {
		for _, domain := range domains {
			if name == domain {
				return true
			}
		}
	}

	return false
}

func (that *NpsHost) AddrRep() *gw.AddrRep {
	addrRep := that.addrRep
	if addrRep == nil {
		addrRep = &gw.AddrRep{
			Gid:  strconv.Itoa(that.ClientId),
			Addr: that.PAddr,
		}

		that.addrRep = addrRep
	}

	return addrRep
}

type NpsTcp struct {
	Id       int    // 编号
	Addr     string // 服务地址
	ClientId int    // 客户端编号
	PAddr    string // 代理地址

	addrRep *gw.AddrRep    `json:"-"` // 代理返回
	serv    *proxy.PrxServ `json:"-"` // 代理服务
}

func (that *NpsTcp) GetId() int {
	return that.Id
}

func (that *NpsTcp) SetId(id int) {
	that.Id = id
}

func (that *NpsTcp) AddrRep() *gw.AddrRep {
	addrRep := that.addrRep
	if addrRep == nil {
		addrRep = &gw.AddrRep{
			Gid:  strconv.Itoa(that.ClientId),
			Addr: that.PAddr,
		}

		that.addrRep = addrRep
	}

	return addrRep
}

var ClientMap = cmap.NewCMapInit()
var HostMap = cmap.NewCMapInit()
var TcpMap = cmap.NewCMapInit()

type NpsIdSlice []NpsId

func (s NpsIdSlice) Len() int {
	return len(s)
}

func (s NpsIdSlice) Less(i, j int) bool {
	return s[i].GetId() < s[j].GetId()
}

func (s NpsIdSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func ReadList(cmap *cmap.CMap, _sort bool) []NpsId {
	npsIds := make([]NpsId, 0)
	cmap.Range(func(key, value interface{}) bool {
		npsId, _ := value.(NpsId)
		if npsId != nil {
			npsIds = append(npsIds, npsId)
		}

		return true
	})

	if _sort {
		sort.Sort(NpsIdSlice(npsIds))
	}

	return npsIds
}

func MapDel(cmap *cmap.CMap, id int) {
	if cmap == nil {
		return
	}

	value, loaded := cmap.LoadAndDelete(id)
	if !loaded {
		return
	}

	MapSave(cmap)
	MapDirty(cmap, value, nil, true)
}

func MapDirty(cmap *cmap.CMap, value interface{}, old interface{}, del bool) {
	if cmap == ClientMap {
		npsClient, _ := value.(*NpsClient)
		if npsClient.Secret == "" {
			secret := ""
			for secret == "" {
				secret = KtRand.RandString(12, RandChars)
				cmap.Range(func(key, value interface{}) bool {
					client, _ := value.(*NpsClient)
					if secret == client.Secret {
						secret = ""
						return false
					}

					return true
				})
			}

			// 自动添加秘钥
			npsClient.Secret = secret
		}

		npsClientO, _ := old.(*NpsClient)
		if npsClientO == nil || npsClientO.Secret != npsClient.Secret {
			client := proxy.PrxMng.Client(0, strconv.Itoa(npsClient.Id))
			if client != nil {
				// 踢出重新授权
				npsClient.Cid = 0
				client.Get().Kick(nil, false, 0)
			}
		} else if npsClientO != nil && npsClientO.Cid != 0 {
			npsClient.Cid = npsClientO.Cid
		}

	} else if cmap == HostMap {
		npsHost, _ := value.(*NpsHost)
		if npsHost != nil {
			// 更新代理地址
			npsHost.addrRep = nil
			if npsHostDirty >= math.MaxInt {
				npsHostDirty = 0

			} else {
				npsHostDirty++
			}
		}

	} else if cmap == TcpMap {
		npsTcp, _ := value.(*NpsTcp)
		id := strconv.Itoa(npsTcp.Id)
		if old == nil && del {
			old, _ = TcpMap.Load(id)
		}

		npsTcpO, _ := old.(*NpsTcp)
		if npsTcp != nil {
			npsTcp.addrRep = nil
		}

		if del {
			if npsTcpO != nil && npsTcpO.serv != nil {
				// 服务删除
				npsTcpO.serv.Close()
			}

		} else {
			if npsTcpO.serv == nil || npsTcpO == nil || npsTcpO.Addr != npsTcp.Addr {
				if npsTcpO.serv != nil {
					// 关闭旧服务
					npsTcpO.serv.Close()
				}

				// 开启新服务
				addr := npsTcpO.Addr
				if strings.IndexByte(addr, ':') < 0 {
					addr = "0.0.0.0:" + addr
				}

				npsTcp.serv = proxy.StartServ(id, addr, 0, proxy.FindProto("tcp", true), nil)

			} else {
				// 旧服务赋值
				npsTcp.serv = npsTcpO.serv
			}
		}
	}
}

func MapSave(cmap *cmap.CMap) {
	if cmap == ClientMap {
		mapSave(cmap, "client.json")

	} else if cmap == HostMap {
		mapSave(cmap, "host.json")

	} else if cmap == TcpMap {
		mapSave(cmap, "tcp.json")
	}
}

func LoadAll() {
	loadSave(ClientMap, "client.json", make([]*NpsClient, 0))
	loadSave(HostMap, "host.json", make([]*NpsHost, 0))
	loadSave(TcpMap, "tcp.json", make([]*NpsTcp, 0))
}

func loadSave(cmap *cmap.CMap, saveFile string, npsIds interface{}) {
	file := KtFile.Open("save/" + saveFile)
	if file == nil {
		return
	}

	defer file.Close()
	bs, _ := io.ReadAll(file)
	json.Unmarshal(bs, npsIds)
	value := reflect.ValueOf(npsIds)
	for i := 0; i < value.Len(); i++ {
		npsId, _ := value.Index(i).Interface().(NpsId)
		if npsId != nil {
			cmap.Store(npsId.GetId(), npsId)
		}
	}

	// cmap逻辑加载
	cmap.Range(func(key, value interface{}) bool {
		MapDirty(cmap, value, nil, false)
		return true
	})
}

func mapSave(cmap *cmap.CMap, saveFile string) {
	// 打开文件，如果文件不存在则创建，如果文件已存在则截断为零
	file := KtFile.Create("save/"+saveFile, false)
	if file == nil {
		return
	}

	defer file.Close()
	file.WriteString("[")
	cmap.Range(func(key, value interface{}) bool {
		bs, err := json.Marshal(value)
		if err != nil {
			AZap.Warn("save err "+saveFile, zap.Error(err))
		}

		if bs != nil {
			file.WriteString("\r\n")
			file.Write(bs)
		}

		return true
	})

	file.WriteString("\r\n]")
}
