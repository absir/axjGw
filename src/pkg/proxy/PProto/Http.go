package PProto

import (
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtCvt"
	"axj/Kt/KtUnsafe"
	lru "github.com/hashicorp/golang-lru"
	"net"
	"strings"
)

type Http struct {
}

type HttpCfg struct {
	BuffSize      int
	ServName      string
	RealIp        string
	CookieAddr    string
	lenRealIp     int
	lenCookieAddr int
	LruSize       int
	lruCache      *lru.Cache
}

func (h Http) Name() string {
	return "http"
}

func (h Http) NewCfg() interface{} {
	return &HttpCfg{BuffSize: 1024}
}

func (t Http) InitCfg(cfg interface{}) {
	c := cfg.(*HttpCfg)
	c.lenRealIp = len(c.RealIp)
	if c.ServName != "" {
		c.lenCookieAddr = len(c.CookieAddr)
		if c.lenCookieAddr > 0 {
			if c.lruCache == nil {
				if c.LruSize < 256 {
					c.LruSize = 256
				}

				c.lruCache, _ = lru.New(c.LruSize)
			}
		}
	}
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

var ContentLength = "Content-Length:"
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
	HostCtx
	// CRLF \r\n
	rn         byte
	rn2        bool
	rnsLen     int // len to rns = return data len
	readLen    int // ReadServerName bs len
	name       string
	contentLen int
	connAddr   string // 客户端地址
}

func (h Http) ReadServerCtx(cfg interface{}, conn *net.TCPConn) interface{} {
	c := cfg.(*HttpCfg)
	if c.lenRealIp <= 0 && c.lenCookieAddr <= 0 {
		return &HostCtx{}
	}

	return &HttpCtx{}
}

func (h Http) ReadServerName(cfg interface{}, ctx interface{}, buffer *KtBuffer.Buffer, data []byte, pName *string, conn *net.TCPConn) (bool, error) {
	c := cfg.(*HttpCfg)
	if c.lenRealIp <= 0 && c.lenCookieAddr <= 0 {
		return hostReadServerName(ctx, buffer, data, pName, Host, HostLen, c.ServName, nil)
	}

	_, err := h.ProcServerData(cfg, ctx, buffer, data, conn)
	if err != nil {
		return true, err
	}

	hCtx := ctx.(*HttpCtx)
	if hCtx.name != "" {
		if !allowName(hCtx.name, c.ServName) {
			return true, SERV_NAME_ERR
		}

		*pName = hCtx.name
		hCtx.readLen = buffer.Len()
		KtBuffer.SetLen(buffer, hCtx.rnsLen)
		return true, nil
	}

	// 不返回数据
	hCtx.rnsLen = 0
	return false, nil
}

func (h Http) ProcServerCtx(cfg interface{}, ctx interface{}, buffer *KtBuffer.Buffer, conn *net.TCPConn) interface{} {
	c := cfg.(*HttpCfg)
	if c.lenRealIp <= 0 && c.lenCookieAddr <= 0 {
		return nil
	}

	hCtx := ctx.(*HttpCtx)
	if hCtx.readLen > hCtx.rnsLen {
		KtBuffer.SetLen(buffer, hCtx.readLen)
		KtBuffer.SetRangeLen(buffer, 0, hCtx.rnsLen, 0)
	}

	hCtx.rnsLen = 0
	hCtx.i = 0
	hCtx.si = 0
	return ctx
}

