package proxy

import (
	"axj/Thrd/Util"
	"github.com/lrita/cmap"
	"sync"
)

type prxMng struct {
	locker   sync.Locker
	idWorker *Util.IdWorker
	connMap  *cmap.Cmap
}

var PrxMng *prxMng
