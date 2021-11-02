package KtIo

import (
	KtBytes2 "axj/Kt/KtBytes"
	"io"
)

func GetVIntReader(reader io.ByteReader) int32 {
	var val int32 = 0
	b, err := reader.ReadByte()
	if err != nil {
		return val
	}

	val = int32(b) & KtBytes2.VINT
	if (b & KtBytes2.VINT_NB) != 0 {
		b, err = reader.ReadByte()
		if err != nil {
			return val
		}

		val += int32(b&KtBytes2.VINT_B) << 7
		if (b & KtBytes2.VINT_NB) != 0 {
			b, err = reader.ReadByte()
			if err != nil {
				return val
			}

			val += int32(b&KtBytes2.VINT_B) << 14
			if (b & KtBytes2.VINT_NB) != 0 {
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

func GetInt64Reader(reader io.ByteReader) int64 {
	var val int64 = 0
	vf := 0
	for i := 0; i < 8; i++ {
		b, err := reader.ReadByte()
		if err != nil {
			return val
		}

		val += int64(b) << vf
		vf += 8
	}

	return val
}

func ReadBytesReader(reader io.Reader, bLen int) ([]byte, error) {
	return readBytesReaderBsLen(reader, make([]byte, bLen), bLen)
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
