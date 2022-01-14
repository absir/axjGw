package KtJson

import (
	"axj/Kt/KtUnsafe"
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

func FromJson(json []byte, target interface{}) error {
	return jsoniter.Unmarshal(json, target)
}

func FromJsonStr(json string, target interface{}) error {
	return jsoniter.Unmarshal(KtUnsafe.StringToBytes(json), target)
}
