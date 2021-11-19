package proxy

import (
	"axj/Thrd/Util"
	"gitee.com/absir_admin/cmap"
	"sync"
)

type prxMng struct {
	locker   sync.Locker
	idWorker *Util.IdWorker
	connMap  *cmap.CMap
}

var PrxMng *prxMng
