package KtRand

import (
	"axj/KtBytes"
	"math/rand"
)

func RandBytes(bLen int) []byte {
	sLen := bLen % 4
	if sLen == 0 {
		sLen = bLen

	} else {
		sLen = bLen - sLen + 4
	}

	bs := make([]byte, sLen)
	for i := 0; i < sLen; i++ {
		KtBytes.SetInt(bs, i, rand.Int31(), &i)
	}

	return bs
}
