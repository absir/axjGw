package proxy

import "sync"

type prxMng struct {
	locker sync.Locker
}

var PrxMng *prxMng
