package proxy

import (
	"axj/Thrd/AZap"
	"axjGW/pkg/proxy/PProto"
	"bytes"
	"net"
)

type PrxProto interface {
	// 协议名
	Name() string
	// 协议配置
	NewCfg() interface{}
	// 服务地址
	ServAddr(cfg interface{}, sName string) string
	// 读取缓冲区大小
	ReadBufferSize(cfg interface{}) int
	// 读取缓冲区最大值
	ReadBufferMax(cfg interface{}) int
	// 读取服务名域名版主对象
	ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{}
	// 读取服务名域名之类的
	ReadServerName(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error)
	// 数据加工
	ProcServerCtx(cfg interface{}, ctx interface{}, conn *net.TCPConn) interface{}
	// 数据加工
	ProcServerData(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, conn *net.TCPConn) ([]byte, error)
}

var protos map[string]PrxProto

func initProtos() {
	protos = map[string]PrxProto{}
	RegProto(&PProto.Tcp{})
	RegProto(&PProto.Http{})
	RegProto(&PProto.Rtsp{})
}

func RegProto(proto PrxProto) {
	protos[proto.Name()] = proto
}

func FindProto(name string, warn bool) PrxProto {
	proto := protos[name]
	if proto == nil && warn {
		AZap.Logger.Warn("FindProto nil At " + name)
	}

	return proto
}
