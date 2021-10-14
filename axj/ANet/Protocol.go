package ANet

import (
	"axj/Kt/KtBytes"
	"axj/Kt/KtIo"
	"axj/Kt/KtUnsafe"
	"bufio"
	"errors"
	"io"
	"sync"
)

const (
	// 头状态
	HEAD_COMPRESS  byte = 0x01 << 7          //数据压缩
	HEAD_ENCRY     byte = 0x01 << 6          //数据加密
	HEAD_REQ       byte = 0x01 << 5          //请求
	HEAD_URI       byte = 0x01 << 4          //请求路由字符
	HEAD_URI_I     byte = 0x01 << 3          //请求路由压缩
	HEAD_DATA      byte = 0x01 << 2          //请求数据
	HEAD_CRC_MSK_M byte = 0x01 << 2          //头数据校验MOD
	HEAD_CRC_MSK        = HEAD_CRC_MSK_M - 1 //头数据校验
	HEAD_CRC_MSK_N      = ^HEAD_CRC_MSK      //头数据校验取反
)

// CRC错误
var ERR_CRC = errors.New("CRC")
var ERR_MAX = errors.New("MAX")

// 数据协议
type Protocol interface {
	// 请求读取
	Req(bs []byte) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 流请求读取
	ReqReader(reader *bufio.Reader, sticky bool, dataMax int32) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 返回数据
	Rep(req int32, uri string, uriI int32, data []byte, sticky bool, head byte) []byte
	// 返回流写入
	RepOut(locker sync.Locker, client Client, buff *[]byte, req int32, uri string, uriI int32, data []byte, sticky bool, head byte) (err error)
	// 批量返回数据头
	RepBH(req int32, uri string, uriI int32, data bool, head byte) []byte
	RepBS(bh []byte, data []byte, sticky bool, head byte) []byte
	RepOutBS(locker sync.Locker, client Client, buff *[]byte, bh []byte, data []byte, sticky bool, head byte) (err error)
}

type ProtocolV struct {
}

func (that ProtocolV) crc(head byte) byte {
	return (head & HEAD_CRC_MSK_N) % HEAD_CRC_MSK_M
}

func (that ProtocolV) Req(bs []byte) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
	head = bs[0]
	// 头部校验
	if that.crc(head) != (head & HEAD_CRC_MSK) {
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

	// 数据解析
	if (head & HEAD_DATA) != 0 {
		data = bs[off:]
	}

	return
}

func (that ProtocolV) ReqReader(reader *bufio.Reader, sticky bool, dataMax int32) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
	head, err = reader.ReadByte()
	if err != nil {
		return
	}

	// 头部校验
	if that.crc(head) != (head & HEAD_CRC_MSK) {
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

func (that ProtocolV) Rep(req int32, uri string, uriI int32, data []byte, sticky bool, head byte) []byte {
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
		dLen = int32(len(data))
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
		copy(KtUnsafe.StringToBytes(uri), bs[off:])
		off += uLen
	}

	if uriI > 0 {
		// 路由压缩
		KtBytes.SetVInt(bs, off, uriI, &off)
	}

	if dLen > 0 {
		// 数据
		if sticky {
			KtBytes.SetVInt(bs, off, dLen, &off)
		}

		copy(data, bs[off:])
	}

	// 最后设置头
	bs[0] = head
	head |= that.crc(head)
	return bs
}

func (that ProtocolV) RepOut(locker sync.Locker, client Client, buff *[]byte, req int32, uri string, uriI int32, data []byte, sticky bool, head byte) (err error) {
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
		defer locker.Unlock()
		locker.Lock()
	}

	// buff准备
	_buff := *buff
	if _buff == nil {
		_buff = make([]byte, 4)
		buff = &_buff
	}

	// 写入头
	head |= that.crc(head)
	_buff[0] = head
	err = client.Write(_buff[0:1], true)
	if err != nil {
		return
	}

	var off int32 = 0
	// 写入请求
	if req > 0 {
		KtBytes.SetVInt(_buff, 0, req, &off)
		err = client.Write(_buff[0:off], true)
		if err != nil {
			return
		}
	}

	// 写入路由
	if uLen > 0 {
		KtBytes.SetVInt(_buff, 0, uLen, &off)
		err = client.Write(_buff[0:off], true)
		if err != nil {
			return
		}

		err = client.Write(KtUnsafe.StringToBytes(uri), true)
		if err != nil {
			return
		}
	}

	// 写入路由压缩
	if uriI > 0 {
		KtBytes.SetVInt(_buff, 0, uriI, &off)
		err = client.Write(_buff[0:off], true)
		if err != nil {
			return
		}
	}

	// 写入数据
	if dLen > 0 {
		// 数据
		if sticky {
			KtBytes.SetVInt(_buff, 0, dLen, &off)
			err = client.Write(_buff[0:off], true)
			if err != nil {
				return
			}
		}

		err = client.Write(data, true)
		if err != nil {
			return
		}
	}

	return
}

func (that ProtocolV) RepBH(req int32, uri string, uriI int32, data bool, head byte) []byte {
	if data {
		head |= HEAD_DATA
	}

	return that.Rep(req, uri, uriI, nil, false, head)
}

func (that ProtocolV) RepBS(bh []byte, data []byte, sticky bool, head byte) []byte {
	// 数据长度
	var dLen int32 = 0
	if data != nil {
		dLen = int32(len(data))
	}

	if dLen > 0 {
		// 数据粘包
		bLen := int32(len(bh)) + dLen
		if sticky {
			bLen += KtBytes.GetVIntLen(dLen)
		}

		// 新数据包
		bs := make([]byte, bLen)
		// 头数据
		copy(bh, bs)

		// 粘包
		if sticky {
			KtBytes.SetVInt(bs, dLen, dLen, &dLen)
		}

		// 数据
		copy(data, bs[dLen:])

		// 头处理
		head |= bs[0]
		head |= HEAD_DATA
		if head != bs[0] {
			head &= HEAD_CRC_MSK_N
			head |= that.crc(head)
			bs[0] = head
		}

		return bs
	}

	return bh
}

func (that ProtocolV) RepOutBS(locker sync.Locker, client Client, buff *[]byte, bh []byte, data []byte, sticky bool, head byte) (err error) {
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
		head &= HEAD_CRC_MSK_N
		head |= that.crc(head)
	}

	// 写入锁
	if locker != nil {
		defer locker.Unlock()
		locker.Lock()
	}

	// buff准备
	_buff := *buff
	if _buff == nil {
		_buff = make([]byte, 4)
		buff = &_buff
	}

	// 写入批量头状态
	_buff[0] = head
	err = client.Write(_buff[0:1], true)
	if err != nil {
		return
	}

	// 写入批量头
	err = client.Write(bh[1:], true)
	if err != nil {
		return
	}

	if dLen > 0 {
		// 写入粘包
		if sticky {
			var off int32
			KtBytes.SetVInt(_buff, 0, dLen, &off)
			err = client.Write(_buff[0:off], true)
			if err != nil {
				return
			}
		}

		// 写入数据
		err = client.Write(data, true)
		if err != nil {
			return
		}
	}

	return
}
