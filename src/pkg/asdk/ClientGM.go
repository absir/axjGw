// +build gomobile

package asdk

import "axj/Kt/KtBuffer"

type Buffer interface{}

func BufferFree(buffer Buffer) {
	Util.PutBuffer(buffer.(*KtBuffer.Buffer))
}
