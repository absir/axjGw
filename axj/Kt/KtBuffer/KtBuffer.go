package KtBuffer

import "bytes"

func IndexByte(b []byte, c byte) int {
	return bytes.IndexByte(b, c)
}

type IBuffer interface {
	Len() int
	WriteByte(c byte) error
	Write(p []byte) (n int, err error)
	Bytes() []byte
	Grow(n int)
	Truncate(n int)
}

func SetLen(buffer IBuffer, len int) {
	bLen := buffer.Len()
	if bLen < len {
		buffer.Grow(len)
		for bLen < len {
			if bLen <= 0 {
				buffer.WriteByte(0)

			} else {
				gLen := len - bLen
				if bLen > gLen {
					buffer.Write(buffer.Bytes()[:gLen])

				} else {
					buffer.Write(buffer.Bytes())
				}
			}

			bLen = buffer.Len()
		}

	} else if bLen > len {
		buffer.Truncate(len)
	}
}

func SetRangeLen(buffer IBuffer, si int, ei int, rLen int) bool {
	sLen := ei - si
	if sLen < 0 {
		return false

	} else if sLen == rLen {
		return true
	}

	bs := buffer.Bytes()
	SetLen(buffer, buffer.Len()-sLen+rLen)
	if sLen > rLen {
		// 可以copy数据移动
		copy(bs[si+rLen:], bs[ei:])

	} else {
		// 按位移动
		bLen := len(bs)
		nBs := buffer.Bytes()
		nLen := len(nBs)
		bE := bLen - ei
		for i := 1; i <= bE; i++ {
			nBs[nLen-i] = bs[bLen-i]
		}
	}

	return true
}

func SetGetBytesSize(buffer *Buffer, size int, max int) []byte {
	buf := buffer.buf
	bLen := cap(buf)
	if bLen >= size {
		return buf[:size]
	}

	mLen := size + smallBufferSize
	tLen := bLen<<1 + smallBufferSize
	if tLen < mLen {
		tLen = mLen
	}

	if tLen > max {
		tLen = max
	}

	buf = make([]byte, tLen)
	buffer.buf = buf
	return buf[:size]
}
