package KtEncry

import (
	"crypto/md5"
	"fmt"
)

// 加密
func SrEncry(bs []byte, keys []byte) {
	if bs == nil || keys == nil {
		return
	}

	bLen := len(bs)
	kLen := len(keys)
	if bLen <= 0 || kLen <= 0 {
		return
	}

	var b byte = 0
	var k = 0
	for i := 0; i < bLen; i++ {
		b ^= bs[i] ^ keys[k]
		bs[i] = b
		k++
		if k >= kLen {
			k = 0
		}
	}

	b = 0
	k = 0
	for i := bLen - 1; i >= 0; i-- {
		b ^= bs[i] ^ (keys[k] + byte(i))
		bs[i] = b
		k++
		if k >= kLen {
			k = 0
		}
	}
}

// 解密
func SrDecry(bs []byte, keys []byte) {
	if bs == nil || keys == nil {
		return
	}

	bLen := len(bs)
	kLen := len(keys)
	if bLen <= 0 || kLen <= 0 {
		return
	}

	var b byte = 0
	var k = 0
	var c byte
	for i := bLen - 1; i >= 0; i-- {
		b ^= keys[k] + byte(i)
		c = bs[i] ^ b
		bs[i] = c
		b ^= c
		k++
		if k >= kLen {
			k = 0
		}
	}

	b = 0
	k = 0
	for i := 0; i < bLen; i++ {
		b ^= keys[k]
		c = bs[i] ^ b
		bs[i] = c
		b ^= c
		k++
		if k >= kLen {
			k = 0
		}
	}
}

func EnMd5(bs []byte) string {
	mbs := md5.Sum(bs)
	return fmt.Sprintf("%x", mbs) //将[]byte转成16进制
}
