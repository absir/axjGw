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
	Output() (error, bool, sync.Locker)
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

func (c ClientSocket) init() {
	c.locker = nil
}

func (c ClientSocket) open(conn *net.TCPConn, size int, out bool) {
	c.conn = conn
	if size <= 0 {
		c.reader = bufio.NewReader(conn)

	} else {
		c.reader = bufio.NewReaderSize(conn, size)
	}

	if out {
		if c.locker == nil {
			c.locker = new(sync.Mutex)
		}

	} else {
		c.locker = nil
	}
}

func (c ClientSocket) close() {
	c.conn = nil
	c.reader = nil
}

var clientSocketPool *sync.Pool = nil

func SetClientSocketPool() {
	pool := new(sync.Pool)
	pool.New = func() interface{} {
		client := new(ClientSocket)
		client.init()
		return client
	}
}

func NewClientSocket(conn *net.TCPConn, size int, out bool) *ClientSocket {
	var client *ClientSocket
	if clientSocketPool == nil {
		client = new(ClientSocket)
		client.init()

	} else {
		client = clientSocketPool.Get().(*ClientSocket)
	}

	client.open(conn, size, out)
	return client
}

func (c ClientSocket) Read() (error, []byte, *bufio.Reader) {
	return nil, nil, c.reader
}

func (c ClientSocket) Sticky() bool {
	return true
}

func (c ClientSocket) Output() (error, bool, sync.Locker) {
	return nil, c.locker != nil, c.locker
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
	if clientSocketPool != nil {
		c.close()
		clientSocketPool.Put(c)
	}

	c.conn.Close()
}

type ClientWebsocket websocket.Conn

func (c ClientWebsocket) Conn() *websocket.Conn {
	conn := websocket.Conn(c)
	return &conn
}

func (c *ClientWebsocket) Read() (error, []byte, *bufio.Reader) {
	var bs []byte
	err := websocket.Message.Receive(c.Conn(), &bs)
	if err != nil {
		return err, nil, nil
	}

	return nil, bs, nil
}

func (c ClientWebsocket) Sticky() bool {
	return false
}

func (c ClientWebsocket) Output() (error, bool, sync.Locker) {
	return nil, false, nil
}

func (c ClientWebsocket) Write(bs []byte, out bool) (err error) {
	err = websocket.Message.Send(c.Conn(), bs)
	return
}

func (c ClientWebsocket) Close() {
	c.Conn().Close()
}

func NewClientWebsocket(conn *websocket.Conn) *ClientWebsocket {
	if conn == nil {
		return nil
	}

	client := ClientWebsocket(*conn)
	return &client
}
