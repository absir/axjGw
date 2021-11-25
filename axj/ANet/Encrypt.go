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
	// 加密 内存申请复用优化
	EncryptLength(data []byte, key []byte) (int32, error)
	// 加密 内存申请复用优化
	EncryptToDest(data []byte, key []byte, dest []byte) error
}

type EncryptSr struct {
}

func (that *EncryptSr) NewKeys() ([]byte, []byte) {
	bs := KtRand.RandBytes(8)
	return bs, bs
}

func (that *EncryptSr) Decrypt(data []byte, key []byte) ([]byte, error) {
	KtEncry.SrDecry(data, key)
	return data, nil
}

func (that *EncryptSr) Encrypt(data []byte, key []byte, isolate bool) ([]byte, error) {
	if isolate {
		data = KtBytes.Copy(data)
	}

	KtEncry.SrEncry(data, key)
	return data, nil
}

func (that *EncryptSr) EncryptLength(data []byte, key []byte) (int32, error) {
	return int32(len(data)), nil
}

func (that *EncryptSr) EncryptToDest(data []byte, key []byte, dest []byte) error {
	copy(dest, data)
	KtEncry.SrEncry(dest, key)
	return nil
}
