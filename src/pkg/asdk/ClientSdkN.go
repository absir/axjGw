//go:build !sdk
// +build !sdk

package asdk

import (
	"axj/Kt/KtBuffer"
	"axj/Thrd/Util"
)

type Buffer *KtBuffer.Buffer

func BufferFree(buffer Buffer) {
	Util.PutBuffer(buffer)
}
