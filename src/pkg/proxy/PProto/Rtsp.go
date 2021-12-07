package PProto

import (
	"axj/Kt/KtBuffer"
	"net"
	"strings"
)

type Rtsp struct {
}

type RtmpCfg struct {
	BuffSize int
	ServName string
}

func (r Rtsp) Name() string {
	return "rtsp"
}

func (r Rtsp) NewCfg() interface{} {
	return &RtmpCfg{
		BuffSize: 10240,
	}
}

func (t Rtsp) InitCfg(cfg interface{}) {
}

func (r Rtsp) ServAddr(cfg interface{}, sName string) string {
	c := cfg.(*RtmpCfg)
	if c.ServName != "" {
		if c.ServName[0] != '.' {
			c.ServName = "." + c.ServName
		}

		return sName + c.ServName
	}

	return ""
}

func (r Rtsp) ReadBufferSize(cfg interface{}) int {
	return cfg.(*RtmpCfg).BuffSize
}

func (r Rtsp) ReadBufferMax(cfg interface{}) int {
	return cfg.(*RtmpCfg).BuffSize
}

func (r Rtsp) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	return &HostCtx{}
}

var OPTIONS = "OPTIONS "
var OPTIONS_LEN = len(OPTIONS)

func urlHostName(name string) string {
	idx := strings.Index(name, "//")
	if idx >= 0 {
		name = name[idx+2:]
	}

	idx = strings.IndexAny(name, " /")
	if idx >= 0 {
		name = name[:idx]
	}

	return name
}

func (r Rtsp) ReadServerName(cfg interface{}, ctx interface{}, buffer *KtBuffer.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	c := cfg.(*RtmpCfg)
	return hostReadServerName(ctx, buffer, data, pName, OPTIONS, OPTIONS_LEN, c.ServName, urlHostName)
}

func (r Rtsp) ProcServerCtx(cfg interface{}, ctx interface{}, buffer *KtBuffer.Buffer, conn *net.TCPConn) interface{} {
	return nil
}

func (r Rtsp) ProcServerData(cfg interface{}, ctx interface{}, buffer *KtBuffer.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	return data, nil
}
