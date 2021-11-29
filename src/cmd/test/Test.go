package main

import (
	"axj/ANet"
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"fmt"
	"time"
)

func main() {
	frameReader := ANet.FrameReader{}
	frameReader.Req = 2

	frame := frameReader.ReqFrame
	frame.Req = 1

	println(frameReader.Req)
	println(frame.Req)

	testMapRange()
}

func testMapRange() {
	rangFun := func(key, value interface{}) bool {
		async := value.(*Util.NotifierAsync)
		async.Start(func() {
			time.Sleep(30 * time.Second)
		})
		return true
	}

	for i := 0; i < 1000000; i++ {
		go func() {
			time.Sleep(60 * time.Second)
		}()
	}

	sb := 0
	cmap := cmap.NewCMapInit()
	for i := 0; i < 1000000; i++ {
		cmap.Store(i, Util.NewNotifierAsync(nil, nil, nil))
		if sb != cmap.SizeBuckets() {
			sb = cmap.SizeBuckets()
			fmt.Printf("%d => %d\n", cmap.CountFast(), cmap.SizeBuckets())
			{
				sTime := time.Now().UnixNano() / 10000000
				cmap.Range(rangFun)

				fmt.Printf("Range %d span %dms\n", cmap.CountFast(), time.Now().UnixNano()/10000000-sTime)
			}

			{
				var buff []interface{} = nil
				sTime := time.Now().UnixNano() / 10000000
				cmap.RangeBuff(rangFun, &buff, 16)

				fmt.Printf("RangeBuff %d span %dms\n", cmap.CountFast(), time.Now().UnixNano()/10000000-sTime)
			}

			{
				var buffs [][]interface{} = nil
				var wait *Util.DoneWait = nil
				sTime := time.Now().UnixNano() / 10000000
				cmap.RangeBuffs(rangFun, &buffs, 16, &wait)

				fmt.Printf("RangeBuffs %d span %dms\n", cmap.CountFast(), time.Now().UnixNano()/10000000-sTime)
			}
		}
	}
}
