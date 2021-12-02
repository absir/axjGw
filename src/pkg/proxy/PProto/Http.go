package PProto

import (
	"axj/Kt/Kt"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
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

var ContentLength = "Content-Length: "
var ContentLengthLen = len(ContentLength)

var Get = "GET "
var GetLen = len(Get)
var Post = "POST "
var PostLen = len(Post)
var Put = "PUT "
var PutLen = len(Put)
var Delete = "DELETE "
var DeleteLen = len(Delete)

type HttpCtx struct {
	name       string
	i          int
	si         int
	hi         int
	rn         int
	got        bool
	contentLen int
	realIpSi   int
	realIpEi   int
	oBuffer    *bytes.Buffer
}

func (that *HttpCtx) reset() {
	that.i = 0
	that.si = 0
	that.hi = 0
	that.rn = 0
	that.got = false
	that.realIpSi = 0
	that.realIpEi = 0
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
		hCtx.i = i + 1
		if b == '\r' || b == '\n' {
			if i > si {
				line := KtUnsafe.BytesToString(bs[si:i])
				// println(line)
				lLen := len(line)
				if realIpLen > 0 && realIpLen < lLen && hCtx.realIpEi == 0 && strings.EqualFold(line[:realIpLen], c.RealIp) {
					str := strings.TrimSpace(line[realIpLen:])
					if str != "" && str[0] == ':' {
						hCtx.realIpSi = si
						hCtx.realIpEi = i
					}

				} else if HostLen < lLen && strings.EqualFold(line[:HostLen], Host) {
					hCtx.hi = si
					name := strings.TrimSpace(line[HostLen+1:])
					done = c.CookieAddr == ""
					cName := false
					if c.ServName != "" && !strings.HasSuffix(name, c.ServName) {
						// CookieAddr ip地址访问，海康摄像头
						if c.CookieAddr != "" && name != "" && name[0] >= '0' && name[0] <= '9' {
							done = false

						} else {
							return true, SERV_NAME_ERR
						}

					} else if c.CookieAddr != "" {
						cName = true
					}

					if c.RealIp != "" || cName {
						name = string(KtUnsafe.StringToBytes(name))
					}

					*pName = name
					if cName {
						hCtx.name = name
					}

					if done {
						break
					}

				} else if c.CookieAddr != "" && CookieLen < lLen && strings.EqualFold(line[:CookieLen], Cookie) {
					if c.CookieAddr != "*" {
						// CookieAddr key值获取
						idx := strings.LastIndex(line, c.CookieAddr)
						if idx >= 0 {
							line = line[idx:]
							idx = strings.IndexAny(line, "; ")
							if idx > 0 {
								line = line[:idx]
							}

						} else {
							// 没有CookieAddr
							done = true
							break
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

			} else if hCtx.rn > 2 || (hCtx.rn > 1 && i > 1 && bs[i-1] == b) {
				done = true
				break
			}

			si = hCtx.i
			hCtx.si = si
			hCtx.rn++

		} else {
			hCtx.rn = 0
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

func (h Http) ProcServerCtx(cfg interface{}, ctx interface{}, conn *net.TCPConn) interface{} {
	c := cfg.(*HttpCfg)
	hCtx := ctx.(*HttpCtx)
	if c.RealIp != "" || hCtx.name != "" {
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

	println(KtUnsafe.BytesToString(bs))
	si := hCtx.i
	for i := hCtx.i; i < bLen; i++ {
		b := bs[i]
		hCtx.i = i + 1
		if b == '\r' || b == '\n' {
			if i > si {
				line := KtUnsafe.BytesToString(bs[si:i])
				// println(line)
				lLen := len(line)
				if hCtx.got {
					if realIpLen > 0 {
						if realIpLen < lLen && hCtx.realIpEi == 0 && strings.EqualFold(line[:realIpLen], c.RealIp) {
							str := strings.TrimSpace(line[realIpLen:])
							if str != "" && str[0] == ':' {
								hCtx.realIpSi = si
								hCtx.realIpEi = i
							}

						} else if HostLen < lLen && strings.EqualFold(line[:HostLen], Host) {
							hCtx.hi = si
						}
					}

				} else {
					// got状态判断
					if GetLen < lLen && strings.EqualFold(line[:GetLen], Get) {
						hCtx.got = true

					} else if PostLen < lLen && strings.EqualFold(line[:PostLen], Post) {
						hCtx.got = true

					} else if PutLen < lLen && strings.EqualFold(line[:GetLen], Put) {
						hCtx.got = true

					} else if DeleteLen < lLen && strings.EqualFold(line[:GetLen], Delete) {
						hCtx.got = true
					}
				}

				if ContentLengthLen < lLen && strings.EqualFold(line[:ContentLengthLen], ContentLength) {
					// ContentLength大小
					hCtx.contentLen = int(KtCvt.ToInt32(strings.TrimSpace(line[ContentLengthLen:])))

				} else if hCtx.name != "" && CookieLen < lLen && strings.EqualFold(line[:CookieLen], Cookie) {
					// CookieAddr 映射更新
					if c.CookieAddr != "*" {
						// CookieAddr key值获取
						idx := strings.LastIndex(line, c.CookieAddr)
						if idx >= 0 {
							line = line[idx:]
							idx = strings.IndexAny(line, "; ")
							if idx > 0 {
								line = line[:idx]
							}

						} else {
							line = ""
						}
					}

					if line != "" {
						c.GetOrCreateLruCache().Add(string(KtUnsafe.StringToBytes(line)), hCtx.name)
					}
				}

			} else if hCtx.got && hCtx.rn > 2 || (hCtx.rn > 1 && i > 1 && bs[i-1] == b) {
				// 请求数据完成
				hCtx.oBuffer.Write(bs)
				bs = hCtx.oBuffer.Bytes()
				buffer.Reset()
				// 真实ip
				if hCtx.realIpEi == 0 {
					// 添加header
					buffer.Write(bs[:hCtx.hi])
					buffer.Write(KtUnsafe.StringToBytes(c.RealIp))
					buffer.Write(KtUnsafe.StringToBytes(": "))
					buffer.Write(KtUnsafe.StringToBytes(Kt.IpAddr(conn.RemoteAddr())))
					buffer.Write(KtUnsafe.StringToBytes("\n"))
					buffer.Write(bs[hCtx.hi:])

				} else {
					// 修改header
					buffer.Write(bs[:hCtx.realIpSi])
					buffer.Write(KtUnsafe.StringToBytes(c.RealIp))
					buffer.Write(KtUnsafe.StringToBytes(": "))
					buffer.Write(KtUnsafe.StringToBytes(Kt.IpAddr(conn.RemoteAddr())))
					buffer.Write(bs[hCtx.realIpEi:])
				}

				bs = buffer.Bytes()
				buffer.Reset()
				hCtx.reset()
				hCtx.oBuffer.Reset()
				println(KtUnsafe.BytesToString(bs))
				return bs, nil
			}

			si = hCtx.i
			hCtx.si = si
			hCtx.rn++

		} else {
			hCtx.rn = 0
		}
	}

	if realIpLen <= 0 || !hCtx.got {
		hCtx.oBuffer.Reset()
		// 最后换行数据 包含Content-Length 固定大小 数据
		for i := bLen - 1; i >= hCtx.contentLen; i-- {
			b := bs[i]
			if b == '\r' || b == '\n' {
				hCtx.oBuffer.Write(bs[:i+1])
				hCtx.contentLen = 0
				break
			}
		}

		// Content-Length 固定大小
		if hCtx.contentLen > 0 && hCtx.contentLen <= bLen {
			hCtx.oBuffer.Write(bs[:hCtx.contentLen])
			hCtx.contentLen = 0
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
			if hCtx.si < 0 {
				hCtx.si = 0
			}
		}

		if hCtx.realIpEi > 0 {
			hCtx.realIpEi -= oBLen
			hCtx.realIpSi -= oBLen
		}
	}

	return oBs, nil
}