func (h Http) ProcServerData(cfg interface{}, ctx interface{}, buffer *KtBuffer.Buffer, data []byte, conn *net.TCPConn) ([]byte, error) {
	hCtx := ctx.(*HttpCtx)
	if hCtx.rnsLen > 0 {
		// dLen 数据处理
		KtBuffer.SetRangeLen(buffer, 0, hCtx.rnsLen, 0)
		hCtx.i -= hCtx.rnsLen
		hCtx.si -= hCtx.rnsLen
		hCtx.rnsLen = 0
	}

	c := cfg.(*HttpCfg)
	buffer.Write(data)
	bs := buffer.Bytes()
	bLen := len(bs)
	si := hCtx.i
	for i := hCtx.i; i < bLen; i = hCtx.i {
		hCtx.i = i + 1
		if hCtx.rn2 {
			// contentLen数据读取开始
			if hCtx.contentLen > 0 {
				hCtx.contentLen--
				si = hCtx.i
				hCtx.si = si
				hCtx.rnsLen = si
				continue
			}

			hCtx.rn2 = false
		}

		b := bs[i]
		if b == '\r' && (hCtx.rn == 0 || hCtx.rn == 23) {
			hCtx.rn += b

		} else if b == '\n' {
			if hCtx.rn == 13 {
				// 可连续
				hCtx.rn += b
				// CRLF
				ei := i - 1
				if ei > si {
					line := KtUnsafe.BytesToString(bs[si:ei])
					// println("line = " + line)
					lLen := len(line)
					if lLen > ContentLengthLen && strings.EqualFold(line[:ContentLengthLen], ContentLength) {
						// ContentLength大小
						hCtx.contentLen = int(KtCvt.ToInt32(strings.TrimSpace(line[ContentLengthLen:])))

					} else if hCtx.name == "" && lLen > HostLen && strings.EqualFold(line[:HostLen], Host) {
						// Host获取
						hCtx.name = string(KtUnsafe.StringToBytes(strings.TrimSpace(line[HostLen:])))
						// RealIp 添加
						if c.lenRealIp > 0 && si > 0 {
							if hCtx.connAddr == "" {
								hCtx.connAddr = Kt.IpAddr(conn.RemoteAddr())
							}

							//RealIp: connAddr\r\n
							rLen := c.lenRealIp + 2 + len(hCtx.connAddr) + 2
							KtBuffer.SetRangeLen(buffer, si, si, rLen)
							bs = buffer.Bytes()
							{
								// 写入RealIp
								off := si
								copy(bs[off:], KtUnsafe.StringToBytes(c.RealIp))
								off += c.lenRealIp
								bs[off] = ':'
								off += 1
								bs[off] = ' '
								off += 1
								copy(bs[off:], KtUnsafe.StringToBytes(hCtx.connAddr))
								off += len(hCtx.connAddr)
								bs[off] = '\r'
								off += 1
								bs[off] = '\n'
							}

							//println(KtUnsafe.BytesToString(bs))
							bLen = len(bs)
							hCtx.i += rLen
							i = hCtx.i - 1
						}

					} else if c.lenRealIp > 0 && si > 0 && lLen > c.lenRealIp && line[c.lenRealIp] == ':' && strings.EqualFold(line[:c.lenRealIp], c.RealIp) {
						// lenRealIp 移除
						KtBuffer.SetRangeLen(buffer, si, hCtx.i, 0)
						bs = buffer.Bytes()
						bLen = len(bs)
						hCtx.i = si
						i = hCtx.i - 1

					} else if c.lenCookieAddr > 0 && hCtx.name != "" && lLen > c.lenCookieAddr && strings.EqualFold(line[:CookieLen], Cookie) {
						// CookieAddr 设置读取
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
							if allowName(hCtx.name, c.ServName) {
								// 设置name缓存
								c.lruCache.Add(string(KtUnsafe.StringToBytes(line)), hCtx.name)

							} else {
								// 读取name缓存
								if val, ok := c.lruCache.Get(line); ok {
									hCtx.name, _ = val.(string)
								}
							}
						}
					}
				}

			} else if hCtx.rn == 36 {
				// 可连续
				hCtx.rn = 23
				// CRLF|CRLF
				hCtx.rn2 = true

			} else {
				hCtx.rn = 0
			}

			if hCtx.rn == 23 {
				hCtx.rnsLen = hCtx.i
			}

			si = hCtx.i
			hCtx.si = si

		} else {
			hCtx.rn = 0
		}
	}

	if hCtx.rnsLen <= 0 {
		return nil, nil
	}

	if hCtx.rnsLen < bLen {
		return bs[:hCtx.rnsLen], nil
	}

	return bs, nil
}
