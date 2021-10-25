package ext

import (
	"github.com/apache/thrift/lib/go/thrift"
	"strings"
)

type TServerSocketIps struct {
	thrift.TServerSocket
	Ips func(ip string) bool
}

func NewTServerSocketIps(socket *thrift.TServerSocket, ips func(ip string) bool) *TServerSocketIps {
	socketIps := new(TServerSocketIps)
	socketIps.TServerSocket = *socket
	socketIps.Ips = ips
	return socketIps
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

		if p.Ips == nil || !p.Ips(ip) {
			socket.Close()
			return nil, thrift.NewTTransportException(thrift.NOT_OPEN, "No Allow ip "+ip)
		}
	}

	return t, err
}
