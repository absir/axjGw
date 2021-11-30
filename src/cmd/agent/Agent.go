package main

import (
	"axj/ANet"
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axjGW/pkg/agent"
	"axjGW/pkg/asdk"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

type config struct {
	Proxy       string // 代理地址
	ClientKey   string // 客户端Key
	ClientCert  string // 客户端证书
	ClientId    string // 客户端唯一编号
	Out         bool
	Encry       bool
	CompressMin int
	DataMax     int
	CheckDrt    int
	RqIMax      int
	ConnDrt     int
	Rules       map[string]*agent.RULE
}

var Config = &config{
	Proxy:       "127.0.0.1:8783",
	Out:         true,
	Encry:       true,
	CompressMin: 1024,
	DataMax:     1024 << 4,
	CheckDrt:    10,
	RqIMax:      0,
	ConnDrt:     30,
}

var Client *asdk.Client

const (
	CLIENT_CERT_FILE = "client.cert"
)

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../../resources")
	APro.Load(nil, "agent.yml")
	loadConfig()

	Client = asdk.NewClient(Config.Proxy, Config.Out, Config.Encry, Config.CompressMin, Config.DataMax, Config.CheckDrt, Config.RqIMax, &Opt{})
	connDrt := time.Duration(Config.ConnDrt) * time.Second
	go func() {
		for !APro.Stopped {
			// 保持连接
			Client.Conn()
			time.Sleep(connDrt)
		}
	}()
	APro.Signal()
}

func loadConfig() {
	KtCvt.BindInterface(Config, APro.Cfg)
	f := APro.Open(CLIENT_CERT_FILE)
	if f != nil {
		data, err := ioutil.ReadAll(f)
		Kt.Panic(err)
		Config.ClientCert = KtUnsafe.BytesToString(data)
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
	hardAddr := ""
	// 获取本机的MAC地址
	inters, err := net.Interfaces()
	if inters != nil {
		for _, inter := range inters {
			hardAddr = strings.ReplaceAll(inter.HardwareAddr.String(), ":", "")
			break
		}
	}

	data, err := json.Marshal([]string{Config.ClientKey, Config.ClientCert, Config.ClientId, hardAddr})
	Kt.Err(err, true)
	return data
}

func (o Opt) OnPush(uri string, data []byte, tid int64) {
	fmt.Println("OnPush " + uri + ", " + strconv.FormatInt(tid, 10))
}

func (o Opt) OnLast(gid string, connVer int32, continues bool) {
	fmt.Println("OnLast " + gid + ", " + strconv.Itoa(int(connVer)) + ", " + strconv.FormatBool(continues))
}

func (o Opt) OnState(adapter *asdk.Adapter, state int, err string, data []byte) {
	fmt.Println("OnState , " + strconv.Itoa(state) + ", " + err)
}

func (o Opt) OnReserve(adapter *asdk.Adapter, req int32, uri string, uriI int32, data []byte) {
	switch req {
	case agent.REQ_CONN:
		// 发送连接
		go connProxy(uri, uriI, data)
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

		return
	case agent.REQ_ON_RULE:
		// 映射规则
		fmt.Println("OnRule " + uri)
		return
	}

	fmt.Println("OnReserve " + strconv.Itoa(int(req)) + ", " + uri + ", " + strconv.Itoa(int(uriI)))
}

type ConnId struct {
	id      int32
	locker  sync.Locker
	closed  bool
	conn    *net.TCPConn
	aConn   ANet.Conn
	aConned bool
}

func (that *ConnId) onError(err error) bool {
	if err == nil {
		return false
	}

	if that.closed {
		return true
	}

	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		return true
	}

	that.closed = true
	conn := that.conn
	if conn != nil {
		conn.SetLinger(0)
		conn.Close()
	}

	aConn := that.aConn
	if aConn != nil {
		aConn.Close()
	}

	that.locker.Unlock()
	if err == io.EOF {
		AZap.Debug("ConnId EOF %d", that.id)

	} else {
		Kt.Err(err, true)
	}

	adap := Client.Conn()
	if adap != nil {
		//  关闭通知
		adap.Rep(Client, agent.REQ_CLOSED, "", that.id, nil, false, 0)
	}

	return true
}

func connProxy(addr string, id int32, data []byte) {
	connId := &ConnId{
		id:     id,
		locker: new(sync.Mutex),
	}

	// 结束关闭
	defer connClose(connId)
	// 代理缓冲大小
	buffSize := 0
	{
		idx := strings.IndexByte(addr, '/')
		if idx >= 0 {
			buffSize = int(KtCvt.ToInt32(addr[:idx]))
			addr = addr[idx+1:]
		}

		if buffSize < 256 {
			buffSize = 256
		}
	}

	// 本地连接
	conn, err := net.Dial("tcp", addr)
	if connId.onError(err) || conn == nil {
		return
	}

	connId.conn = conn.(*net.TCPConn)

	// 代理连接
	var aConn ANet.Conn = nil
	{
		{
			conn, err := Client.DialConn()
			if connId.onError(err) || conn == nil {
				return
			}

			aConn, _ = conn.(ANet.Conn)
		}

		connId.aConn = aConn
		// 代理连接协议
		aProcessor, _ := Client.GetProcessor().(*ANet.Processor)
		if aConn == nil || aProcessor == nil {
			return
		}

		// 代理连接连接id请求
		err = aProcessor.Rep(nil, true, aConn, nil, false, agent.REQ_CONN, "", id, nil, false, 0)
		if connId.onError(err) {
			return
		}

		connId.aConned = true

		// conn本地连接 数据接收 写入到 aConn代理连接
		go func() {
			buff := make([]byte, buffSize)
			for !connId.closed {
				n, err := conn.Read(buff)
				if connId.onError(err) {
					return
				}

				if n > 0 {
					aConn.Write(buff[:n])
				}
			}
		}()

		// 缓冲数据写入到conn本地连接
		conn.Write(data)
		// data gc
		data = nil
	}

	// aConn代理连接 数据接收 写入到 conn本地连接
	var buff []byte = nil
	for !connId.closed || true {
		// 循环写入
		err, data, reader := aConn.ReadA()
		if connId.onError(err) {
			return
		}

		if reader == nil {
			conn.Write(data)

		} else {
			if buff == nil {
				buff = make([]byte, buffSize)
			}

			n, err := reader.Read(buff)
			if connId.onError(err) {
				return
			}

			if n > 0 {
				conn.Write(buff[:n])
			}
		}
	}
}

func connClose(connId *ConnId) {
	// 关闭连接
	connId.onError(io.EOF)
	if !connId.aConned {
		// 关闭通知
		Client.Req("", KtBytes.GetVIntBytes(connId.id), false, 60, nil)
	}
}
