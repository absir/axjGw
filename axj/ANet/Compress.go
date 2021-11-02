package ANet

import (
	"bytes"
	"compress/gzip"
	"io"
)

// 数据压缩
type Compress interface {
	// 压缩
	Compress(data []byte) ([]byte, error)
	// 解压
	UnCompress(data []byte) ([]byte, error)
}

type CompressZip struct {
}

func (that *CompressZip) Compress(data []byte) ([]byte, error) {
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

func (that *CompressZip) UnCompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer reader.Close()
	return io.ReadAll(reader)
}
