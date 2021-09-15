package anet

import (
	"axj/KtEncry"
	"axj/KtRand"
	"bytes"
	"compress/gzip"
	"io"
)

func Req() (client Client, protocol Protocol, compress Compress, encry Encry, decryKey []byte, err error, req int32, uri string, uriI int32, data []byte) {
	err, bs, read := client.Read()
	if err != nil {
		return
	}

	var head byte
	if bs != nil {
		err, head, req, uri, uriI, data = protocol.Req(bs)

	} else if read != nil {
		err, head, req, uri, uriI, data = protocol.ReqReader(read, client.Sticky())

	} else {
		err = io.EOF
		return
	}

	if err != nil {
		return
	}

	if data != nil {
		// 数据处理
		if (head&HEAD_ENCRY) != 0 && encry != nil && decryKey != nil {
			data, err = encry.Decry(data, decryKey)
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

func Rep(client Client, buff *[]byte, protocol Protocol, compress Compress, compressMin int, encry Encry, encryKey []byte, err error, req int32, uri string, uriI int32, data []byte) error {
	err, out, locker := client.Output()
	if err != nil {
		return err
	}

	var head byte = 0
	if data != nil {
		// 数据处理
		if compress != nil {
			bLen := len(data)
			if bLen > compressMin {
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

		if encry != nil && encryKey != nil {
			data, err = encry.Encry(data, encryKey)
			if err != nil {
				return err
			}
		}
	}

	if out {
		return protocol.RepClient(locker, client, buff, req, uri, uriI, data, client.Sticky(), head)

	} else {
		return client.Write(protocol.Rep(req, uri, uriI, data, client.Sticky(), head), false)
	}
}

type Processor struct {
	Protocol Protocol
	Compress Compress
	Encry    Encry
}

// 数据压缩
type Compress interface {
	// 压缩
	Compress(data []byte) ([]byte, error)
	// 解压
	UnCompress(data []byte) ([]byte, error)
}

// 数据加密
type Encry interface {
	// 生成密钥
	NewKeys() ([]byte, []byte)
	// 解密
	Decry(data []byte, key []byte) ([]byte, error)
	// 加密
	Encry(data []byte, key []byte) ([]byte, error)
}

type CompressZip struct {
}

func (c CompressZip) Compress(data []byte) ([]byte, error) {
	buffer := new(bytes.Buffer)
	writer := gzip.NewWriter(buffer)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (c CompressZip) UnCompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer reader.Close()
	return io.ReadAll(reader)
}

type EncrySr struct {
}

func (e EncrySr) NewKeys() ([]byte, []byte) {
	bs := KtRand.RandBytes(8)
	return bs, bs
}

func (e EncrySr) Decry(data []byte, key []byte) ([]byte, error) {
	KtEncry.SrDecry(data, key)
	return data, nil
}

func (e EncrySr) Encry(data []byte, key []byte) ([]byte, error) {
	return data, nil
}
