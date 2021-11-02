package ANet

import (
	"golang.org/x/net/websocket"
	"io"
	"net"
)

type Reader interface {
	io.Reader
	io.ByteReader
}

type Conn interface {
	// 读取
	ReadA() (error, []byte, Reader)
	// 粘包
	Sticky() bool
	// 流写入
	Out() *[]byte
	// 写入
	Write(bs []byte) error
	// 关闭
	Close()
}

type ConnSocket struct {
	conn  *net.TCPConn
	rBuff []byte
	wBuff []byte
}

func NewConnSocket(conn *net.TCPConn) *ConnSocket {
	if conn == nil {
		return nil
	}

	that := new(ConnSocket)
	that.conn = conn
	return that
}

func (that *ConnSocket) Conn() *net.TCPConn {
	return that.conn
}

func (that *ConnSocket) Read(b []byte) (int, error) {
	return that.conn.Read(b)
}

func (that *ConnSocket) ReadByte() (byte, error) {
	buff := that.rBuff
	if buff == nil {
		buff = make([]byte, 1)
		that.rBuff = buff
	}

	i, err := that.Conn().Read(buff)
	if i > 0 {
		return buff[0], err
	}

	return 0, err
}

func (that *ConnSocket) ReadA() (error, []byte, Reader) {
	return nil, nil, that
}

func (that *ConnSocket) Sticky() bool {
	return true
}

func (that *ConnSocket) Out() *[]byte {
	return &that.wBuff
}

func (that *ConnSocket) Write(bs []byte) error {
	_, err := that.Conn().Write(bs)
	return err
}

func (that *ConnSocket) Close() {
	that.Conn().Close()
}

type ConnWebsocket websocket.Conn

func NewConnWebsocket(conn *websocket.Conn) *ConnWebsocket {
	if conn == nil {
		return nil
	}

	that := ConnWebsocket(*conn)
	return &that
}

func (that *ConnWebsocket) Conn() *websocket.Conn {
	conn := websocket.Conn(*that)
	return &conn
}

func (that *ConnWebsocket) ReadA() (error, []byte, Reader) {
	var bs []byte
	err := websocket.Message.Receive(that.Conn(), &bs)
	if err != nil {
		return err, nil, nil
	}

	return nil, bs, nil
}

func (that *ConnWebsocket) Sticky() bool {
	return false
}

func (that *ConnWebsocket) Out() *[]byte {
	return nil
}

func (that *ConnWebsocket) Write(bs []byte) error {
	return websocket.Message.Send(that.Conn(), bs)
}

func (that *ConnWebsocket) Close() {
	that.Conn().Close()
}
