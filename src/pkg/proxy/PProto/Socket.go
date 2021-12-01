package PProto

import (
	"bytes"
	"net"
)

type Socket struct {
}

type SocketCfg struct {
	BuffSize int
}

func (h Socket) Name() string {
	return "socket"
}

func (h Socket) NewCfg() interface{} {
	return &SocketCfg{
		BuffSize: 4096,
	}
}

func (h Socket) ServAddr(cfg interface{}, sName string) string {
	return ""
}

func (h Socket) ReadBufferSize(cfg interface{}) int {
	return cfg.(*SocketCfg).BuffSize
}

func (h Socket) ReadBufferMax(cfg interface{}) int {
	return cfg.(*SocketCfg).BuffSize
}

func (h Socket) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	return nil
}

func (h Socket) ReadServerName(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	buffer.Write(data)
	return true, nil
}

func (h Socket) ProcServerCtx(cfg interface{}, ctx interface{}, conn *net.TCPConn) interface{} {
	return nil
}

func (h Socket) ProcServerData(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	return data, nil
}
