package KtJson

import (
	"axj/KtUnsafe"
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

	return KtUnsafe.BytesToString(b), err
}
