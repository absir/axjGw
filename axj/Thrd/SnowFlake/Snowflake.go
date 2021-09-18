package SnowFlake

import (
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
	twepoch        = int64(1483228800000)             //开始时间截 (2017-01-01)
	workerIdBits   = uint(10)                         //机器id所占的位数
	sequenceBits   = uint(12)                         //序列所占的位数
	workerIdMax    = int32(-1 ^ (-1 << workerIdBits)) //支持的最大机器id数量
	sequenceMask   = int64(-1 ^ (-1 << sequenceBits)) //
	workerIdShift  = sequenceBits                     //机器id左移位数
	timestampShift = sequenceBits + workerIdBits      //时间戳左移位数
)

// A IdWorker struct holds the basic information needed for a snowflake generator worker
type IdWorker struct {
	sync.Mutex
	timestamp int64
	workerId  int64
	sequence  int64
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

// Generate creates and returns a unique snowflake ID
func (s *IdWorker) Generate() int64 {
	s.Lock()
	defer s.Unlock()

	now := time.Now().UnixNano() / 1000000
	if s.timestamp == now {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.timestamp {
				now = time.Now().UnixNano() / 1000000
			}
		}

	} else {
		s.sequence = 0
	}

	s.timestamp = now
	return (now-twepoch)<<timestampShift | (s.workerId << workerIdShift) | (s.sequence)
}

func GetWorkerId(id int64) int32 {
	return (int32)(id >> workerIdShift)
}
