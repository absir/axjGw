package main

import (
	"axj/ANet"
	"axj/Thrd/AZap"
	"go.uber.org/zap"
	"io"
	"net"
	"strconv"
	"sync"
)

type ConnId struct {
	id        int32
	locker    sync.Locker
	closed    bool
	conn      *net.TCPConn
	aConn     ANet.Conn
	aConnBuff []byte
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
		if nErr, ok := err.(*net.OpError); ok {
			AZap.Debug("ConnId Close %d, %v", that.id, nErr.Error())

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
		if that.onError(err) {
			return
		}

		if n > 0 {
			that.aConnWrite(n)
		}
	}
}

// aConn代理连接 数据接收 写入到 conn本地连接
func (that *ConnId) aConnLoop() {
	var buff []byte = nil
	var n int
	for !that.closed {
		// 循环写入
		err, data, reader := that.aConn.ReadA()
		if that.onError(err) {
			return
		}

		if reader == nil {
			that.conn.Write(data)

		} else {
			if buff == nil {
				buff = make([]byte, len(that.aConnBuff))
			}

			n, err = reader.Read(buff)
			if that.onError(err) {
				return
			}

			if n > 0 {
				_, err = that.conn.Write(buff[:n])
				if that.onError(err) {
					return
				}
			}
		}
	}
}
