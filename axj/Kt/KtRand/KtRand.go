package KtRand

import (
	"axj/Kt/KtBytes"
	"axj/Kt/KtUnsafe"
	"math/rand"
)

func RandBytes(bLen int) []byte {
	bs := make([]byte, bLen)
	sLen := 0
	if bLen > 4 {
		// 最小4个倍数
		sLen = (bLen >> 2) << 2
	}

	i := 0
	for i < sLen {
		KtBytes.SetInt(bs, i, rand.Int31(), &i)
	}

	for ; i < bLen; i++ {
		bs[i] = byte(rand.Int31())
	}

	return bs
}

func RandString(bLen int, chars []byte) string {
	cLen := int32(len(chars))
	bs := make([]byte, bLen)
	for i := 0; i < bLen; i++ {
		bs[i] = chars[rand.Int31n(cLen)]
	}

	return KtUnsafe.BytesToString(bs)
}
