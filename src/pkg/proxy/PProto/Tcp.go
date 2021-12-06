package PProto

import (
	"bytes"
	"net"
)

type Tcp struct {
}

type TcpCfg struct {
	BuffSize int
}

func (t Tcp) Name() string {
	return "tcp"
}

func (t Tcp) NewCfg() interface{} {
	return &TcpCfg{
		BuffSize: 4096,
	}
}

func (t Tcp) InitCfg(cfg interface{}) {
}

func (t Tcp) ServAddr(cfg interface{}, sName string) string {
	return ""
}

func (t Tcp) ReadBufferSize(cfg interface{}) int {
	return cfg.(*TcpCfg).BuffSize
}

func (t Tcp) ReadBufferMax(cfg interface{}) int {
	return cfg.(*TcpCfg).BuffSize
}

func (t Tcp) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	return nil
}

func (t Tcp) ReadServerName(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	buffer.Write(data)
	return true, nil
}

func (t Tcp) ProcServerCtx(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, conn *net.TCPConn) interface{} {
	return nil
}

func (t Tcp) ProcServerData(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	return data, nil
}
