package nps

import (
	"axj/Kt/KtFile"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"axj/Thrd/cmap"
	"axjGW/gen/gw"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"reflect"
	"strconv"
	"strings"
)

type NpsId interface {
	GetId() int
}

type NpsClient struct {
	Id     int    //编号
	Name   string //名称
	Secret string //秘钥
}

func (that *NpsClient) GetId() int {
	return that.Id
}

type NpsHost struct {
	Id       int    // 编号
	Domains  string // 域名
	ClientId int    // 客户端编号
	PAddr    string // 代理地址

	domains     []string    // 精准域名
	wildDomains []string    // 泛域名
	addrRep     *gw.AddrRep // 代理返回
}

func (that *NpsHost) GetId() int {
	return that.Id
}

func (that *NpsHost) Allow(name string, wild bool) bool {
	domains := that.domains
	if domains == nil {
		domains = make([]string, 0)
		wildDomains := make([]string, 0)
		var strs = KtStr.SplitStrS(that.Domains, ",", true, 0, false)
		for i := 0; i < len(strs); i++ {
			str := strs[i].(string)
			if strings.IndexByte(str, '*') < 0 {
				domains = append(domains, str)

			} else {
				wildDomains = append(wildDomains, strings.ReplaceAll(str, "*", ""))
			}
		}

		that.wildDomains = wildDomains
		that.domains = domains
	}

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
	Port     int    //  端口号
	ClientId int    // 客户端编号
	PAddr    string // 代理地址

	addrRep *gw.AddrRep // 代理返回
}

func (that *NpsTcp) GetId() int {
	return that.Id
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
