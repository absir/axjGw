package gateway

import (
	"axj/APro"
	"axj/Kt"
	"axj/KtCvt"
	"axj/KtStr"
	"container/list"
	"github.com/apache/thrift/lib/go/thrift"
	"go.etcd.io/etcd/client/v3"
	"net"
	"runtime"
	"time"
)

type Config struct {
	addr    string
	addrPub string
	etcd    string
}

var cfg = Config{}

func main() {
	// 初始化配置
	APro.Caller(func(skip int) (pc uintptr, file string, line int, ok bool) {
		return runtime.Caller(0)
	}, "../public")
	APro.Load(nil, "config.yaml")

	// 默认配置
	cfg.addr = "127.0.0.1:8082"
	cfg.etcd = ""

	// etcd注册
	client, err := clientv3.New(
		APro.SubCfgBind(
			"etcd.config",
			clientv3.Config{
				Endpoints: KtCvt.ToArray(KtStr.SplitStrBr(cfg.etcd, ",;", true, 0, false, 0, false).(*list.List), KtCvt.String).([]string),
			}).(clientv3.Config))
	Kt.Panic(err)
	client.ser

	KtCvt.BindInterface(cfg, APro.Cfg["server"])
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9999")
	Kt.Panic(err)
	configure := &thrift.TConfiguration{}
	KtCvt.BindInterface(configure, APro.Cfg["thrift"])
	factory = thrift.NewTCompactProtocolFactoryConf(configure)
	for true {
		conn, err := net.DialTCP("tcp", nil, addr)
		Kt.Err(err, false)
		if conn != nil {
			handle(conn)
		}

		time.Sleep(8 * time.Millisecond)
	}
}
