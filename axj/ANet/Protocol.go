package ANet

import (
	"axj/Kt/KtBytes"
	"axj/Kt/KtIo"
	"axj/Kt/KtUnsafe"
	"errors"
	"io"
	"sync"
)

const (
	// 头状态
	HEAD_COMPRESS  byte = 0x01 << 7              //数据压缩
	HEAD_ENCRY     byte = 0x01 << 6              //数据加密
	HEAD_REQ       byte = 0x01 << 5              //请求
	HEAD_URI       byte = 0x01 << 4              //请求路由字符
	HEAD_URI_I     byte = 0x01 << 3              //请求路由压缩
	HEAD_DATA      byte = 0x01 << 2              //请求数据
	HEAD_CRC_MSK_N      = 2                      //头数据校验位数
	HEAD_CRC_MSK_M byte = 0x01 << HEAD_CRC_MSK_N //头数据校验MOD
	HEAD_CRC_MSK        = HEAD_CRC_MSK_M - 1     //头数据校验

)

// CRC错误
var ERR_CRC = errors.New("CRC")
var ERR_MAX = errors.New("MAX")

// 数据协议
type Protocol interface {
	// 请求读取
	Req(bs []byte, pid *int64) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 流请求读取
	ReqReader(reader Reader, sticky bool, pid *int64, dataMax int32) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 返回数据 dLength data强制长度，不copydata
	Rep(req int32, uri string, uriI int32, data []byte, dLength int32, sticky bool, head byte, id int64) ([]byte, int32)
	// 返回流写入
	RepOut(locker sync.Locker, conn Conn, buff *[]byte, req int32, uri string, uriI int32, data []byte, head byte, id int64) (err error)
	// 批量返回数据头
	RepBH(req int32, uri string, uriI int32, data bool, head byte) []byte
	// 批量通过RepBH生成RepBs数据 dLength data强制长度，不copydata
	RepBS(bh []byte, data []byte, dLength int32, sticky bool, head byte) ([]byte, int32)
	RepOutBS(locker sync.Locker, conn Conn, buff *[]byte, bh []byte, data []byte, head byte) (err error)
}

type ProtocolV struct {
}

func (that *ProtocolV) crc(head byte) byte {
	h := head >> HEAD_CRC_MSK_N
	return (h << HEAD_CRC_MSK_N) | (h % HEAD_CRC_MSK_M)
}

func (that *ProtocolV) isCrc(head byte) bool {
	return ((head >> HEAD_CRC_MSK_N) % HEAD_CRC_MSK_M) == (head & HEAD_CRC_MSK)
}

func (that *ProtocolV) Req(bs []byte, pid *int64) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
	head = bs[0]
	// 头部校验
	if !that.isCrc(head) {
		err = ERR_CRC
		return
	}

	// 数据准备
	req = 0
	uri = ""
	uriI = 0
	data = nil
	var off int32 = 1

	// 请求解析
	if (head & HEAD_REQ) != 0 {
		req = KtBytes.GetVInt(bs, off, &off)
	}

	// 路由解析
	if (head & HEAD_URI) != 0 {
		bLen := KtBytes.GetVInt(bs, off, &off)
		end := off + bLen
		uri = KtUnsafe.BytesToString(bs[off:end])
		off = end
	}

	// 路由压缩解析
	if (head & HEAD_URI_I) != 0 {
		uriI = KtBytes.GetVInt(bs, off, &off)
	}

	// pid 读取
	if pid != nil && req == REQ_PUSHI {
		*pid = KtBytes.GetInt64(bs, off, &off)
	}

	// 数据解析
	if (head & HEAD_DATA) != 0 {
		data = bs[off:]
	}

	return
}

