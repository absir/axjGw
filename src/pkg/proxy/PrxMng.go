package proxy

import (
	"axj/Thrd/Util"
	"axj/Thrd/cmap"
	"sync"
)

type prxMng struct {
	locker   sync.Locker
	idWorker *Util.IdWorker
	connMap  *cmap.CMap
}

var PrxMng *prxMng
