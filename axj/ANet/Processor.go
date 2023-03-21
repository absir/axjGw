package ANet

import (
	"axj/Kt/KtBuffer"
	"axj/Thrd/Util"
	"io"
)

type Processor interface {
	Get() *ProcessorV
	ReqOpen(i int, pBuffer **KtBuffer.Buffer, conn Conn, decryKey []byte) (error, int32, string, int32, []byte)
	Req(pBuffer **KtBuffer.Buffer, conn Conn, decryKey []byte) (error, int32, string, int32, []byte)
	ReqPId(pBuffer **KtBuffer.Buffer, conn Conn, decryKey []byte) (error, int32, string, int32, int64, []byte)
	ReqFrame(frame *ReqFrame, decryKey []byte) (err error, req int32, uri string, uriI int32, data []byte)
	Rep(bufferP bool, conn Conn, encryKey []byte, compress bool, req int32, uri string, uriI int32, data []byte, isolate bool, id int64) error
	RepCData(bufferP bool, conn Conn, encryKey []byte, req int32, uri string, uriI int32, cData []byte, isolate bool, id int64) error
	RepComprCData(bufferP bool, conn Conn, encryKey []byte, compr Compress, req int32, uri string, uriI int32, data []byte, cData bool, isolate bool, id int64) error
}

type ProcessorV struct {
	Protocol    Protocol
	Compress    Compress
	CompressMin int
	Encrypt     Encrypt
	DataMax     int32
}

func (that *ProcessorV) Get() *ProcessorV {
	return that
}

func (that *ProcessorV) ReqOpen(i int, pBuffer **KtBuffer.Buffer, conn Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	return that.Req(pBuffer, conn, decryKey)
}

func (that *ProcessorV) Req(pBuffer **KtBuffer.Buffer, conn Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	return req(pBuffer, conn, that.Protocol, that.Compress, that.Encrypt, decryKey, nil, that.DataMax)
}

func (that *ProcessorV) ReqPId(pBuffer **KtBuffer.Buffer, conn Conn, decryKey []byte) (error, int32, string, int32, int64, []byte) {
	var pid int64 = 0
	err, req, uri, uriI, data := req(pBuffer, conn, that.Protocol, that.Compress, that.Encrypt, decryKey, &pid, that.DataMax)
	return err, req, uri, uriI, pid, data
}

func (that *ProcessorV) ReqFrame(frame *ReqFrame, decryKey []byte) (err error, req int32, uri string, uriI int32, data []byte) {
	head := frame.Head
	data = frame.Data
	if data != nil {
		if (head&HEAD_ENCRY) != 0 && decryKey != nil && that.Encrypt != nil {
			data, err = that.Encrypt.Decrypt(data, decryKey)
			if err != nil {
				return
			}
		}

		if (head&HEAD_COMPRESS) != 0 && that.Compress != nil {
			data, err = that.Compress.UnCompress(data)
			if err != nil {
				return
			}
		}
	}

	req = frame.Req
	uri = frame.Uri
	uriI = frame.UriI
	return
}

func (that *ProcessorV) Rep(bufferP bool, conn Conn, encryKey []byte, compress bool, req int32, uri string, uriI int32, data []byte, isolate bool, id int64) error {
	compr := that.Compress
	if !compress {
		compr = nil
	}

	return that.RepComprCData(bufferP, conn, encryKey, compr, req, uri, uriI, data, false, isolate, id)
}

/*
 * 推送数据 cData已压缩状态
 */
func (that *ProcessorV) RepCData(bufferP bool, conn Conn, encryKey []byte, req int32, uri string, uriI int32, cData []byte, isolate bool, id int64) error {
	return that.RepComprCData(bufferP, conn, encryKey, nil, req, uri, uriI, cData, true, isolate, id)
}

func (that *ProcessorV) RepComprCData(bufferP bool, conn Conn, encryKey []byte, compr Compress, req int32, uri string, uriI int32, data []byte, cData bool, isolate bool, id int64) error {
	// 内存池
	var pBuffer **KtBuffer.Buffer = nil
	if bufferP {
		var buffer *KtBuffer.Buffer
		pBuffer = &buffer
	}

	err := rep(pBuffer, conn, that.Protocol, compr, that.CompressMin, that.Encrypt, encryKey, req, uri, uriI, data, cData, isolate, id)
	if bufferP {
		Util.PutBuffer(*pBuffer)
	}

	return err
}

func req(pBuffer **KtBuffer.Buffer, conn Conn, protocol Protocol, compress Compress, decrypt Encrypt, decryKey []byte, pId *int64, dataMax int32) (err error, req int32, uri string, uriI int32, data []byte) {
	err, bs, read := conn.ReadA()
	if err != nil {
		return
	}

	var head byte
	if bs != nil {
		err, head, req, uri, uriI, data = protocol.Req(bs, pId)

	} else if read != nil {
		err, head, req, uri, uriI, data = protocol.ReqReader(read, conn.Sticky(), pId, dataMax, pBuffer)

	} else {
		err = io.EOF
		return
	}

	if err != nil {
		return
	}

	if data != nil {
		// 数据处理
		if (head&HEAD_ENCRY) != 0 && decrypt != nil && decryKey != nil {
			data, err = decrypt.Decrypt(data, decryKey)
			if err != nil {
				return
			}
		}

		if (head&HEAD_COMPRESS) != 0 && compress != nil {
			data, err = compress.UnCompress(data)
			if err != nil {
				return
			}
		}
	}

	return
}

func rep(pBuffer **KtBuffer.Buffer, conn Conn, protocol Protocol, compress Compress, compressMin int, encrypt Encrypt, encryKey []byte, req int32, uri string, uriI int32, data []byte, cData bool, isolate bool, id int64) (err error) {
	if req < 0 {
		// 纯写入data
		return conn.Write(data)
	}

	var head byte = 0
	if data != nil {
		bLen := len(data)
		if bLen > 0 {
			// 数据处理
			if cData {
				head |= HEAD_COMPRESS

			} else if compress != nil {
				if compressMin > 0 && bLen > compressMin {
					var bs []byte
					bs, err = compress.Compress(data)
					if err != nil {
						return err
					}

					if len(bs) < bLen {
						head |= HEAD_COMPRESS
						data = bs
					}
				}
			}
		}
	}

	var encryptLen int32 = 0
	if data != nil {
		if encrypt != nil && encryKey != nil {
			encryptLen, err = encrypt.EncryptLength(data, encryKey)
			if err != nil {
				return err
			}

			if encryptLen <= 0 {
				data, err = encrypt.Encrypt(data, encryKey, isolate)
				if err != nil {
					return err
				}
			}

			head |= HEAD_ENCRY
		}
	}

	bs, off := protocol.Rep(req, uri, uriI, data, encryptLen, conn.Sticky(), head, id, pBuffer)
	if encryptLen > 0 {
		// 数据加密到
		err = encrypt.EncryptToDest(data, encryKey, bs[off:])
		if err != nil {
			return err
		}
	}

	return conn.Write(bs)
}
