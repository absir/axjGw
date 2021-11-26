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
	// 写入异步
	IsWriteAsync() bool
	// 关闭
	Close()
	// 远程地址
	RemoteAddr() string
}

type ConnSocket struct {
	conn  *net.TCPConn
	out   bool
	wBuff []byte
	rBuff []byte
}

func NewConnSocket(conn *net.TCPConn, out bool) *ConnSocket {
	if conn == nil {
		return nil
	}

	that := new(ConnSocket)
	that.conn = conn
	that.out = out
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
	if that.out {
		return &that.wBuff
	}

	return nil
}

func (that *ConnSocket) Write(bs []byte) error {
	_, err := that.Conn().Write(bs)
	return err
}

func (that *ConnSocket) IsWriteAsync() bool {
	return false
}

func (that *ConnSocket) Close() {
	conn := that.Conn()
	conn.SetLinger(0)
	conn.Close()
}

func (that *ConnSocket) RemoteAddr() string {
	return that.Conn().RemoteAddr().String()
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

func (that *ConnWebsocket) IsWriteAsync() bool {
	return false
}

func (that *ConnWebsocket) Close() {
	that.Conn().Close()
}

func (that *ConnWebsocket) RemoteAddr() string {
	return that.Conn().RemoteAddr().String()
}
