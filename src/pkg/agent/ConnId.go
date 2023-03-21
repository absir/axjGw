package agent

import (
	"axj/ANet"
	"axj/ANets"
	"axj/Kt/Kt"
	"axj/Kt/KtBuffer"
	"axj/Kt/KtCvt"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/pkg/asdk"
	"go.uber.org/zap"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ConnId struct {
	id          int32
	locker      sync.Locker
	closed      bool
	conn        *net.TCPConn
	aConn       ANet.Conn
	aConnBuff   []byte
	aConnBuffer *KtBuffer.Buffer
}

var Client *asdk.Client
var CloseDelay int
var CloseDelayIn int

var DialBytes = make([]byte, 1)

func dialAddr(addr string, timeout time.Duration) error {
	if strings.HasPrefix(addr, "-") {
		// arp mac -> ip
		idx := strings.LastIndexByte(addr, ':')
		if idx > 0 {
			mac := addr[0:idx]
			ip := ANets.Config.FindIp(mac, 0)
			if ip == "" {
				return Kt.NewErrReason("Find No Ip " + mac)

			} else {
				addr = ip + addr[idx:]
			}
		}
	}

	var conn net.Conn
	var err error
	if timeout > 0 {
		conn, err = net.DialTimeout("tcp", addr, timeout)

	} else {
		conn, err = net.Dial("tcp", addr)
	}

	if err == nil && conn != nil {
		_, err = conn.Write(DialBytes)
	}

	//fmt.Printf("dialAddr %s %t \r\n", addr, err == nil)
	//Kt.Err(err, false)
	if conn != nil {
		conn.Close()
	}

	if err != nil {
		return err
	}

	return nil
}

func DialProxy(addr string, id int32, timeout time.Duration) {
	err := dialAddr(addr, timeout)
	req := REQ_DIAL
	if err != nil {
		req = REQ_DIAL_ERR
		AZap.Warn("REQ_DIAL Err "+addr+" , "+strconv.Itoa(int(id)), zap.Error(err))
	}

	adap := Client.Conn()
	if adap != nil {
		adap.Rep(Client, req, "", id, nil, false, 0)
	}
}

func ConnProxy(addr string, id int32, data []byte, buffer asdk.Buffer) {
	connId := &ConnId{
		id:     id,
		locker: new(sync.Mutex),
	}

	{
		// 代理缓冲大小
		buffSize := 0
		{
			idx := strings.IndexByte(addr, '/')
			if idx >= 0 {
				buffSize = int(KtCvt.ToInt32(addr[:idx]))
				addr = addr[idx+1:]
			}

			if buffSize < 256 {
				buffSize = 256
			}
		}

		// 本地连接
		{
			if strings.HasPrefix(addr, "-") {
				// arp mac -> ip
				idx := strings.LastIndexByte(addr, ':')
				if idx > 0 {
					mac := addr[0:idx]
					ip := ANets.Config.FindIp(mac, 0)
					if ip == "" {
						connId.onError(Kt.NewErrReason("Find No Ip " + mac))
						return

					} else {
						addr = ip + addr[idx:]
					}
				}
			}

			conn, err := net.Dial("tcp", addr)
			if connId.onError(err) || conn == nil {
				return
			}

			connId.conn = conn.(*net.TCPConn)
		}

		// 代理连接
		{
			conn, err := Client.DialConn()
			if connId.onError(err) || conn == nil {
				return
			}

			aConn, _ := conn.(ANet.Conn)

			connId.aConn = aConn
			// 代理连接协议
			aProcessor, _ := Client.GetProcessor().(*ANet.ProcessorV)
			if aConn == nil || aProcessor == nil {
				return
			}

			// 代理连接连接id请求
			err = aProcessor.Rep(true, aConn, nil, false, REQ_CONN, "", id, nil, false, 0)
			if connId.onError(err) {
				// 内存池释放
				asdk.BufferFree(buffer)
				return
			}

			// aConnBuff
			connId.aConnBuff = Util.GetBufferBytes(buffSize, &connId.aConnBuffer)
		}

		// 连接数据写入
		if data != nil {
			_, err := connId.conn.Write(data)
			// 内存池释放
			asdk.BufferFree(buffer)
			if connId.onError(err) {
				// 内存池回收
				Util.PutBuffer(connId.aConnBuffer)
				return
			}
		}
	}

	// 双向数据代理
	go connId.connLoop()
	connId.aConnLoop()
}

func (that *ConnId) onError(err error) bool {
	if err == nil {
		return false
	}

	if that.closed {
		return true
	}

	immed := that.aConnBuff == nil
	that.locker.Lock()
	if that.closed {
		that.locker.Unlock()
		return true
	}

	that.closed = true
	conn := that.conn
	if conn != nil {
		if immed {
			conn.SetLinger(0)
		}

		conn.Close()
	}

	that.locker.Unlock()
	if err != io.EOF {
		if _, ok := err.(*net.OpError); ok {
			AZap.Debug("ConnId Close %d, %v", that.id, err)

		} else {
			AZap.Warn("Conn Err "+strconv.Itoa(int(that.id)), zap.Error(err))
		}
	}

	if that.aConnBuff != nil {
		// 写入关闭软关闭
		if that.aConnWrite(0) == nil {
			return true
		}
	}

	// 直接关闭
	aConn := that.aConn
	if aConn != nil {
		aConn.Close(immed)
	}

	repClosedId(that.id)
	return true
}

// 代理连接 数据写入
func (that *ConnId) aConnWrite(n int) error {
	if n <= 0 {
		that.aConn.Close(false)
		return nil
	}

	return that.aConn.Write(that.aConnBuff[:n])
}

// conn本地连接 数据接收 写入到 aConn代理连接
func (that *ConnId) connLoop() {
	buff := that.aConnBuff
	for !that.closed {
		n, err := that.conn.Read(buff)
		if that.onError(err) || n <= 0 {
			break
		}

		err = that.aConnWrite(n)
		if that.onError(err) {
			break
		}
	}

	Util.PutBuffer(that.aConnBuffer)
	if CloseDelay > 0 {
		that.aConn.SetLinger(CloseDelay)
		that.aConn.Close(false)
	}
}

// aConn代理连接 数据接收 写入到 conn本地连接
func (that *ConnId) aConnLoop() {
	var buffer *KtBuffer.Buffer = nil
	var buff []byte = nil
	var n int
	for !that.closed {
		// 循环写入
		err, data, reader := that.aConn.ReadA()
		if that.onError(err) {
			break
		}

		if reader == nil {
			that.conn.Write(data)

		} else {
			if buff == nil {
				buff = Util.GetBufferBytes(len(that.aConnBuff), &buffer)
			}

			n, err = reader.Read(buff)
			if that.onError(err) {
				break
			}

			if n > 0 {
				_, err = that.conn.Write(buff[:n])
				if that.onError(err) {
					break
				}
			}
		}
	}

	// 内存池回收
	Util.PutBuffer(buffer)
	if CloseDelay > 0 {
		that.conn.SetLinger(CloseDelay)
		that.conn.Close()
	}
}

func repClosedId(id int32) {
	adap := Client.Conn()
	if adap != nil {
		//  关闭通知
		adap.Rep(Client, REQ_CLOSED, "", id, nil, false, 0)
	}
}
