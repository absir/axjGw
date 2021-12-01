package PProto

import (
	"axj/Kt/KtUnsafe"
	"bytes"
	"net"
)

type Rtmp struct {
}

type RtmpCfg struct {
	ServName string
}

func (r Rtmp) Name() string {
	return "rtmp"
}

func (r Rtmp) NewCfg() interface{} {
	return new(RtmpCfg)
}

func (r Rtmp) ServAddr(cfg interface{}, sName string) string {
	c := cfg.(*RtmpCfg)
	if c.ServName != "" {
		if c.ServName[0] != '.' {
			c.ServName = "." + c.ServName
		}

		return sName + c.ServName
	}

	return ""
}

func (r Rtmp) ReadBufferSize(cfg interface{}) int {
	return 2048
}

func (r Rtmp) ReadBufferMax(cfg interface{}) int {
	return 256
}

func (r Rtmp) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	return nil
}

func (r Rtmp) ReadServerName(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	buffer.Write(data)
	idx := bytes.Index(buffer.Bytes(), []byte("f49634947547.proxy.com:9083"))
	if idx > 0 {
		println(buffer.Bytes()[idx:])
	}

	println(KtUnsafe.BytesToString(buffer.Bytes()))
	return false, nil
}

func (r Rtmp) ProcServerCtx(cfg interface{}, ctx interface{}, conn *net.TCPConn) interface{} {
	return nil
}

func (r Rtmp) ProcServerData(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	return data, nil
}
