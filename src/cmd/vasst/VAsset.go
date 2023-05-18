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
	"encoding/json"
	"fmt"
	porcupine "github.com/Picovoice/porcupine/binding/go/v2"
	pvrecorder "github.com/Picovoice/pvrecorder/sdk/go"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/gen2brain/malgo"
	"io/ioutil"
	"log"
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
	inputChannels:  -1,
	outputChannels: -1,
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

	// 打印音频设备
	printAudioDevices()

	// 播放声音
	f, err := os.Open("E:\\doc\\恐龙\\upan\\success.mp3")
	if err != nil {
		log.Fatal(err)
	}

	streamer, format, err := mp3.Decode(f)
	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done

	// 监听唤醒词
	go ListenKeywords()

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

func printAudioDevices() {
	context, err := malgo.InitContext(nil, malgo.ContextConfig{}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = context.Uninit()
		context.Free()
	}()

	// Capture devices.
	infos, err := context.Devices(malgo.Capture)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Capture Devices")
	for i, info := range infos {
		fmt.Printf("    %d: %s\n", i, strings.Replace(info.Name(), "\x00", "", -1))
	}
}

func ListenKeywords() {
	builtInKeyword := porcupine.BuiltInKeyword("hey siri")
	fmt.Printf("", builtInKeyword)

	p := porcupine.Porcupine{}
	p.AccessKey = ""
	p.BuiltInKeywords = append(p.BuiltInKeywords, builtInKeyword)
	p.Sensitivities = []float32{1}

	err := p.Init()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := p.Delete()
		if err != nil {
			log.Fatalf("Failed to release resources: %s", err)
		}
	}()

	recorder := pvrecorder.PvRecorder{
		DeviceIndex:    Config.inputChannels,
		FrameLength:    porcupine.FrameLength,
		BufferSizeMSec: 1000,
		LogOverflow:    0,
	}

	if err := recorder.Init(); err != nil {
		log.Fatalf("Error: %s.\n", err.Error())
	}
	defer recorder.Delete()

	log.Printf("Using device: %s", recorder.GetSelectedDevice())

	if err := recorder.Start(); err != nil {
		log.Fatalf("Error: %s.\n", err.Error())
	}

	log.Printf("Listening...")

	for {
		pcm, err := recorder.Read()
		if err != nil {
			log.Fatalf("Error: %s.\n", err.Error())
		}

		keywordIndex, err := p.Process(pcm)
		if keywordIndex >= 0 {
			fmt.Printf("keywordIndex = %d\n", keywordIndex)
		}
	}

}
