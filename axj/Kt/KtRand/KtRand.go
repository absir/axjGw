package KtRand

import (
	KtBytes2 "axj/Kt/KtBytes"
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
	for ; i < sLen; {
		KtBytes2.SetInt(bs, i, rand.Int31(), &i)
	}

	for ; i < bLen; i++ {
		bs[i] = byte(rand.Int31())
	}

	return bs
}
