package anet

import (
	"bufio"
	"sync"
)

type Client interface {
	// 读取
	Read() (error, []byte, *bufio.Reader)
	// 粘包
	Sticky() bool
	// 流写入
	Output() (err error, out bool, locker sync.Locker)
	// 写入
	Write(bs []byte, out bool) (err error)
}
