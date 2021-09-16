package ANet

import (
	"axj/Kt/KtBytes"
	"axj/Kt/KtEncry"
	"axj/Kt/KtRand"
)

// 数据加密
type Encrypt interface {
	// 生成密钥
	NewKeys() ([]byte, []byte)
	// 解密
	Decrypt(data []byte, key []byte) ([]byte, error)
	// 加密 isolate 安全隔离
	Encrypt(data []byte, key []byte, isolate bool) ([]byte, error)
}

type EncryptSr struct {
}

func (e EncryptSr) NewKeys() ([]byte, []byte) {
	bs := KtRand.RandBytes(8)
	return bs, bs
}

func (e EncryptSr) Decrypt(data []byte, key []byte) ([]byte, error) {
	KtEncry.SrDecry(data, key)
	return data, nil
}

func (e EncryptSr) Encrypt(data []byte, key []byte, isolate bool) ([]byte, error) {
	if isolate {
		data = KtBytes.Copy(data)
	}

	return data, nil
}
