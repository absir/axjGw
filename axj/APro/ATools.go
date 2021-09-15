package APro

import (
	"axj/KtStr"
)

// 获取IP地址
func Ip(addr string) string {
	i := KtStr.IndexByte(addr, ':', 0)
	if i >= 0 {
		addr = addr[0:i]
	}

	return addr
}
