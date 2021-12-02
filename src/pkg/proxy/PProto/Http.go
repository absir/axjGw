package PProto

import (
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtUnsafe"
	"bytes"
	lru "github.com/hashicorp/golang-lru"
	"net"
	"strings"
	"sync"
)

type Http struct {
}

type HttpCfg struct {
	BuffSize   int
	ServName   string
	RealIp     string
	CookieAddr string
	locker     sync.Locker
	LruSize    int
	lruCache   *lru.Cache
}

func (h *HttpCfg) GetOrCreateLruCache() *lru.Cache {
	if h.lruCache == nil {
		h.locker.Lock()
		if h.lruCache == nil {
			if h.LruSize < 256 {
				h.LruSize = 256
			}

			h.lruCache, _ = lru.New(h.LruSize)
		}

		h.locker.Unlock()
	}

	return h.lruCache
}

func (h Http) Name() string {
	return "http"
}

func (h Http) NewCfg() interface{} {
	return &HttpCfg{BuffSize: 1024, locker: new(sync.Mutex)}
}

func (h Http) ServAddr(cfg interface{}, sName string) string {
	c := cfg.(*HttpCfg)
	if c.ServName != "" {
		if c.ServName[0] != '.' {
			c.ServName = "." + c.ServName
		}

		return sName + c.ServName
	}

	return ""
}

func (h Http) ReadBufferSize(cfg interface{}) int {
	return cfg.(*HttpCfg).BuffSize
}

func (h Http) ReadBufferMax(cfg interface{}) int {
	return cfg.(*HttpCfg).BuffSize
}

var Host = "Host:"
var HostLen = len(Host)

var Cookie = "Cookie:"
var CookieLen = len(Cookie)

var Content = "Content"
var ContentLen = len(Content)

type HttpCtx struct {
	i        int
	si       int
	hi       int
	realIpSi int
	realIpEi int
	got      bool
	oBuffer  *bytes.Buffer
}

func (that *HttpCtx) reset() {
	that.i = 0
	that.si = 0
	that.hi = 0
	that.realIpSi = 0
	that.realIpEi = 0
	that.got = false
}

func (h Http) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	return new(HttpCtx)
}

func (h Http) ReadServerName(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	buffer.Write(data)
	bs := buffer.Bytes()
	bLen := len(bs)
	c := cfg.(*HttpCfg)
	hCtx := ctx.(*HttpCtx)
	si := hCtx.si
	realIpLen := 0
	if c.RealIp != "" {
		realIpLen = len(c.RealIp)
	}

	done := false
	for i := hCtx.i; i < bLen; i++ {
		b := bs[i]
		hCtx.i = i
		if b == '\r' || b == '\n' {
			if i > si {
				line := KtUnsafe.BytesToString(bs[si:i])
				//println(line)
				lLen := len(line)
				if realIpLen > 0 && realIpLen < lLen && hCtx.realIpEi == 0 && strings.EqualFold(line[:realIpLen], c.RealIp) {
					str := strings.TrimSpace(line[realIpLen:])
					if str != "" && str[0] == ':' {
						hCtx.realIpSi = si
						hCtx.realIpEi = i
					}

				} else if HostLen < lLen && strings.EqualFold(line[:HostLen], Host) {
					name := strings.TrimSpace(line[HostLen+1:])
					done = c.CookieAddr == ""
					if c.ServName != "" && !strings.HasSuffix(name, c.ServName) {
						// CookieAddr ip地址访问，海康摄像头
						if c.CookieAddr != "" && name != "" && name[0] >= '0' && name[0] <= '9' {
							done = false

						} else {
							return true, SERV_NAME_ERR
						}
					}

					if c.RealIp != "" {
						name = string(KtUnsafe.StringToBytes(name))
					}

					*pName = name
					if done {
						return true, nil
					}

					hCtx.hi = si

				} else if c.CookieAddr != "" && CookieLen < lLen && strings.EqualFold(line[:CookieLen], Cookie) {
					if c.CookieAddr != "*" {
						// CookieAddr key值获取
						idx := strings.LastIndex(line, c.CookieAddr)
						if idx > 0 {
							line = line[idx:]
							idx = strings.IndexAny(line, "; ")
							if idx > 0 {
								line = line[:idx]
							}
						}
					}

					name := *pName
					if c.ServName != "" && !strings.HasSuffix(name, c.ServName) {
						// 读取域名映射缓存
						if val, ok := c.GetOrCreateLruCache().Get(line); ok {
							name, _ = val.(string)
							if name != "" {
								*pName = name
							}
						}

					} else {
						// 添加域名映射缓存
						if c.RealIp == "" {
							name = string(KtUnsafe.StringToBytes(name))
						}

						c.GetOrCreateLruCache().Add(string(KtUnsafe.StringToBytes(line)), name)
					}

					done = true
					break

				} else if ContentLen < lLen && strings.EqualFold(line[:ContentLen], Content) {
					done = true
					break
				}
			}

			si = i + 1
			hCtx.si = si
		}
	}

	if done {
		if c.RealIp != "" {
			bs = KtBytes.Copy(bs)
			buffer.Reset()
			si = hCtx.hi
			// 真实ip
			if hCtx.realIpEi == 0 {
				// 添加header
				buffer.Write(bs[:si])
				buffer.Write(KtUnsafe.StringToBytes(c.RealIp))
				buffer.Write(KtUnsafe.StringToBytes(": "))
				buffer.Write(KtUnsafe.StringToBytes(Kt.IpAddr(conn.RemoteAddr())))
				buffer.Write(KtUnsafe.StringToBytes("\n"))
				buffer.Write(bs[si:])

			} else {
				// 修改header
				buffer.Write(bs[:hCtx.realIpSi])
				buffer.Write(KtUnsafe.StringToBytes(c.RealIp))
				buffer.Write(KtUnsafe.StringToBytes(": "))
				buffer.Write(KtUnsafe.StringToBytes(Kt.IpAddr(conn.RemoteAddr())))
				buffer.Write(bs[hCtx.realIpEi:])
			}
		}

		return true, nil
	}

	return false, nil
}

