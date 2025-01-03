package cmap

import "axj/Thrd/Util"

func NewCMapInit() *CMap {
	m := &CMap{}
	n := m.getNode()
	n.initBuckets()
	return m
}

func (that *CMap) CountFast() int64 {
	return that.count
}

func (that *CMap) SizeBuckets() int {
	n := that.getNode()
	return len(n.buckets)
}

func (that *CMap) RangeLock(f func(k, v interface{}) bool) bool {
	n := that.getNode()
	nLen := len(n.buckets)
	for i := 0; i < nLen; i++ {
		b := n.getBucket(uintptr(i))
		if !b.walkLock(f) {
			return false
		}
	}

	return true
}

func (that *CMap) RangeBuff(f func(k, v interface{}) bool, pBuff *[]interface{}, buffPSizeMax int) bool {
	n := that.getNode()
	nLen := len(n.buckets)
	buffGcMax := 0
	for i := 0; i < nLen; i++ {
		b := n.getBucket(uintptr(i))
		if !b.walkBuff(f, pBuff, buffPSizeMax, &buffGcMax) {
			return false
		}
	}

	that.rangeBuffGc(pBuff, buffPSizeMax, buffGcMax)
	return true
}

func (that *CMap) rangeBuffGc(pBuff *[]interface{}, buffPSizeMax int, buffGcMax int) {
	if pBuff == nil {
		return
	}

	buff := *pBuff
	if buff == nil {
		return
	}

	bLen := len(*pBuff)
	if bLen > (buffGcMax<<1) && bLen > (buffGcMax+buffPSizeMax) {
		*pBuff = nil
	}
}

func (that *CMap) RangeBuffs(f func(k, v interface{}) bool, pBuffs *[][]interface{}, buffPSizeMax int, pWait **Util.DoneWait) {
	if pBuffs == nil {
		that.RangeBuff(f, nil, buffPSizeMax)
		return
	}

	n := that.getNode()
	nLen := len(n.buckets)
	buffs := *pBuffs
	if buffs != nil {
		bLen := len(buffs)
		if bLen < nLen {
			buffs = nil

		} else if bLen > (nLen << 1) {
			buffs = nil
		}
	}

	if buffs == nil {
		buffs = make([][]interface{}, nLen)
		*pBuffs = buffs
	}

	var wait *Util.DoneWait = nil
	for idx := nLen; idx > 0; idx -= mInitSize {
		if idx <= mInitSize {
			n.walkWait(nil, f, idx, buffs, buffPSizeMax)

		} else {
			if wait == nil {
				if pWait == nil {
					wait = Util.NewWaitDone(nil)

				} else {
					wait = *pWait
					if wait == nil {
						wait = Util.NewWaitDone(nil)
						*pWait = wait
					}
				}
			}

			go n.walkWait(wait, f, idx, buffs, buffPSizeMax)
			wait.Add()
		}
	}

	if wait != nil {
		wait.Wait()
	}
}

func (that *node) walkWait(wait *Util.DoneWait, f func(k, v interface{}) bool, idx int, buffs [][]interface{}, buffPSizeMax int) {
	if wait != nil {
		defer wait.Done()
	}

	for i := 0; i < mInitSize; i++ {
		idx--
		if idx < 0 {
			break
		}

		b := that.getBucket(uintptr(idx))
		b.walkBuff(f, &buffs[idx], buffPSizeMax, nil)
	}
}

func (that *bucket) walkLock(f func(k, v interface{}) bool) bool {
	that.mu.RLock()
	defer that.mu.RUnlock()
	for key, val := range that.m {
		if !f(key, val) {
			return false
		}
	}

	return true
}

func (that *bucket) walkBuffClear(buff []interface{}, mLen2 int) {
	for i := 0; i < mLen2; i++ {
		buff[i] = nil
	}
}

func (that *bucket) walkBuff(f func(k, v interface{}) bool, pBuff *[]interface{}, buffPSizeMax int, buffGcMax *int) bool {
	if pBuff == nil {
		return that.walk(f)
	}

	// buff
	buff := *pBuff
	// buffPSizeMax 不能小于1
	if buffPSizeMax < 1 {
		buffPSizeMax = 1
	}

	var kvLen int // key,val数 = len(that.m) * 2
	// 读锁，获取对象
	{
		that.mu.RLock()
		kvLen = len(that.m) << 1
		if buff != nil {
			bLen := len(buff)
			if bLen < kvLen {
				buff = nil

			} else {
				if buffGcMax == nil {
					if bLen > (kvLen<<1) && bLen > (kvLen+buffPSizeMax) {
						buff = nil
					}

				} else if *buffGcMax < kvLen {
					*buffGcMax = kvLen
				}
			}
		}

		if buff == nil {
			if kvLen > 0 && kvLen < buffPSizeMax {
				buff = make([]interface{}, kvLen<<1)

			} else {
				buff = make([]interface{}, kvLen+buffPSizeMax)
			}

			*pBuff = buff
		}

		if kvLen > 0 {
			i := 0
			for k, v := range that.m {
				buff[i] = k
				i++
				buff[i] = v
				i++
			}
		}

		that.mu.RUnlock()
	}

	// bucket.map为空值
	if kvLen <= 0 {
		return true
	}

	// 释放key,val,可以gc
	defer that.walkBuffClear(buff, kvLen)
	for i := 1; i < kvLen; i += 2 {
		if !f(buff[i-1], buff[i]) {
			return false
		}
	}

	return true
}
