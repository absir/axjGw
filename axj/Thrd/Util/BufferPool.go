package Util

import (
	"axj/Kt/KtBuffer"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"bytes"
	"sort"
	"sync"
)

type bufferPool struct {
	maxSize int
	pool    *sync.Pool
}

var bufferPools []*bufferPool

func newBuffer() interface{} {
	return new(bytes.Buffer)
}

func SetBufferPoolsStr(str string) {
	if str == "" {
		bufferPools = nil
	}

	strs := KtStr.SplitByte(str, ',', true, 0, 0)
	maxSizes := make([]int, len(strs))
	for i, s := range strs {
		maxSizes[i] = int(KtCvt.ToInt32(s))
	}

	SetBufferPools(maxSizes)
}

func SetBufferPools(maxSizes []int) {
	if maxSizes == nil || len(maxSizes) == 0 {
		bufferPools = nil
		return
	}

	sort.IntsAreSorted(maxSizes)
	pools := make([]*bufferPool, len(maxSizes))
	size := 0
	idx := 0
	for _, maxSize := range maxSizes {
		if maxSize <= size {
			AZap.Warn("SetBufferPools Err %d next %d", maxSize, size)
			continue
		}

		size = maxSize
		pools[idx] = &bufferPool{
			maxSize: maxSize,
			pool: &sync.Pool{
				New: newBuffer,
			},
		}

		idx++
	}

	if idx == 0 {
		bufferPools = nil
		return
	}

	if idx != len(maxSizes) {
		nPools := make([]*bufferPool, idx)
		for i, pool := range pools {
			nPools[i] = pool
		}

		pools = nPools
	}

	bufferPools = pools
}

func getBufferPool(size int) *bufferPool {
	pools := bufferPools
	if pools == nil {
		return nil
	}

	for _, pool := range pools {
		if size <= pool.maxSize {
			return pool
		}
	}

	return nil
}

func GetBuffer(size int, force bool) *bytes.Buffer {
	if size < 0 {
		size = 0
	}

	pool := getBufferPool(size)
	if pool == nil {
		if !force {
			return nil
		}

		buffer := new(bytes.Buffer)
		buffer.Grow(size)
		return buffer
	}

	buffer := pool.pool.Get().(*bytes.Buffer)
	buffer.Reset()
	buffer.Grow(size)
	return buffer
}

func PutBuffer(buffer *bytes.Buffer) {
	if buffer == nil {
		return
	}

	pool := getBufferPool(buffer.Cap())
	if pool != nil {
		pool.pool.Put(buffer)
	}
}

func GetBufferBytes(size int, pBuffer **bytes.Buffer) []byte {
	if size <= 0 {
		return KtBytes.EMPTY_BYTES
	}

	if pBuffer != nil {
		pool := getBufferPool(size)
		if pool != nil {
			buffer := pool.pool.Get().(*bytes.Buffer)
			buffer.Reset()
			KtBuffer.SetLen(buffer, size)
			*pBuffer = buffer
			return buffer.Bytes()
		}
	}

	return make([]byte, size)
}