func (that *ProtocolV) ReqReader(reader Reader, sticky bool, pid *int64, dataMax int32) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
	head, err = reader.ReadByte()
	if err != nil {
		return
	}

	// 头部校验
	if !that.isCrc(head) {
		err = ERR_CRC
		return
	}

	// 数据准备
	req = 0
	uri = ""
	uriI = 0
	data = nil

	// 请求解析
	if (head & HEAD_REQ) != 0 {
		req = KtIo.GetVIntReader(reader)
	}

	// 路由解析
	if (head & HEAD_URI) != 0 {
		bLen := KtIo.GetVIntReader(reader)
		var bs []byte
		bs, err = KtIo.ReadBytesReader(reader, int(bLen))
		if err != nil {
			return
		}

		uri = KtUnsafe.BytesToString(bs)
	}

	// 路由压缩解析
	if (head & HEAD_URI_I) != 0 {
		uriI = KtIo.GetVIntReader(reader)
	}

	// pid 读取
	if pid != nil && req == REQ_PUSHI {
		*pid = KtIo.GetInt64Reader(reader)
	}

	// 数据解析
	if (head & HEAD_DATA) != 0 {
		if sticky {
			// 粘包
			bLen := KtIo.GetVIntReader(reader)
			if bLen > dataMax {
				err = ERR_MAX
				return
			}

			data, err = KtIo.ReadBytesReader(reader, int(bLen))

		} else {
			// 不粘包
			data, err = io.ReadAll(reader)
		}
	}

	return
}

func (that *ProtocolV) Rep(req int32, uri string, uriI int32, data []byte, dLength int32, sticky bool, head byte, id int64) ([]byte, int32) {
	if req == REQ_PUSHI {
		if id == 0 {
			req = REQ_PUSH
		}

	} else {
		id = 0
	}

	// 头长度
	var bLen int32 = 1

	rLen := KtBytes.GetVIntLen(req)
	if req > 0 {
		head |= HEAD_REQ
		// 请求长度
		bLen += rLen
	}

	uLen := int32(len(uri))
	if uLen > 0 {
		head |= HEAD_URI
		// 路由长度
		bLen += KtBytes.GetVIntLen(uLen)
		bLen += uLen
	}

	if uriI > 0 {
		head |= HEAD_URI_I
		// 路由压缩长度
		bLen += KtBytes.GetVIntLen(uriI)
	}

	var dLen int32 = 0
	if data != nil {
		// 强制data大小
		if dLength > 0 {
			dLen = dLength

		} else {
			dLen = int32(len(data))
		}
	}

	if req == REQ_PUSHI {
		bLen += 8
	}

	if dLen > 0 {
		head |= HEAD_DATA
		// 数据长度
		if sticky {
			bLen += KtBytes.GetVIntLen(dLen)
		}

		bLen += dLen
	}

	bs := make([]byte, bLen)
	var off int32 = 1
	if req > 0 {
		// 请求
		KtBytes.SetVInt(bs, off, req, &off)
	}

	if uLen > 0 {
		// 路由
		KtBytes.SetVInt(bs, off, uLen, &off)
		copy(bs[off:], KtUnsafe.StringToBytes(uri))
		off += uLen
	}

	if uriI > 0 {
		// 路由压缩
		KtBytes.SetVInt(bs, off, uriI, &off)
	}

	if req == REQ_PUSHI {
		// 设置消息编号
		KtBytes.SetInt64(bs, off, id, &off)
	}

	var dOff int32 = 0
	if dLen > 0 {
		// 数据
		if sticky {
			KtBytes.SetVInt(bs, off, dLen, &off)
		}

		if dLength > 0 {
			dOff = off

		} else {
			copy(bs[off:], data)
		}
	}

	// 最后设置头
	head |= that.crc(head)
	bs[0] = head
	return bs, dOff
}