var Get = "GET "
var GetLen = len(Get)

func (h Http) ProcServerCtx(cfg interface{}, ctx interface{}, conn *net.TCPConn) interface{} {
	c := cfg.(*HttpCfg)
	if c.RealIp != "" {
		hCtx := ctx.(*HttpCtx)
		hCtx.reset()
		hCtx.oBuffer = new(bytes.Buffer)
		return hCtx
	}

	return nil
}

func (h Http) ProcServerData(cfg interface{}, ctx interface{}, buffer *bytes.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	buffer.Write(data)
	bs := buffer.Bytes()
	bLen := len(bs)
	c := cfg.(*HttpCfg)
	hCtx := ctx.(*HttpCtx)
	realIpLen := 0
	if c.RealIp != "" {
		realIpLen = len(c.RealIp)
	}

	si := hCtx.i
	for i := hCtx.i; i < bLen; i++ {
		b := bs[i]
		if b == '\r' || b == '\n' {
			if i > si {
				line := KtUnsafe.BytesToString(bs[si:i])
				// println(line)
				lLen := len(line)
				if hCtx.got {
					if realIpLen > 0 && realIpLen < lLen && hCtx.realIpEi == 0 && strings.EqualFold(line[:realIpLen], c.RealIp) {
						str := strings.TrimSpace(line[realIpLen:])
						if str != "" && str[0] == ':' {
							hCtx.realIpSi = si
							hCtx.realIpEi = i
						}

					} else if HostLen < lLen && strings.EqualFold(line[:HostLen], Host) {
						name := strings.TrimSpace(line[HostLen+1:])
						if c.RealIp != "" {
							name = string(KtUnsafe.StringToBytes(name))
							hCtx.oBuffer.Reset()
							hCtx.oBuffer.Write(bs)
							bs = hCtx.oBuffer.Bytes()
							buffer.Reset()
							// 真实ip
							if hCtx.realIpEi == 0 {
								// 添加header
								buffer.Write(bs[:si])
								buffer.Write(KtUnsafe.StringToBytes(c.RealIp))
								buffer.Write(KtUnsafe.StringToBytes(": "))
								buffer.Write(KtUnsafe.StringToBytes(Kt.IpAddr(conn.RemoteAddr())))
								buffer.Write(KtUnsafe.StringToBytes("\n"))
								buffer.Write(bs[si:])

							} else {
								// 修改header
								buffer.Write(bs[:hCtx.realIpSi])
								buffer.Write(KtUnsafe.StringToBytes(c.RealIp))
								buffer.Write(KtUnsafe.StringToBytes(": "))
								buffer.Write(KtUnsafe.StringToBytes(Kt.IpAddr(conn.RemoteAddr())))
								buffer.Write(bs[hCtx.realIpEi:])
							}
						}

						bs = buffer.Bytes()
						buffer.Reset()
						hCtx.reset()
						hCtx.oBuffer.Reset()
						return bs, nil
					}

				} else {
					if GetLen < lLen && strings.EqualFold(line[:GetLen], Get) {
						hCtx.got = true
					}
				}
			}

			si = i + 1
			hCtx.i = si
			if !hCtx.got {
				hCtx.oBuffer.Reset()
				hCtx.oBuffer.Write(bs[:si])
			}
		}
	}

	oBs := hCtx.oBuffer.Bytes()
	hCtx.oBuffer.Reset()
	oBLen := len(oBs)
	if oBLen > 0 {
		buffer.Reset()
		buffer.Write(bs[oBLen:])
		if hCtx.i > 0 {
			hCtx.i -= oBLen
		}

		if hCtx.si > 0 {
			hCtx.si -= oBLen
		}

		if hCtx.realIpEi > 0 {
			hCtx.realIpEi -= oBLen
			hCtx.realIpSi -= oBLen
		}
	}

	return oBs, nil
}
