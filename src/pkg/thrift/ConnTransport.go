package thrift

import (
	"context"
	"net"
)

type ConnTransport net.TCPConn

func (c ConnTransport) Flush(ctx context.Context) (err error) {
	return nil
}

func (c ConnTransport) RemainingBytes() (num_bytes uint64) {
	const maxSize = ^uint64(0)
	return maxSize // the truth is, we just don't know unless framed is used
}

func (c ConnTransport) Open() error {
	return nil
}

func (c ConnTransport) IsOpen() bool {
	return true
}
