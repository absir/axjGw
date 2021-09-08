package thrift

import (
	"github.com/apache/thrift/lib/go/thrift"
	"strings"
)

type TServerSocketIps struct {
	thrift.TServerSocket
}

func (p *TServerSocketIps) Allow(ip string) bool {
	return true
}

func (p *TServerSocketIps) Accept() (thrift.TTransport, error) {
	t, err := p.TServerSocket.Accept()
	if t != nil {
		socket := t.(*thrift.TSocket)
		ip := socket.Conn().RemoteAddr().String()
		i := strings.Index(ip, ":")
		if i >= 0 {
			ip = ip[0:i]
		}

		if !p.Allow(ip) {
			socket.Close()
			return nil, thrift.NewTTransportException(thrift.NOT_OPEN, "No Allow ip "+ip)
		}
	}

	return t, err
}
