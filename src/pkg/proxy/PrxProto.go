package proxy

import "bytes"

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