func (that *ProtocolV) RepOut(locker sync.Locker, conn Conn, buff *[]byte, req int32, uri string, uriI int32, data []byte, head byte, id int64) (err error) {
	if req == REQ_PUSHI {
		if id == 0 {
			req = REQ_PUSH
		}

	} else {
		id = 0
	}

	err = nil
	// 头状态准备
	if req > 0 {
		head |= HEAD_REQ
	}

	uLen := int32(len(uri))
	if uLen > 0 {
		head |= HEAD_URI
	}

	if uriI > 0 {
		head |= HEAD_URI_I
	}

	var dLen int32 = 0
	if data != nil {
		dLen = int32(len(data))
	}

	if dLen > 0 {
		head |= HEAD_DATA
	}

	// 写入锁
	if locker != nil {
		locker.Lock()
	}

	// buff准备
	_buff := *buff
	if _buff == nil || len(_buff) < 4 {
		_buff = make([]byte, 4)
		*buff = _buff
	}

	// 写入头
	head |= that.crc(head)
	_buff[0] = head
	err = conn.Write(_buff[0:1])
	if err != nil {
		// 写入锁释放
		if locker != nil {
			locker.Unlock()
		}
		return
	}

	var off int32 = 0
	// 写入请求
	if req > 0 {
		KtBytes.SetVInt(_buff, 0, req, &off)
		err = conn.Write(_buff[0:off])
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}
	}

	// 写入路由
	if uLen > 0 {
		KtBytes.SetVInt(_buff, 0, uLen, &off)
		err = conn.Write(_buff[0:off])
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}

		err = conn.Write(KtUnsafe.StringToBytes(uri))
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}
	}

	// 写入路由压缩
	if uriI > 0 {
		KtBytes.SetVInt(_buff, 0, uriI, &off)
		err = conn.Write(_buff[0:off])
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}
	}

	// 写入消息编号
	if req == REQ_PUSHI {
		KtBytes.SetInt32(_buff, 0, int32(id), &off)
		err = conn.Write(_buff[0:off])
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}

		KtBytes.SetInt32(_buff, 0, int32(id>>32), &off)
		err = conn.Write(_buff[0:off])
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}
	}

	// 写入数据
	if dLen > 0 {
		// 数据
		if conn.Sticky() {
			KtBytes.SetVInt(_buff, 0, dLen, &off)
			err = conn.Write(_buff[0:off])
			if err != nil {
				// 写入锁释放
				if locker != nil {
					locker.Unlock()
				}
				return
			}
		}

		err = conn.Write(data)
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}
	}

	// 写入锁释放
	if locker != nil {
		locker.Unlock()
	}
	return
}

func (that *ProtocolV) RepBH(req int32, uri string, uriI int32, data bool, head byte) []byte {
	if data {
		head |= HEAD_DATA
	}

	bs, _ := that.Rep(req, uri, uriI, nil, 0, false, head, 0)
	return bs
}

func (that ProtocolV) RepBS(bh []byte, data []byte, dLength int32, sticky bool, head byte) ([]byte, int32) {
	// 数据长度
	var dLen int32 = 0
	if data != nil {
		if dLength > 0 {
			dLen = dLength

		} else {
			dLen = int32(len(data))
		}
	}

	if dLen > 0 {
		off := int32(len(bh))
		// 数据粘包
		bLen := off + dLen
		if sticky {
			bLen += KtBytes.GetVIntLen(dLen)
		}

		// 新数据包
		bs := make([]byte, bLen)
		// 头数据
		copy(bs, bh)

		// 粘包
		if sticky {
			KtBytes.SetVInt(bs, off, dLen, &off)
		}

		var dOff int32 = 0

		if dLength > 0 {
			dOff = off

		} else {
			// 数据
			copy(bs[off:], data)
		}

		// 头处理
		head |= bs[0]
		head |= HEAD_DATA
		if head != bs[0] {
			head = that.crc(head)
			bs[0] = head
		}

		return bs, dOff
	}

	return bh, 0
}

func (that *ProtocolV) RepOutBS(locker sync.Locker, conn Conn, buff *[]byte, bh []byte, data []byte, head byte) (err error) {
	err = nil
	var dLen int32 = 0
	if data != nil {
		dLen = int32(len(data))
	}

	// 头处理
	head |= bh[0]
	if dLen > 0 {
		head |= HEAD_DATA
	}

	if head != bh[0] {
		head = that.crc(head)
	}

	// 写入锁
	if locker != nil {
		locker.Lock()
	}

	// buff准备
	_buff := *buff
	if _buff == nil {
		_buff = make([]byte, 4)
		*buff = _buff
	}

	// 写入批量头状态
	_buff[0] = head
	err = conn.Write(_buff[0:1])
	if err != nil {
		// 写入锁释放
		if locker != nil {
			locker.Unlock()
		}
		return
	}

	// 写入批量头
	err = conn.Write(bh[1:])
	if err != nil {
		// 写入锁释放
		if locker != nil {
			locker.Unlock()
		}
		return
	}

	if dLen > 0 {
		// 写入粘包
		if conn.Sticky() {
			var off int32
			KtBytes.SetVInt(_buff, 0, dLen, &off)
			err = conn.Write(_buff[0:off])
			if err != nil {
				// 写入锁释放
				if locker != nil {
					locker.Unlock()
				}
				return
			}
		}

		// 写入数据
		err = conn.Write(data)
		if err != nil {
			// 写入锁释放
			if locker != nil {
				locker.Unlock()
			}
			return
		}
	}

	// 写入锁释放
	if locker != nil {
		locker.Unlock()
	}
	return
}
