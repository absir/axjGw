package main

import (
	"axj/APro"
	"axj/Kt/Kt"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/pkg/agent"
	"axjGW/pkg/asdk"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	snowboy "github.com/brentnd/go-snowboy"
	"github.com/gordonklaus/portaudio"
	"io/ioutil"
	"net"
	"os"
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
	// 音频相关
	inputChannels  int // 输出频道
	outputChannels int // 输出频道
	sampleRate     int // 采样率
}

var Config = &config{
	Proxy:          "127.0.0.1:8783",
	ProxyHash:      false,
	SendP:          true,
	ReadP:          true,
	Encry:          true,
	CompressMin:    1024,
	DataMax:        1024 << 4,
	CheckDrt:       10,
	RqIMax:         0,
	ConnDrt:        30,
	CloseDelay:     30,
	inputChannels:  1,
	outputChannels: 0,
	sampleRate:     16000,
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
	APro.Load(nil, "vasset.yml")
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

	// 监听唤醒词
	go Listen()

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
	// 自定义reqId处理
	}

	// 内存池释放
	asdk.BufferFree(buffer)
	fmt.Println("OnReserve " + strconv.Itoa(int(req)) + ", " + uri + ", " + strconv.Itoa(int(uriI)))
}

// Sound represents a sound stream with io.Reader interface.
type Sound struct {
	stream *portaudio.Stream
	data   []int16
}

// Read is the implementation of the io.Reader interface.
func (s *Sound) Read(p []byte) (int, error) {
	s.stream.Read()

	buf := &bytes.Buffer{}
	for _, v := range s.data {
		binary.Write(buf, binary.LittleEndian, v)
	}

	copy(p, buf.Bytes())
	return len(p), nil
}

func Listen() {
	framesPerBuffer := make([]int16, Config.sampleRate)
	// initialize the audio recording interface
	err := portaudio.Initialize()
	if err != nil {
		fmt.Errorf("Error initialize audio interface: %s", err)
		return
	}

	defer portaudio.Terminate()

	// open the sound input for the microphone
	stream, err := portaudio.OpenDefaultStream(Config.inputChannels, Config.outputChannels, float64(Config.sampleRate), len(framesPerBuffer), framesPerBuffer)
	if err != nil {
		fmt.Errorf("Error open default audio stream: %s", err)
		return
	}
	defer stream.Close()

	// open the snowboy detector
	d := snowboy.NewDetector(os.Args[1])
	defer d.Close()

	d.HandleFunc(snowboy.NewHotword(os.Args[2], 0.5), func(string) {
		fmt.Println("Handle func for snowboy Hotword")
	})

	d.HandleSilenceFunc(500*time.Millisecond, func(string) {
		fmt.Println("Silence detected")
	})

	sr, nc, bd := d.AudioFormat()
	fmt.Printf("sample rate=%d, num channels=%d, bit depth=%d\n", sr, nc, bd)

	err = stream.Start()
	if err != nil {
		fmt.Errorf("Error on stream start: %s", err)
		return
	}

	sound := &Sound{stream, framesPerBuffer}

	d.ReadAndDetect(sound)
}
