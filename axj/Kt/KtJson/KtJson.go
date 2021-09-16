package KtJson

import (
	KtUnsafe2 "axj/Kt/KtUnsafe"
	jsoniter "github.com/json-iterator/go"
)

func ToJson(obj interface{}) ([]byte, error) {
	return jsoniter.Marshal(obj)
}

func ToJsonStr(obj interface{}) (string, error) {
	b, err := ToJson(obj)
	if b == nil {
		return "", err
	}

	return KtUnsafe2.BytesToString(b), err
}
