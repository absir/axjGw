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
	for i := range n.buckets {
		b := n.getBucket(uintptr(i))
		if !b.walkLock(f) {
			return false
		}
	}

	return true
}

func (that *CMap) RangeBuff(f func(k, v interface{}) bool, buff *[]interface{}, buffPSizeMax int) bool {
	n := that.getNode()
	for i := range n.buckets {
		b := n.getBucket(uintptr(i))
		if !b.walkBuff(f, buff, buffPSizeMax) {
			return false
		}
	}

	return true
}

func (that *CMap) RangeBuffs(f func(k, v interface{}) bool, pWait **Util.DoneWait, pBuffs *[][]interface{}, buffPSizeMax int) {
	if pBuffs == nil {
		that.RangeBuff(f, nil, buffPSizeMax)
		return
	}

	var wait *Util.DoneWait
	if pWait == nil {
		wait = Util.NewWaitDone(nil)

	} else {
		wait = *pWait
		if wait == nil {
			wait = Util.NewWaitDone(nil)
			*pWait = wait
		}
	}

	n := that.getNode()
	nLen := len(n.buckets)
	buffs := *pBuffs
	if buffs == nil || len(buffs) != nLen {
		buffs = make([][]interface{}, nLen)
		*pBuffs = buffs
	}

	for i := nLen - 1; i >= 0; i-- {
		b := n.getBucket(uintptr(i))
		if i == 0 {
			b.walkBuff(f, &buffs[i], buffPSizeMax)

		} else {
			go b.walkWait(wait, f, &buffs[i], buffPSizeMax)
			wait.Add()
		}
	}

	wait.Wait()
}

func (that *bucket) walkWait(wait *Util.DoneWait, f func(k, v interface{}) bool, buff *[]interface{}, buffPSizeMax int) {
	defer wait.Done()
	that.walkBuff(f, buff, buffPSizeMax)
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

func (that *bucket) walkBuff(f func(k, v interface{}) bool, pBuff *[]interface{}, buffPSizeMax int) bool {
	if pBuff == nil {
		return that.walk(f)
	}

	// buff
	buff := *pBuff
	var mLen2 int
	if buffPSizeMax < 1 {
		buffPSizeMax = 1
	}

	// 读锁，获取对象
	{
		that.mu.Lock()
		mLen2 = len(that.m) << 1
		if buff != nil {
			bLen := len(buff)
			if bLen < mLen2 {
				buff = nil

			} else if bLen > (mLen2<<1) && bLen > (mLen2+buffPSizeMax) {
				buff = nil
			}
		}

		if buff == nil {
			if mLen2 < buffPSizeMax {
				buff = make([]interface{}, mLen2<<1)

			} else {
				buff = make([]interface{}, mLen2+buffPSizeMax)
			}

			*pBuff = buff
		}

		i := 0
		for k, v := range that.m {
			buff[i] = k
			i++
			buff[i] = v
			i++
		}

		that.mu.Unlock()
	}

	// gc清理
	defer that.walkBuffClear(buff, mLen2)
	for i := 1; i < mLen2; i += 2 {
		if !f(buff[i-1], buff[i]) {
			return false
		}
	}

	return true
}
