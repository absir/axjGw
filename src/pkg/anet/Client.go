package anet

import (
	"axj/KtBytes"
	"axj/KtUnsafe"
	"bufio"
	"errors"
	"io"
	"sync"
)

const (
	// 头标识
	HEAD_REQ       byte = 0x01 << 7          //请求
	HEAD_URI       byte = 0x01 << 6          //请求路由字符
	HEAD_URI_I     byte = 0x01 << 5          //请求路由压缩
	HEAD_DATA      byte = 0x01 << 4          //请求数据
	HEAD_COMPRESS  byte = 0x01 << 3          //数据压缩
	HEAD_ENCRY     byte = 0x01 << 2          //数据加密
	HEAD_CRC_MSK_M byte = 0x01 << 2          //头数据校验MOD
	HEAD_CRC_MSK   byte = HEAD_CRC_MSK_M - 1 //头数据校验
	HEAD_CRC_MSK_N byte = ^HEAD_CRC_MSK      //头数据校验取反

	// 特殊请求
	REQ_BEAT  int32 = 1 // 心跳
	REQ_PUSH  int32 = 2 // 推送
	REQ_URI   int32 = 3 // 路由查询
	REQ_ROUTE int32 = 4 // 路由表HASH
)

var CRC_ERR = errors.New("CRC")

type Client interface {
	// 读取
	Read() (error, []byte, *bufio.Reader)
	// 流写入
	Output() (locker sync.Locker, writer io.Writer, sticky bool)
	// 写入
	Write([]byte) (err error, succ bool)
}

// 数据协议
type Protocol interface {
	// 请求读取
	Req(bs []byte) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 流请求读取
	ReqReader(reader bufio.Reader, sticky bool) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 返回数据
	Rep(req int32, uri string, uriI int32, data []byte, sticky bool, head byte) []byte
	// 返回流写入
	RepClient(locker sync.Locker, client Client, buff *[]byte, req int32, uri string, uriI int32, data []byte, sticky bool, head byte) (err error, succ bool)
}

// 数据压缩
type Compress interface {
	// 压缩
	Compress(data []byte) []byte
	// 解压
	UnCompress(data []byte) []byte
}

// 数据加密
type Encry interface {
	// 生成密钥
	NewKeys() ([]byte, []byte)
	// 加密
	Encry(data []byte, key []byte) []byte
	// 解密
	Decry(data []byte, key []byte) []byte
}

type ProtocolV struct {
}

func (p ProtocolV) crc(head byte) byte {
	return (head & HEAD_CRC_MSK_N) % HEAD_CRC_MSK_M
}

func (p ProtocolV) Req(bs []byte) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
	head = bs[0]
	// 头部校验
	if p.crc(head) != (head & HEAD_CRC_MSK) {
		err = CRC_ERR
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

func (p ProtocolV) ReqReader(reader *bufio.Reader, sticky bool) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
	head, err = reader.ReadByte()
	if err != nil {
		return
	}

	// 头部校验
	if p.crc(head) != (head & HEAD_CRC_MSK) {
		err = CRC_ERR
		return
	}

	// 数据准备
	req = 0
	uri = ""
	uriI = 0
	data = nil

	// 请求解析
	if (head & HEAD_REQ) != 0 {
		req = KtBytes.GetVIntReader(reader)
	}

	// 路由解析
	if (head & HEAD_URI) != 0 {
		bLen := KtBytes.GetVIntReader(reader)
		var bs []byte
		bs, err = KtBytes.ReadBytesReader(reader, int(bLen))
		if err != nil {
			return
		}

		uri = KtUnsafe.BytesToString(bs)
	}

	// 路由压缩解析
	if (head & HEAD_URI_I) != 0 {
		uriI = KtBytes.GetVIntReader(reader)
	}

	// 数据解析
	if (head & HEAD_DATA) != 0 {
		if sticky {
			// 粘包
			bLen := KtBytes.GetVIntReader(reader)
			data, err = KtBytes.ReadBytesReader(reader, int(bLen))

		} else {
			// 不粘包
			data, err = io.ReadAll(reader)
		}
	}

	return
}

func (p ProtocolV) Rep(req int32, uri string, uriI int32, data []byte, sticky bool, head byte) []byte {
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
	head |= p.crc(head)
	return bs
}

func (p ProtocolV) RepClient(locker sync.Locker, client Client, buff *[]byte, req int32, uri string, uriI int32, data []byte, sticky bool, head byte) (err error, succ bool) {
	err = nil
	succ = true
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
	head |= p.crc(head)
	_buff[0] = head
	err, succ = client.Write(_buff[0:1])
	if err != nil || !succ {
		return
	}

	var off int32 = 0
	// 写入请求
	if req > 0 {
		KtBytes.SetVInt(_buff, 0, req, &off)
		err, succ = client.Write(_buff[0:off])
		if err != nil || !succ {
			return
		}
	}

	// 写入路由
	if uLen > 0 {
		KtBytes.SetVInt(_buff, 0, uLen, &off)
		err, succ = client.Write(_buff[0:off])
		if err != nil || !succ {
			return
		}

		err, succ = client.Write(KtUnsafe.StringToBytes(uri))
		if err != nil || !succ {
			return
		}
	}

	// 写入路由压缩
	if uriI > 0 {
		KtBytes.SetVInt(_buff, 0, uriI, &off)
		err, succ = client.Write(_buff[0:off])
		if err != nil || !succ {
			return
		}
	}

	// 写入数据
	if dLen > 0 {
		// 数据
		if sticky {
			KtBytes.SetVInt(_buff, 0, dLen, &off)
			err, succ = client.Write(_buff[0:off])
			if err != nil || !succ {
				return
			}
		}

		err, succ = client.Write(data)
		if err != nil || !succ {
			return
		}
	}

	return
}
