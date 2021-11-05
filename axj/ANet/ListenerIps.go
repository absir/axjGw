package ANet

import (
	"errors"
	"net"
)

var ERR_DENIED = errors.New("DENIED")

type ListenerIps struct {
	lis net.Listener
	ips func(ip string) bool
}

func NewListenerIps(lis net.Listener, ips func(ip string) bool) *ListenerIps {
	that := new(ListenerIps)
	that.lis = lis
	that.ips = ips
	return that
}

func (that ListenerIps) Accept() (net.Conn, error) {
	conn, err := that.lis.Accept()
	if conn != nil {
		if that.ips != nil && !that.ips(conn.RemoteAddr().String()) {
			conn.Close()
			return nil, ERR_DENIED
		}
	}

	return conn, err
}

func (that ListenerIps) Close() error {
	return that.lis.Close()
}

func (that ListenerIps) Addr() net.Addr {
	return that.lis.Addr()
}
