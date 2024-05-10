package Util

import (
	"axj/Kt/Kt"
	"errors"
	"fmt"
	"sync"
	"time"
)

/*
* IdWorker
*
* 1                                               42           52             64
* +-----------------------------------------------+------------+---------------+
* | timestamp(ms)                                 | workerId   | sequence      |
* +-----------------------------------------------+------------+---------------+
* | 0000000000 0000000000 0000000000 0000000000 0 | 0000000000 | 0000000000 00 |
* +-----------------------------------------------+------------+---------------+
*
* 1. 41位时间截(毫秒级)，注意这是时间截的差值（当前时间截 - 开始时间截)。可以使用约70年: (1L << 41) / (1000L * 60 * 60 * 24 * 365) = 69
* 2. 10位数据机器位，可以部署在1024个节点
* 3. 12位序列，毫秒内的计数，同一机器，同一时间截并发4096个序号
 */
const (
	twepoch        = int64(1483228800)           //开始时间截 (2017-01-01)
	workerIdBits   = uint(10)                    //机器id所占的位数
	workerIdMax    = 1<<workerIdBits - 1         //支持的最大机器id数量
	worderIdMask   = workerIdMax                 //机器id掩码
	sequenceBits   = uint(12)                    //序列所占的位数
	sequenceMask   = int(1)<<sequenceBits - 1    //序号掩码
	workerIdShift  = sequenceBits                //机器id左移位数
	timestampShift = sequenceBits + workerIdBits //时间戳左移位数
)

// A IdWorker struct holds the basic information needed for a snowflake generator worker
type IdWorker struct {
	sync.Mutex
	timestamp int64
	workerId  int64
	sequence  int
}

// NewNode returns a new snowflake worker that can be used to generate snowflake IDs
func NewIdWorker(workerId int32) (*IdWorker, error) {
	if workerId < 0 || workerId > workerIdMax {
		return nil, errors.New(fmt.Sprintf("workerId must be between 0 and %d", workerIdMax))
	}

	return &IdWorker{
		timestamp: 0,
		workerId:  int64(workerId),
		sequence:  0,
	}, nil
}

func NewIdWorkerPanic(workerId int32) *IdWorker {
	idWorker, err := NewIdWorker(workerId)
	Kt.Panic(err)
	return idWorker
}

// Generate creates and returns a unique snowflake ID
func (that *IdWorker) Generate() int64 {
	return that.GenerateM(0, 0)
}

func (that *IdWorker) GenerateM(mod int, flg int) int64 {
	that.Lock()
	now := time.Now().UnixNano() / int64(time.Millisecond)
	if now > that.timestamp {
		// now时间增长后sequence回归
		that.sequence = flg

	} else {
		// 保证now时间不回滚
		now = that.timestamp
		sequence := that.sequence
		if mod > 1 {
			sequence = (((sequence+mod-1)/mod)*mod + flg) & sequenceMask

		} else {
			sequence = (sequence + 1) & sequenceMask
		}

		if that.sequence < sequence {
			// sequence增长
			that.sequence = sequence

		} else {
			// sequence无法增长,now++
			time.Sleep(time.Millisecond)
			now++
			that.sequence = flg
		}
	}

	that.timestamp = now
	id := (now-twepoch)<<timestampShift | (that.workerId << workerIdShift) | int64(that.sequence)
	that.Unlock()
	return id
}

func (that *IdWorker) MinId() int64 {
	return (1)<<timestampShift | (that.workerId << workerIdShift) | int64(that.sequence)
}

func (that *IdWorker) Timestamp(nanoTime int64) int64 {
	now := nanoTime / int64(time.Millisecond)
	return (now-twepoch)<<timestampShift | (that.workerId << workerIdShift)
}

func (that *IdWorker) GetWorkerId(id int64) int32 {
	return (int32)(id>>workerIdShift) & worderIdMask
}
