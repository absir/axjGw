package main

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/pkg/agent"
	"axjGW/pkg/asdk"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type config struct {
	Proxy       string // 代理地址
	ProxyHash   bool   // 代理地址一致性hash
	ClientKey   string // 客户端Key
	ClientCert  string // 客户端证书
	ClientId    string // 客户端唯一编号
	SendP       bool
	ReadP       bool
	Encry       bool
	CompressMin int
	DataMax     int
	CheckDrt    int
	RqIMax      int
	ConnDrt     time.Duration
	CloseDelay  int // 关闭延迟秒数
	Rules       map[string]*agent.RULE
}

var Config = &config{
	Proxy:       "127.0.0.1:8783",
	ProxyHash:   false,
	SendP:       true,
	ReadP:       true,
	Encry:       true,
	CompressMin: 1024,
	DataMax:     1024 << 4,
	CheckDrt:    10,
	RqIMax:      0,
	ConnDrt:     30,
	CloseDelay:  30,
}

var Machineid string
var Client *asdk.Client

const (
	CERT_FILE    = "client.cert"
	BIN_MAC_FILE = "bin/client.mac"
)

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../resources")
	APro.Load(nil, "agent.yml")
	loadConfig()
	Config.ConnDrt *= time.Second
	// 内存池
	Util.SetBufferPoolsS(APro.GetCfg("bPools", KtCvt.String, "256,512,1024,5120,10240").(string))
	Client = asdk.NewClient(Config.Proxy, Config.SendP, Config.ReadP, Config.Encry, Config.CompressMin, Config.DataMax, Config.CheckDrt, Config.RqIMax, &Opt{})
	if Config.ProxyHash {
		Client.AddrHash = Kt.HashCode(KtUnsafe.StringToBytes(Machineid))
	}

	agent.Client = Client
	agent.CloseDelay = Config.CloseDelay
	go func() {
		for !APro.Stopped {
			// 保持连接
			Client.Conn()
			time.Sleep(Config.ConnDrt)
		}
	}()
	// 启动完成
	AZap.Info("Agent %s all AXJ started", agent.Version)
	APro.Signal()
}

func loadConfig() {
	KtCvt.BindInterface(Config, APro.Cfg)
	f := APro.Open(CERT_FILE)
	if f != nil {
		data, err := ioutil.ReadAll(f)
		Kt.Panic(err)
		Config.ClientCert = KtUnsafe.BytesToString(data)
	}

	// 获取本机的MAC地址
	f = APro.Open(BIN_MAC_FILE)
	if f != nil {
		data, err := ioutil.ReadAll(f)
		Kt.Panic(err)
		Machineid = KtUnsafe.BytesToString(data)
	}

	if Machineid == "" {
		inters, err := net.Interfaces()
		Kt.Panic(err)
		if inters != nil {
			for _, inter := range inters {
				Machineid = strings.ReplaceAll(inter.HardwareAddr.String(), ":", "")
				if Machineid != "" {
					f = APro.Create(BIN_MAC_FILE, false)
					f.WriteString(Machineid)
					f.Close()
					break
				}
			}
		}
	}
}

type Opt struct {
}

func (o Opt) LoadStorage(name string) string {
	return ""
}

func (o Opt) SaveStorage(name string, value string) {
}

func (o Opt) LoginData(adapter *asdk.Adapter) []byte {
	data, err := json.Marshal([]string{Config.ClientKey, Config.ClientCert, Config.ClientId, Machineid, agent.Version})
	Kt.Err(err, true)
	return data
}

func (o Opt) OnPush(uri string, data []byte, tid int64, buffer asdk.Buffer) {
	fmt.Println("OnPush " + uri + ", " + strconv.FormatInt(tid, 10))
	asdk.BufferFree(buffer)
}

func (o Opt) OnLast(gid string, connVer int32, continues bool) {
	fmt.Println("OnLast " + gid + ", " + strconv.Itoa(int(connVer)) + ", " + strconv.FormatBool(continues))
}

func (o Opt) OnState(adapter *asdk.Adapter, state int, err string, data []byte, buffer asdk.Buffer) {
	fmt.Println("OnState , " + strconv.Itoa(state) + ", " + err)
	asdk.BufferFree(buffer)
}

func (o Opt) OnReserve(adapter *asdk.Adapter, req int32, uri string, uriI int32, data []byte, buffer asdk.Buffer) {
	switch req {
	case agent.REQ_DIAL:
		// 连接代理
		var timeout int64 = 0
		if data != nil {
			timeout = KtBytes.GetInt64(data, 0, nil)
		}

		go agent.DialProxy(uri, uriI, time.Duration(timeout))
		return
	case agent.REQ_CONN:
		// 发送连接
		go agent.ConnProxy(uri, uriI, data, buffer)
		return
	case agent.REQ_RULES:
		// 本地映射配置
		if Config.Rules != nil {
			bs, err := json.Marshal(Config.Rules)
			if err != nil {
				Kt.Err(err, false)

			} else if bs != nil {
				adapter.Rep(Client, agent.REQ_RULES, "", 0, bs, false, 0)
			}
		}

		// 内存池释放
		asdk.BufferFree(buffer)
		return
	case agent.REQ_ON_RULE:
		// 映射规则
		fmt.Println("OnRule " + uri)
		// 内存池释放
		asdk.BufferFree(buffer)
		return
	}

	// 内存池释放
	asdk.BufferFree(buffer)
	fmt.Println("OnReserve " + strconv.Itoa(int(req)) + ", " + uri + ", " + strconv.Itoa(int(uriI)))
}
