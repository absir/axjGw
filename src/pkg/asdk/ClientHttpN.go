//go:build httpN
// +build httpN

package asdk

import "axj/Kt/Kt"

func HttpAddr(url string, hash int) (string, error) {
	return "", Kt.NewErrReason("http no build")
}
