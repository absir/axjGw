package ANet

import (
	"axj/Kt/Kt"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"time"
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
	// 写入
	Write(bs []byte) error
	// 写入异步
	IsWriteAsync() bool
	// 状态延迟
	SetLinger(sec int) error
	// 关闭
	Close(immed bool)
	// 远程地址
	RemoteAddr() string
	// EPoll模式
	ConnPoll() *ConnPoll
}

type ConnSocket struct {
	conn  *net.TCPConn
	rBuff []byte
}

func CloseDelayTcp(conn *net.TCPConn, drt time.Duration) {
	if drt < 1 {
		drt = 1
	}

	time.Sleep(drt)
	conn.SetLinger(0)
	conn.Close()
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

func (that *ConnSocket) Write(bs []byte) error {
	_, err := that.Conn().Write(bs)
	return err
}

func (that *ConnSocket) IsWriteAsync() bool {
	return false
}

func (that *ConnSocket) SetLinger(sec int) error {
	return that.Conn().SetLinger(sec)
}

func (that *ConnSocket) Close(immed bool) {
	conn := that.Conn()
	if immed {
		conn.SetLinger(0)
	}

	conn.Close()
}

func (that *ConnSocket) RemoteAddr() string {
	return Kt.IpAddr(that.Conn().RemoteAddr())
}

func (that *ConnSocket) ConnPoll() *ConnPoll {
	return nil
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

func (that *ConnWebsocket) Write(bs []byte) error {
	return websocket.Message.Send(that.Conn(), bs)
}

func (that *ConnWebsocket) IsWriteAsync() bool {
	return false
}

func (that *ConnWebsocket) SetLinger(sec int) error {
	return nil
}

func (that *ConnWebsocket) Close(immed bool) {
	that.Conn().Close()
}

func (that *ConnWebsocket) RemoteAddr() string {
	request := that.Conn().Request()
	return Kt.IpAddrStr(request.RemoteAddr)
}

func (that *ConnWebsocket) ConnPoll() *ConnPoll {
	return nil
}
