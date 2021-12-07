package ANet

import (
	"axj/Kt/KtBuffer"
	"axj/Kt/KtBytes"
	"axj/Kt/KtIo"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/Util"
	"errors"
	"io"
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

// 最大FRAME数
var FRAME_MAX = 128

// CRC错误
var ERR_CRC = errors.New("CRC")
var ERR_MAX = errors.New("MAX")
var ERR_FRAME_MAX = errors.New("FRAME_MAX")

type ReqFrame struct {
	Head   byte
	Req    int32
	Uri    string
	UriI   int32
	Fid    int64
	Data   []byte
	Buffer *KtBuffer.Buffer
}

type FrameReader struct {
	ReqFrame
	readerB int8
	readerI int
	i32     int32
	i64     int64
}

func (that *FrameReader) IsDone() bool {
	return that.readerB == -1
}

func (that *FrameReader) DoneFrame() *ReqFrame {
	if that.readerB == -1 {
		frame := that.ReqFrame
		that.Reset()
		return &frame
	}

	return nil
}

func (that *FrameReader) Reset() {
	that.Req = 0
	that.Uri = ""
	that.UriI = 0
	that.Fid = 0
	// next set nil
	// that.Data = nil
	that.Buffer = nil
	that.readerB = 0
	that.next()
}

func (that *FrameReader) next() {
	that.Data = nil
	that.readerI = 0
	that.i32 = 0
	that.i64 = 0
}

func (that *FrameReader) readByte(pBs *[]byte) (byte, bool) {
	bs := *pBs
	bLen := len(bs)
	if bLen > 0 {
		b := bs[0]
		*pBs = bs[1:]
		return b, true
	}

	return 0, false
}

func (that *FrameReader) readVInt(pBs *[]byte) int32 {
	bs := *pBs
	bLen := len(bs)
	i32 := that.i32
	for i := 0; i < bLen; i++ {
		b := bs[i]
		switch that.readerI {
		case 0:
			i32 += int32(b) & KtBytes.VINT
			break
		case 1:
			i32 += int32(b&KtBytes.VINT_B) << 7
			break
		case 2:
			i32 += int32(b&KtBytes.VINT_B) << 14
			break
		case 3:
			i32 += int32(b) << 21
			break
		}

		if that.readerI >= 3 || (b&KtBytes.VINT_NB) == 0 {
			*pBs = bs[i+1:]
			that.readerI = 0
			that.i32 = 0
			return i32

		} else {
			that.readerI++
		}
	}

	return -1
}

func (that *FrameReader) readLong(pBs *[]byte) int64 {
	bs := *pBs
	bLen := len(bs)
	i64 := that.i64
	for i := 0; i < bLen; i++ {
		b := bs[i]
		switch that.readerI {
		case 0:
			i64 += int64(b)
			break
		case 1:
			i64 += int64(b) << 8
			break
		case 2:
			i64 += int64(b) << 16
			break
		case 3:
			i64 += int64(b) << 24
			break
		case 4:
			i64 += int64(b) << 32
			break
		case 5:
			i64 += int64(b) << 40
			break
		case 6:
			i64 += int64(b) << 48
			break
		case 7:
			i64 += int64(b) << 56
			break
		}

		if that.readerI >= 7 {
			*pBs = bs[i+1:]
			that.readerI = 0
			that.i64 = 0
			return i64

		} else {
			that.readerI++
		}
	}

	return -1
}

func (that *FrameReader) readDataLen(pBs *[]byte, dataMax int32, bufferP bool) (bool, error) {
	dLen := that.readVInt(pBs)
	if dLen >= 0 {
		if dLen > dataMax {
			return false, ERR_MAX
		}

		if bufferP {
			that.Data = Util.GetBufferBytes(int(dLen), &that.Buffer)

		} else {
			that.Data = make([]byte, dLen)
		}

		that.i32 = dLen
		return true, nil
	}

	return false, nil
}

func (that *FrameReader) readData(pBs *[]byte) []byte {
	bs := *pBs
	bLen := len(bs)
	dLen := int(that.i32)
	if dLen <= bLen {
		copy(that.Data[that.readerI:], bs[:dLen])
		*pBs = bs[dLen:]
		data := that.Data
		that.Data = nil
		that.readerI = 0
		that.i32 = 0
		return data

	} else {
		copy(that.Data[that.readerI:], bs)
		that.readerI += bLen
		that.i32 -= int32(bLen)
		return nil
	}
}

// 数据协议
type Protocol interface {
	// 请求读取
	Req(bs []byte, pid *int64) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 流请求读取
	ReqReader(reader Reader, sticky bool, pid *int64, dataMax int32, pBuffer **KtBuffer.Buffer) (err error, head byte, req int32, uri string, uriI int32, data []byte)
	// 帧请求读取
	ReqFrame(pBs *[]byte, reader *FrameReader, dataMax int32, bufferP bool) error
	// 返回数据 dLength data强制长度，不copydata
	Rep(req int32, uri string, uriI int32, data []byte, dLength int32, sticky bool, head byte, id int64, pBuffer **KtBuffer.Buffer) ([]byte, int32)
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

func (that *ProtocolV) ReqReader(reader Reader, sticky bool, pid *int64, dataMax int32, pBuffer **KtBuffer.Buffer) (err error, head byte, req int32, uri string, uriI int32, data []byte) {
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
		if bLen > dataMax {
			err = ERR_MAX
			return
		}

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

			if pBuffer == nil {
				data, err = KtIo.ReadBytesReader(reader, int(bLen))

			} else {
				dLen := int(bLen)
				data = Util.GetBufferBytes(dLen, pBuffer)
				data, err = KtIo.ReadBytesReaderBsLen(reader, data, dLen)
			}

		} else {
			// 不粘包
			data, err = io.ReadAll(reader)
		}
	}

	return
}

