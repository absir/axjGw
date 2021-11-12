package KtBytes

import (
	"axj/Kt/KtUnsafe"
	"fmt"
)

func Copy(bs []byte) []byte {
	if bs == nil {
		return nil
	}

	bLen := len(bs)
	if bLen <= 0 {
		return bs
	}

	dst := make([]byte, bLen)
	copy(dst, bs)
	return dst
}

func GetIntBytes(val int32) []byte {
	bytes := make([]byte, 4)
	SetInt(bytes, 0, val, nil)
	return bytes
}

func SetInt(bs []byte, off int, val int32, offP *int) {
	bs[off] = byte(val >> 24)
	off++
	bs[off] = byte(val >> 16)
	off++
	bs[off] = byte(val >> 8)
	off++
	bs[off] = byte(val)
	off++
	if offP != nil {
		*offP = off
	}
}

func GetInt(bs []byte, off int, offP *int) int32 {
	val := int32(bs[off]&0xFF) << 24
	off++
	val += int32(bs[off]&0xFF) << 16
	off++
	val += int32(bs[off]&0xFF) << 8
	off++
	val += int32(bs[off] & 0xFF)
	off++
	if offP != nil {
		*offP = off
	}

	return val
}

const (
	VINT_NB    byte = 0x80
	VINT_B          = VINT_NB - 1
	VINT            = int32(VINT_B)
	VINT_1_MAX      = VINT
	VINT_2_MAX      = VINT_1_MAX + (VINT << 7)
	VINT_3_MAX      = VINT_2_MAX + (VINT << 14)
	VINT_4_MAX      = VINT_3_MAX + (0XFF << 21)
)

func GetVIntLen(vInt int32) int32 {
	if vInt <= VINT_1_MAX {
		return 1
	}

	if vInt <= VINT_2_MAX {
		return 2
	}

	if vInt <= VINT_3_MAX {
		return 3
	}

	return 4
}

func GetVIntBytes(val int32) []byte {
	bytes := make([]byte, GetVIntLen(val))
	SetVInt(bytes, 0, val, nil)
	return bytes
}

func SetVInt(bs []byte, off int32, val int32, offP *int32) {
	if val > VINT_1_MAX {
		bs[off] = byte(val)&VINT_B | VINT_NB
		off++
		if val > VINT_2_MAX {
			bs[off] = (byte(val>>7)&VINT_B | VINT_NB)
			off++
			if val > VINT_3_MAX {
				if val > VINT_4_MAX {
					panic(fmt.Sprint("vInt err max %d, %d", VINT_4_MAX, val))

				} else {
					bs[off] = byte(val>>14)&VINT_B | VINT_NB
					off++
					bs[off] = byte(val >> 21)
					off++
				}

			} else {
				bs[off] = byte(val>>14) & VINT_B
				off++
			}

		} else {
			bs[off] = byte(val>>7) & VINT_B
			off++
		}

	} else {
		bs[off] = byte(val) & VINT_B
		off++
	}

	if offP != nil {
		*offP = off
	}
}

func GetVInt(bs []byte, off int32, offP *int32) int32 {
	b := bs[off]
	off++
	val := int32(b) & VINT
	if (b & VINT_NB) != 0 {
		b = bs[off]
		off++
		val += int32(b&VINT_B) << 7
		if (b & VINT_NB) != 0 {
			b = bs[off]
			off++
			val += int32(b&VINT_B) << 14
			if (b & VINT_NB) != 0 {
				b = bs[off]
				off++
				val += int32(b) << 21
			}
		}
	}

	if offP != nil {
		*offP = off
	}

	return val
}

func GetInt64(bs []byte, off int32, offP *int32) int64 {
	var val int64 = 0
	vf := 0
	for i := 0; i < 8; i++ {
		b := bs[off]
		off++
		val += int64(b) << vf
		vf += 8
	}

	if offP != nil {
		*offP = off
	}

	return val
}

func SetInt32(bs []byte, off int32, val int32, offP *int32) {
	for i := 0; i < 4; i++ {
		bs[off] = byte(val)
		off++
		val >>= 8
	}

	if offP != nil {
		*offP = off
	}
}

func SetInt64(bs []byte, off int32, val int64, offP *int32) {
	for i := 0; i < 8; i++ {
		bs[off] = byte(val)
		off++
		val >>= 8
	}

	if offP != nil {
		*offP = off
	}
}

func IndexByte(bs []byte, b byte, start int, end int) int {
	if start < 0 {
		start = 0
	}

	bLen := len(bs)
	if end < 0 || end > bLen {
		end = bLen
	}

	return KtUnsafe.IndexByte(bs, b, start, end)
}
