package proxy

import (
	"bytes"
	"net"
	"sync"
)

type PrxAdap struct {
	proto  *PrxProto
	conn   *net.TCPConn
	buff   []byte
	buffer bytes.Buffer
	locker sync.Locker
	cond   *sync.Cond
}
