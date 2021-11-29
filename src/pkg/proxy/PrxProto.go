package proxy

import (
	"axj/Thrd/AZap"
	"bytes"
)

type PrxProto interface {
	// 协议名
	Name() string
	// 读取缓冲区大小
	ReadBufferSize() int
	// 读取缓冲区最大值
	ReadBufferMax() int
	// 读取服务名域名之类的
	ReadServerName(buffer *bytes.Buffer, data []byte, name *string) (bool, error)
}

var protos map[string]PrxProto

func initProtos() {
	protos = map[string]PrxProto{}
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
