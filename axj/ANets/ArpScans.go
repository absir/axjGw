package ANets

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type ArpScans struct {
	locker    sync.Locker
	ifaces    *sync.Map
	rcvr      func(scan *ArpScan, ip net.IP, addr net.HardwareAddr)
	err       func(reason string, iface *net.Interface, err error)
	stopDrt   time.Duration
	startTime int64
}

func NewArpScans(rcvr func(scan *ArpScan, ip net.IP, addr net.HardwareAddr), err func(reason string, iface *net.Interface, err error), stopDrt time.Duration) *ArpScans {
	that := new(ArpScans)
	that.locker = new(sync.Mutex)
	that.ifaces = new(sync.Map)
	that.rcvr = rcvr
	that.err = err
	that.stopDrt = stopDrt
	return that
}

func ifaceId(iface net.Interface) string {
	id := iface.HardwareAddr.String()
	addrs, _ := iface.Addrs()
	if addrs != nil {
		for _, addr := range addrs {
			id += addr.String()
		}
	}

	return id
}

func (that *ArpScans) ScanAll() {
	ifaces, err := net.Interfaces()
	if err != nil {
		AZap.Logger.Error("Interfaces Err", zap.Error(err))
		return
	}

	that.locker.Lock()
	now := time.Now().UnixNano()
	for _, iface := range ifaces {
		id := ifaceId(iface)
		if val, _ := that.ifaces.Load(id); val != nil {
			scan, _ := val.(*ArpScan)
			if scan != nil {
				if scan.reqTime < now {
					scan.reqTime = now
				}

				continue
			}
		}

		i := iface
		scan := NewArpScan(&i, that.rcvr, that.err, that.stopDrt)
		that.ifaces.Store(id, scan)
		if scan.reqTime < now {
			scan.reqTime = now
		}
	}

	// 启动、关闭、清理
	that.ifaces.Range(that.scanAllRange)
	that.locker.Unlock()
}

func (that *ArpScans) scanAllRange(key, value interface{}) bool {
	scan, _ := value.(*ArpScan)
	if scan == nil {
		that.ifaces.Delete(key)

	} else if scan.reqTime < that.startTime {
		Util.GoSubmit(scan.Stop)
		that.ifaces.Delete(key)

	} else {
		Util.GoSubmit(scan.ReqAll)
	}

	return true
}

func (that *ArpScans) Stop() {
	// 关闭
	that.startTime = int64(^uint64(0) >> 1)
	that.ifaces.Range(that.scanAllRange)
}

func (that *ArpScans) ScanFilter(locker bool, filter func(iface *net.Interface) bool, fun func(scan *ArpScan)) {
	ifaces, err := net.Interfaces()
	if err != nil {
		AZap.Logger.Error("Interfaces Err", zap.Error(err))
		return
	}

	if locker {
		that.locker.Lock()
		defer that.locker.Unlock()
	}

	for _, iface := range ifaces {
		i := iface
		if filter != nil && !filter(&i) {
			continue
		}

		var scan *ArpScan = nil
		id := ifaceId(iface)
		if val, _ := that.ifaces.Load(id); val != nil {
			scan, _ = val.(*ArpScan)
		}

		if scan == nil {
			scan = NewArpScan(&i, that.rcvr, that.err, that.stopDrt)
			that.ifaces.Store(id, scan)
		}

		fun(scan)
	}
}
