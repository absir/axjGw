package ext

import (
	"context"
	"github.com/apache/thrift/lib/go/thrift"
)

type TSocketAlive struct {
	*thrift.TSocket
}

func NewTSocketAlive(socket *thrift.TSocket) *TSocketAlive {
	that := &TSocketAlive{
		TSocket: socket,
	}

	return that
}

func (that TSocketAlive) Read(p []byte) (n int, err error) {
	n, err = that.TSocket.Read(p)
	if err != nil {
		that.Close()
	}

	return n, err
}

func (that TSocketAlive) Write(p []byte) (n int, err error) {
	n, err = that.TSocket.Write(p)
	if err != nil {
		that.Close()
	}

	return n, err
}

func (that TSocketAlive) Flush(ctx context.Context) (err error) {
	err = that.TSocket.Flush(ctx)
	if err != nil {
		that.Close()
	}

	return err
}
