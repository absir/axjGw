// +build !wsN

package asdk

import (
	"axj/ANet"
	"golang.org/x/net/websocket"
)

func wsDial(addr string) (ANet.Conn, error) {
	conn, err := websocket.Dial(addr, "", addr)
	if conn == nil || err != nil {
		return nil, err
	}

	return ANet.NewConnWebsocket(conn), err
}
