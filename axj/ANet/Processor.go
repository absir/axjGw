package ANet

import (
	"io"
)

func Req(client Client, protocol Protocol, compress Compress, decrypt Encrypt, decryKey []byte, dataMax int32) (err error, req int32, uri string, uriI int32, data []byte) {
	err, bs, read := client.Read()
	if err != nil {
		return
	}

	var head byte
	if bs != nil {
		err, head, req, uri, uriI, data = protocol.Req(bs)

	} else if read != nil {
		err, head, req, uri, uriI, data = protocol.ReqReader(read, client.Sticky(), dataMax)

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

func Rep(client Client, buff *[]byte, protocol Protocol, compress Compress, compressMin int, encrypt Encrypt, encryKey []byte, req int32, uri string, uriI int32, data []byte, isolate bool) (err error) {
	if req < 0 {
		// 纯写入data
		return client.Write(data, false)
	}

	err, out, locker := client.Output()
	if err != nil {
		return err
	}

	var head byte = 0
	if data != nil {
		// 数据处理
		if compress != nil {
			bLen := len(data)
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

		if encrypt != nil && encryKey != nil {
			data, err = encrypt.Encrypt(data, encryKey, isolate)
			if err != nil {
				return err
			}

			head |= HEAD_ENCRY
		}
	}

	if out {
		return protocol.RepOut(locker, client, buff, req, uri, uriI, data, client.Sticky(), head)

	} else {
		return client.Write(protocol.Rep(req, uri, uriI, data, client.Sticky(), head), false)
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

func (r *RepBatch) Init(protocol Protocol, compress Compress, compressMin int, req int32, uri string, uriI int32, data []byte) error {
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

	r.data = data
	r.protocol = protocol
	r.repOrBH = func(sticky bool) []byte {
		if data == nil {
			return protocol.Rep(req, uri, uriI, data, sticky, head)

		} else {
			return protocol.RepBH(req, uri, uriI, true, head)
		}
	}

	r.bs = nil
	r.bss = nil
	r.bh = nil
	r.bhs = nil
	return nil
}

func (r *RepBatch) Rep(client Client, buff *[]byte, encrypt Encrypt, encryKey []byte) error {
	if r.data == nil {
		// 无数据写入
		if client.Sticky() {
			if r.bss == nil {
				r.bss = r.repOrBH(true)
			}

			return client.Write(r.bss, false)

		} else {
			if r.bs == nil {
				r.bs = r.repOrBH(false)
			}

			return client.Write(r.bs, false)
		}
	}

	// 有数据通用头
	var bh []byte
	if client.Sticky() {
		bh = r.bhs
		if bh == nil {
			bh = r.repOrBH(true)
			r.bhs = bh
		}

	} else {
		bh = r.bh
		if bh == nil {
			bh = r.repOrBH(false)
			r.bh = bh
		}
	}

	if encrypt == nil || encryKey == nil {
		// 无加密数据写入
		if client.Sticky() {
			if r.bss == nil {
				r.bss = r.protocol.RepBS(r.bhs, r.data, true, 0)
			}

			return client.Write(r.bss, false)

		} else {
			if r.bs == nil {
				r.bs = r.protocol.RepBS(r.bhs, r.data, false, 0)
			}

			return client.Write(r.bs, false)
		}
	}

	// 流写入检测
	err, out, locker := client.Output()
	if err != nil {
		return err
	}

	// 加密数据隔离
	var data []byte
	data, err = encrypt.Encrypt(r.data, encryKey, true)
	if err != nil {
		return err
	}

	head := HEAD_ENCRY
	if out {
		return r.protocol.RepOutBS(locker, client, buff, bh, data, client.Sticky(), head)

	} else {
		return client.Write(r.protocol.RepBS(bh, data, client.Sticky(), head), false)
	}
}

type Processor struct {
	Protocol    Protocol
	Compress    Compress
	CompressMin int
	Encrypt     Encrypt
	DataMax     int32
}
