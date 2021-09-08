package main

import (
	"axj/APro"
	"axj/Kt"
	"axj/KtCvt"
	"context"
	"github.com/apache/thrift/lib/go/thrift"
	"net"
	"time"
)

type Config struct {
	addr string
}

var cfg = Config{}

var factory *thrift.TCompactProtocolFactory

func main() {
	bs := []int{0, 1, 3}
	println(bs[0])
	as := bs[0:1]
	as[0] = 1
	println(bs[0])

	cfg.addr = "127.0.0.1:8082"
	APro.Load(nil, "config.yaml")
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

func handle(conn *net.TCPConn) {
	defer conn.Close()
	trans := &ConnTrans{conn: conn}
	proto := factory.GetProtocol(trans)
	ctx := context.Background()
	for true {
		proto.ReadI32(ctx)
	}
}

type ConnTrans struct {
	conn *net.TCPConn
}

func (c ConnTrans) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}

func (c ConnTrans) Write(p []byte) (n int, err error) {
	return c.conn.Write(p)
}

func (c ConnTrans) Close() error {
	return c.conn.Close()
}

func (c ConnTrans) Flush(ctx context.Context) (err error) {
	return nil
}

func (c ConnTrans) RemainingBytes() (num_bytes uint64) {
	const maxSize = ^uint64(0)
	return maxSize // the truth is, we just don't know unless framed is used
}

func (c ConnTrans) Open() error {
	return nil
}

func (c ConnTrans) IsOpen() bool {
	return c.conn != nil
}
