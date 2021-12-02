package PProto

import (
	"axj/Kt/KtUnsafe"
	"bytes"
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

type RtspCtx struct {
	i  int
	si int
}

var OPTIONS = "OPTIONS "
var OPTIONS_LEN = len(OPTIONS)

func (r Rtsp) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	return &RtspCtx{}
}

func (r Rtsp) ReadServerName(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	buffer.Write(data)
	bs := buffer.Bytes()
	bLen := len(bs)
	hCtx := ctx.(*RtspCtx)
	si := hCtx.si
	for i := hCtx.i; i < bLen; i++ {
		b := bs[i]
		hCtx.i = i
		if b == '\r' || b == '\n' {
			if i > si {
				line := KtUnsafe.BytesToString(bs[si:i])
				// println(line)
				lLen := len(line)
				if lLen > OPTIONS_LEN && strings.EqualFold(line[:OPTIONS_LEN], OPTIONS) {
					name := strings.TrimSpace(line[OPTIONS_LEN:])
					idx := strings.Index(name, "//")
					if idx >= 0 {
						name = name[idx+2:]
					}

					idx = strings.IndexAny(name, ":/")
					if idx >= 0 {
						name = name[:idx]
					}

					*pName = name
					return true, nil
				}
			}

			si = i + 1
			hCtx.si = si
		}
	}

	return false, nil
}

func (r Rtsp) ProcServerCtx(cfg interface{}, ctx interface{}, conn *net.TCPConn) interface{} {
	return nil
}

func (r Rtsp) ProcServerData(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	return data, nil
}
