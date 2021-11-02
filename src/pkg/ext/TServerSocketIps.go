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

func (that *TServerSocketIps) Accept() (thrift.TTransport, error) {
	t, err := that.TServerSocket.Accept()
	if t != nil {
		socket := t.(*thrift.TSocket)
		ip := socket.Conn().RemoteAddr().String()
		i := strings.Index(ip, ":")
		if i >= 0 {
			ip = ip[0:i]
		}

		if that.Ips == nil || !that.Ips(ip) {
			socket.Close()
			return nil, thrift.NewTTransportException(thrift.NOT_OPEN, "No Allow ip "+ip)
		}
	}

	return t, err
}
