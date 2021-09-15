package KtIo

import (
	"axj/KtBytes"
	"io"
)

func GetVIntReader(reader io.ByteReader) int32 {
	var val int32 = 0
	b, err := reader.ReadByte()
	if err != nil {
		return val
	}

	val = int32(b) & KtBytes.VINT
	if (b & KtBytes.VINT_NB) != 0 {
		b, err = reader.ReadByte()
		if err != nil {
			return val
		}

		val += int32(b&KtBytes.VINT_B) << 7
		if (b & KtBytes.VINT_NB) != 0 {
			b, err = reader.ReadByte()
			if err != nil {
				return val
			}

			val += int32(b&KtBytes.VINT_B) << 14
			if (b & KtBytes.VINT_NB) != 0 {
				b, err = reader.ReadByte()
				if err != nil {
					return val
				}

				val += int32(b) << 21
			}
		}
	}

	return val
}

func ReadBytesReader(reader io.Reader, bLen int) ([]byte, error) {
	return readBytesReaderBsLen(reader, make([]byte, bLen), bLen)
}

func ReadBytesReaderBs(reader io.Reader, bs []byte) ([]byte, error) {
	return readBytesReaderBsLen(reader, bs, len(bs))
}

func readBytesReaderBsLen(reader io.Reader, bs []byte, bLen int) ([]byte, error) {
	var off int
	var err error
	for {
		off, err = reader.Read(bs)
		if err != nil {
			return nil, err
		}

		bLen -= off
		if bLen <= 0 {
			return bs, nil
		}

		bs = bs[off:]
	}
}
