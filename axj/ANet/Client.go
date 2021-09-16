package ANet

import (
	"bufio"
	"golang.org/x/net/websocket"
	"net"
	"sync"
)

type Client interface {
	// 读取
	Read() (error, []byte, *bufio.Reader)
	// 粘包
	Sticky() bool
	// 流写入
	Output() (err error, out bool, locker sync.Locker)
	// 写入
	Write(bs []byte, out bool) (err error)
	// 关闭
	Close()
}

type ClientSocket struct {
	conn   *net.TCPConn
	reader *bufio.Reader
	locker sync.Locker
}

func NewClientSocket(conn *net.TCPConn, size int, out bool) *ClientSocket {
	client := new(ClientSocket)
	client.conn = conn
	if size <= 0 {
		client.reader = bufio.NewReader(conn)

	} else {
		client.reader = bufio.NewReaderSize(conn, size)
	}

	if out {
		client.locker = new(sync.Mutex)

	} else {
		client = nil
	}

	return client
}

func (c ClientSocket) Read() (error, []byte, *bufio.Reader) {
	return nil, nil, c.reader
}

func (c ClientSocket) Sticky() bool {
	return true
}

func (c ClientSocket) Output() (err error, out bool, locker sync.Locker) {
	return nil, locker != nil, locker
}

func (c ClientSocket) Write(bs []byte, out bool) (err error) {
	if !out && c.locker != nil {
		c.locker.Lock()
		defer c.locker.Unlock()
	}

	_, err = c.conn.Write(bs)
	return
}

func (c ClientSocket) Close() {
	c.conn.Close()
}

type ClientWebsocket struct {
	conn *websocket.Conn
}

func (c ClientWebsocket) Read() (error, []byte, *bufio.Reader) {
	var bs []byte
	err := websocket.Message.Receive(c.conn, &bs)
	if err != nil {
		return err, nil, nil
	}

	return nil, bs, nil
}

func (c ClientWebsocket) Sticky() bool {
	return false
}

func (c ClientWebsocket) Output() (err error, out bool, locker sync.Locker) {
	return nil, false, nil
}

func (c ClientWebsocket) Write(bs []byte, out bool) (err error) {
	err = websocket.Message.Send(c.conn, bs)
	return
}

func (c ClientWebsocket) Close() {
	c.conn.Close()
}

func NewClientWebsocket(conn *websocket.Conn, size int) *ClientWebsocket {
	client := new(ClientWebsocket)
	client.conn = conn
	return client
}