func (that *ProtocolV) ReqFrame(pBs *[]byte, reader *FrameReader, dataMax int32, bufferP bool) error {
	for {
		switch reader.readerB {
		case 0:
			// 读取头
			b, ok := reader.readByte(pBs)
			if !ok {
				return nil
			}

			reader.Head = b
			// 头校验
			if !that.isCrc(reader.Head) {
				return ERR_CRC
			}

			reader.readerB++
			break
		case 1:
			if (reader.Head & HEAD_REQ) != 0 {
				// 读取Req
				reader.Req = reader.readVInt(pBs)
				if reader.Req < 0 {
					return nil
				}
			}
			reader.readerB++
			break
		case 2:
			if (reader.Head & HEAD_URI) != 0 {
				// 读取Uri
				ok, err := reader.readDataLen(pBs, dataMax, false)
				if err != nil {
					return err
				}

				if !ok {
					return nil
				}

				// 转到读取Uri[string]
				reader.readerB++

			} else {
				// 直接转到读取UriI
				reader.readerB += 2
			}
			break
		case 3:
			// 读取Uri[string]
			data := reader.readData(pBs)
			if data == nil {
				return nil
			}

			reader.Uri = KtUnsafe.BytesToString(data)
			reader.readerB++
			break
		case 4:
			if (reader.Head & HEAD_URI_I) != 0 {
				// 读取UriI
				reader.UriI = reader.readVInt(pBs)
				if reader.UriI < 0 {
					return nil
				}
			}
			reader.readerB++
			break
		case 5:
			// 读取fid
			if reader.Req == REQ_PUSHI {
				reader.Fid = reader.readLong(pBs)
				if reader.Fid < 0 {
					return nil
				}
			}
			reader.readerB++
		case 6:
			if (reader.Head & HEAD_DATA) != 0 {
				// 读取DATA
				ok, err := reader.readDataLen(pBs, dataMax, bufferP)
				if err != nil {
					return err
				}

				if !ok {
					return nil
				}

				// 转到读取Data[[]byte]
				reader.readerB++

			} else {
				// 直接转到frameDone
				reader.readerB += 2
			}
			break
		case 7:
			// 读取Data[[]byte]
			data := reader.readData(pBs)
			if data == nil {
				return nil
			}

			reader.Data = data
			reader.readerB++
			break
		default:
			// frameDone
			reader.readerB = -1
			return nil
		}
	}
}

func (that *ProtocolV) Rep(req int32, uri string, uriI int32, data []byte, dLength int32, sticky bool, head byte, id int64, pBuffer **KtBuffer.Buffer) ([]byte, int32) {
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

	bs := Util.GetBufferBytes(int(bLen), pBuffer)
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
