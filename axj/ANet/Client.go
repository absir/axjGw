package ANet

import (
	"axj/Thrd/Util"
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

func (that ClientSocket) PInit() {
	that.locker = nil
}

func (that ClientSocket) PRelease() bool {
	if that.conn != nil {
		that.conn = nil
		that.reader = nil
		return true
	}

	return false
}

func (that ClientSocket) pInit(conn *net.TCPConn, size int, out bool) {
	that.conn = conn
	if size <= 0 {
		that.reader = bufio.NewReader(conn)

	} else {
		that.reader = bufio.NewReaderSize(conn, size)
	}

	if out {
		if that.locker == nil {
			that.locker = new(sync.Mutex)
		}

	} else {
		that.locker = nil
	}
}

var clientSocketPool = Util.NewAllocPool(false, func() Util.Pool {
	return new(ClientSocket)
})

func SetClientSocketPool(pool bool) {
	clientSocketPool.SetPool(pool)
}

func NewClientSocket(conn *net.TCPConn, size int, out bool) *ClientSocket {
	client := clientSocketPool.Get().(*ClientSocket)
	client.pInit(conn, size, out)
	return client
}

func (that ClientSocket) Read() (error, []byte, *bufio.Reader) {
	return nil, nil, that.reader
}

func (that ClientSocket) Sticky() bool {
	return true
}

func (that ClientSocket) Output() (error, bool, sync.Locker) {
	return nil, that.locker != nil, that.locker
}

func (that ClientSocket) Write(bs []byte, out bool) (err error) {
	if !out && that.locker != nil {
		that.locker.Lock()
		defer that.locker.Unlock()
	}

	_, err = that.conn.Write(bs)
	return
}

func (that ClientSocket) Close() {
	clientSocketPool.Put(that, false)
	that.conn.Close()
}

type ClientWebsocket websocket.Conn

func (that ClientWebsocket) Conn() *websocket.Conn {
	conn := websocket.Conn(that)
	return &conn
}

func (that ClientWebsocket) Read() (error, []byte, *bufio.Reader) {
	var bs []byte
	err := websocket.Message.Receive(that.Conn(), &bs)
	if err != nil {
		return err, nil, nil
	}

	return nil, bs, nil
}

func (that ClientWebsocket) Sticky() bool {
	return false
}

func (that ClientWebsocket) Output() (error, bool, sync.Locker) {
	return nil, false, nil
}

func (that ClientWebsocket) Write(bs []byte, out bool) (err error) {
	err = websocket.Message.Send(that.Conn(), bs)
	return
}

func (that ClientWebsocket) Close() {
	that.Conn().Close()
}

func NewClientWebsocket(conn *websocket.Conn) *ClientWebsocket {
	if conn == nil {
		return nil
	}

	client := ClientWebsocket(*conn)
	return &client
}
