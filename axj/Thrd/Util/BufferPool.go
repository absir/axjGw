package Util

import (
	"axj/Kt/KtBuffer"
	"axj/Kt/KtBytes"
	"axj/Kt/KtCvt"
	"axj/Kt/KtStr"
	"axj/Thrd/AZap"
	"sort"
	"sync"
)

type bufferPool struct {
	maxSize int
	pool    *sync.Pool
}

var bufferPools []*bufferPool

func newBuffer() interface{} {
	return new(KtBuffer.Buffer)
}

func SetBufferPoolsS(maxSizes string) {
	if maxSizes == "" {
		bufferPools = nil
	}

	ss := KtStr.SplitByte(maxSizes, ',', true, 0, 0)
	maxs := make([]int, len(ss))
	for i, s := range ss {
		maxs[i] = int(KtCvt.ToInt32(s))
	}

	SetBufferPools(maxs)
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

func GetBuffer(size int, force bool) *KtBuffer.Buffer {
	if size < 0 {
		size = 0
	}

	pool := getBufferPool(size)
	if pool == nil {
		if !force {
			return nil
		}

		buffer := new(KtBuffer.Buffer)
		KtBuffer.SetGetBytesSize(buffer, size, size)
		buffer.Reset()
		return buffer
	}

	buffer := pool.pool.Get().(*KtBuffer.Buffer)
	KtBuffer.SetGetBytesSize(buffer, size, size)
	buffer.Reset()
	return buffer
}

func PutBuffer(buffer *KtBuffer.Buffer) {
	if buffer == nil {
		return
	}

	pool := getBufferPool(buffer.Cap())
	if pool != nil {
		pool.pool.Put(buffer)
	}
}

func GetBufferBytes(size int, pBuffer **KtBuffer.Buffer) []byte {
	if size <= 0 {
		return KtBytes.EMPTY_BYTES
	}

	if pBuffer != nil {
		buffer := *pBuffer
		if buffer == nil {
			pool := getBufferPool(size)
			if pool != nil {
				buffer = pool.pool.Get().(*KtBuffer.Buffer)
				bs := KtBuffer.SetGetBytesSize(buffer, size, pool.maxSize)
				*pBuffer = buffer
				return bs
			}

		} else {
			return KtBuffer.SetGetBytesSize(buffer, size, size)
		}
	}

	return make([]byte, size)
}
