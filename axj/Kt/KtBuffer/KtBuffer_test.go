package KtBuffer

import (
	"bytes"
	"fmt"
	"testing"
)

func TestSetRangeLen(t *testing.T) {
	type args struct {
		si   int
		ei   int
		rLen int
	}

	tests := []*args{
		{si: 2, ei: 6, rLen: 0},
		{si: 2, ei: 6, rLen: 2},
		{si: 2, ei: 6, rLen: 4},
		{si: 2, ei: 6, rLen: 6},
		{si: 2, ei: 6, rLen: 8},
		{si: 2, ei: 6, rLen: 10},
		{si: 2, ei: 6, rLen: 11},
	}

	buffer := new(bytes.Buffer)
	for _, test := range tests {
		buffer.Reset()
		for i := 0; i < 10; i++ {
			buffer.WriteByte(byte(i))
		}

		SetRangeLen(buffer, test.si, test.ei, test.rLen)
		println(fmt.Printf("%v", buffer.Bytes()))
	}
}
