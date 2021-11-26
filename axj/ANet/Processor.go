package ANet

import (
	"io"
	"sync"
)

type Processor struct {
	Protocol    Protocol
	Compress    Compress
	CompressMin int
	Encrypt     Encrypt
	DataMax     int32
}

func (that *Processor) Req(conn Conn, decryKey []byte) (error, int32, string, int32, []byte) {
	return req(conn, that.Protocol, that.Compress, that.Encrypt, decryKey, nil, that.DataMax)
}

func (that *Processor) ReqPId(conn Conn, decryKey []byte) (error, int32, string, int32, int64, []byte) {
	var pid int64 = 0
	err, req, uri, uriI, data := req(conn, that.Protocol, that.Compress, that.Encrypt, decryKey, &pid, that.DataMax)
	return err, req, uri, uriI, pid, data
}

func (that *Processor) ReqFrame(frame *ReqFrame, decryKey []byte) (err error, req int32, uri string, uriI int32, data []byte) {
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

func (that *Processor) Rep(locker sync.Locker, out bool, conn Conn, encryKey []byte, compress bool, req int32, uri string, uriI int32, data []byte, isolate bool, id int64) error {
	compr := that.Compress
	if !compress {
		compr = nil
	}

	return rep(locker, out, conn, that.Protocol, compr, that.CompressMin, that.Encrypt, encryKey, req, uri, uriI, data, false, isolate, id)
}

/*
 * 推送已压缩数据
 */
func (that *Processor) RepCData(locker sync.Locker, out bool, conn Conn, encryKey []byte, req int32, uri string, uriI int32, cData []byte, isolate bool, id int64) error {
	return rep(locker, out, conn, that.Protocol, that.Compress, that.CompressMin, that.Encrypt, encryKey, req, uri, uriI, cData, true, isolate, id)
}

func req(conn Conn, protocol Protocol, compress Compress, decrypt Encrypt, decryKey []byte, pId *int64, dataMax int32) (err error, req int32, uri string, uriI int32, data []byte) {
	err, bs, read := conn.ReadA()
	if err != nil {
		return
	}

	var head byte
	if bs != nil {
		err, head, req, uri, uriI, data = protocol.Req(bs, pId)

	} else if read != nil {
		err, head, req, uri, uriI, data = protocol.ReqReader(read, conn.Sticky(), pId, dataMax)

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

func rep(locker sync.Locker, out bool, conn Conn, protocol Protocol, compress Compress, compressMin int, encrypt Encrypt, encryKey []byte, req int32, uri string, uriI int32, data []byte, cData bool, isolate bool, id int64) (err error) {
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

	// 流写入
	var wBuff *[]byte = nil
	if out {
		wBuff = conn.Out()
	}

	var encryptLen int32 = 0
	if data != nil {
		if encrypt != nil && encryKey != nil {
			if wBuff == nil {
				encryptLen, err = encrypt.EncryptLength(data, encryKey)
				if err != nil {
					return err
				}
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

	if wBuff == nil {
		bs, off := protocol.Rep(req, uri, uriI, data, encryptLen, conn.Sticky(), head, id)
		if encryptLen > 0 {
			// 数据加密到
			err = encrypt.EncryptToDest(data, encryKey, bs[off:])
			if err != nil {
				return err
			}
		}

		return conn.Write(bs)

	} else {
		return protocol.RepOut(locker, conn, wBuff, req, uri, uriI, data, head, id)
	}
}

// 批量返回
type RepBatch struct {
	data     []byte
	protocol Protocol
	repOrBH  func(sticky bool) []byte
	bs       []byte
	bss      []byte
	bh       []byte
	bhs      []byte
}

func (that *RepBatch) init(protocol Protocol, compress Compress, compressMin int, req int32, uri string, uriI int32, data []byte) error {
	var head byte = 0
	dLen := 0
	if data != nil {
		dLen = len(data)
	}

	if dLen > 0 {
		head |= HEAD_DATA
		if compressMin > 0 && dLen > compressMin && compress != nil {
			bs, err := compress.Compress(data)
			if err != nil {
				return err
			}

			if len(bs) < dLen {
				head |= HEAD_COMPRESS
				data = bs
			}
		}

	} else {
		data = nil
	}

	that.data = data
	that.protocol = protocol
	that.repOrBH = func(sticky bool) []byte {
		if data == nil {
			bs, _ := protocol.Rep(req, uri, uriI, data, 0, sticky, head, 0)
			return bs

		} else {
			return protocol.RepBH(req, uri, uriI, true, head)
		}
	}

	that.bs = nil
	that.bss = nil
	that.bh = nil
	that.bhs = nil
	return nil
}

func (that *RepBatch) rep(locker sync.Locker, out bool, conn Conn, encrypt Encrypt, encryKey []byte) error {
	if that.data == nil {
		// 无数据写入
		if conn.Sticky() {
			if that.bss == nil {
				that.bss = that.repOrBH(true)
			}

			return conn.Write(that.bss)

		} else {
			if that.bs == nil {
				that.bs = that.repOrBH(false)
			}

			return conn.Write(that.bs)
		}
	}

	// 有数据通用头
	var bh []byte
	if conn.Sticky() {
		bh = that.bhs
		if bh == nil {
			bh = that.repOrBH(true)
			that.bhs = bh
		}

	} else {
		bh = that.bh
		if bh == nil {
			bh = that.repOrBH(false)
			that.bh = bh
		}
	}

	// 流写入
	var wBuff *[]byte = nil
	if out {
		wBuff = conn.Out()
	}

	var head byte = 0
	data := that.data
	var encryptLen int32 = 0
	var err error = nil
	if data != nil {
		if encrypt != nil && encryKey != nil {
			if wBuff == nil {
				encryptLen, err = encrypt.EncryptLength(that.data, encryKey)
				if err != nil {
					return err
				}
			}

			if encryptLen <= 0 {
				data, err = encrypt.Encrypt(data, encryKey, true)
				if err != nil {
					return err
				}
			}

			head |= HEAD_ENCRY
		}
	}

	if wBuff == nil {
		bs, off := that.protocol.RepBS(bh, data, encryptLen, conn.Sticky(), head)
		if encryptLen > 0 {
			// 数据加密到
			err = encrypt.EncryptToDest(data, encryKey, bs[off:])
			if err != nil {
				return err
			}
		}

		return conn.Write(bs)

	} else {
		return that.protocol.RepOutBS(locker, conn, wBuff, bh, data, head)
	}
}
