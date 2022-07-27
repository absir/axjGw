//go:build httpN
// +build httpN

package asdk

import "axj/Kt/Kt"

func HttpAddr(url string) (string, error) {
	return "", Kt.NewErrReason("http no build")
}
