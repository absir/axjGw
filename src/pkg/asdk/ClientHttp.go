//go:build !httpN
// +build !httpN

package asdk

import (
	"axj/Kt/Kt"
	"axj/Kt/KtStr"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
)

func HttpAddr(url string) (string, error) {
	rep, err := http.Get(url)
	if rep == nil || err != nil {
		return "", err
	}

	defer rep.Body.Close()
	body, err := ioutil.ReadAll(rep.Body)
	if body == nil || err != nil {
		return "", err
	}

	str := string(body)
	if !strings.HasPrefix(str, "addr:") {
		if len(str) > 32 {
			return "", Kt.NewErrReason("rep addr err, " + str[:32])

		} else {
			return "", Kt.NewErrReason("rep addr err, " + str)
		}

	}

	str = str[len("addr:"):]
	if strings.IndexByte(str, ',') < 0 {
		return str, nil
	}

	// 随机地址
	strs := KtStr.SplitByte(str, ',', true, 0, 0)
	return strs[rand.Int31n(int32(len(strs)))], nil
}
