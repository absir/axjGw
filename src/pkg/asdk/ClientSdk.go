//go:build sdk
// +build sdk

package asdk

import (
	"axj/Kt/KtBuffer"
	"axj/Thrd/Util"
)

type Buffer interface{}

func BufferFree(buffer Buffer) {
	if buffer == nil {
		return
	}

	Util.PutBuffer(buffer.(*KtBuffer.Buffer))
}
