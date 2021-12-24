package ANets

import (
	"axj/APro"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"go.uber.org/zap"
	"net"
	"strings"
	"sync"
	"time"
)

type config struct {
	ResetDrt  int64         // 重置间隔
	PassDrt   int64         // 过期间隔
	CheckDrt  time.Duration // 检查间隔
	Timeout   time.Duration // 获取超时
	Debug     bool
	locker    sync.Locker
	addrMap   *sync.Map
	scanning  bool
	scanTime  int64
	resetTime int64
	scans     *ArpScans
}

func (that *config) Write(p []byte) (n int, err error) {
	AZap.Error(KtUnsafe.BytesToString(p))
	n = len(p)
	return
}

var Config *config

func init() {
	Config = &config{
		ResetDrt: 600,
		PassDrt:  30,
		CheckDrt: 200,
		Timeout:  5000,
		Debug:    true,
	}

	APro.SubCfgBind("arp", Config)
	Config.ResetDrt *= int64(time.Second)
	Config.PassDrt *= int64(time.Second)
	Config.CheckDrt *= time.Millisecond
	Config.Timeout *= time.Millisecond
	Config.locker = new(sync.Mutex)
	Config.addrMap = new(sync.Map)
	Config.scans = NewArpScans(Config.scanRecr, Config.scanErr, time.Duration(Config.ResetDrt))
}

type AddrIp struct {
	ip        net.IP
	ipS       *string
	passTime  int64
	pass2Time int64
	scan      *ArpScan
}

func (that *AddrIp) IpStr() string {
	if that.ipS == nil {
		ipS := that.ip.String()
		that.ipS = &ipS
	}

	return *that.ipS
}

func (that *AddrIp) ReqIp() {
	if that.scan != nil {
		that.scan.ReqIp(that.ip)
	}
}

func sAddr(addr string) string {
	addr = strings.ReplaceAll(addr, ":", "")
	addr = strings.ReplaceAll(addr, "-", "")
	addr = strings.ReplaceAll(addr, "_", "")
	addr = strings.ToLower(addr)
	return addr
}

func (that *config) FindIp(addr string, timeout time.Duration) string {
	addr = sAddr(addr)
	now := time.Now().UnixNano()
	var addrIp *AddrIp = nil
	if val, _ := that.addrMap.Load(addr); val != nil {
		addrIp, _ = val.(*AddrIp)
		if addrIp != nil && addrIp.passTime > now {
			return addrIp.IpStr()
		}
	}

	if addrIp != nil && addrIp.pass2Time > now {
		// IP检查
		Util.GoSubmit(addrIp.ReqIp)

	} else if !that.scanning {
		// 扫描
		if addrIp == nil || that.scanTime <= now {
			Util.GoSubmit(that.scanRun)
		}
	}

	if addrIp != nil {
		return addrIp.IpStr()
	}

	// 获取超时
	if timeout <= 0 {
		timeout = that.Timeout
	}

	for ; timeout > 0; timeout -= that.CheckDrt {
		time.Sleep(that.CheckDrt)
		if val, _ := that.addrMap.Load(addr); val != nil {
			addrIp, _ = val.(*AddrIp)
			if addrIp != nil {
				return addrIp.IpStr()
			}
		}
	}

	return ""
}

func (that *config) scanRun() {
	that.locker.Lock()
	if that.scanning {
		that.locker.Unlock()
		return
	}

	// 扫描开启
	that.scanning = true
	that.scans.ScanAll()
	now := time.Now().UnixNano()
	that.scanTime = now + that.PassDrt

	// 重置清理
	that.resetTime = now - that.ResetDrt
	that.addrMap.Range(that.scanRange)

	// 扫描完成
	that.scanning = false
	that.locker.Unlock()
}

func (that *config) scanRange(key, value interface{}) bool {
	addrIp, _ := value.(*AddrIp)
	if addrIp == nil || addrIp.passTime <= that.resetTime {
		that.addrMap.Delete(key)
	}

	return true
}

func (that *config) scanRecr(scan *ArpScan, ip net.IP, addr net.HardwareAddr) {
	passTime := time.Now().UnixNano() + that.PassDrt
	that.addrMap.Store(sAddr(addr.String()), &AddrIp{
		ip:        ip,
		passTime:  passTime,
		pass2Time: passTime + that.PassDrt,
		scan:      scan,
	})
}

func (that *config) scanErr(reason string, iface *net.Interface, err error, ig bool) {
	if err == nil {
		if that.Debug {
			if ig {
				return
			}

			AZap.Warn("ScanAll Err " + reason + "  (" + iface.HardwareAddr.String() + "." + iface.Name + ") ")
		}

	} else {
		AZap.LoggerS.Error("ScanAll Err "+reason+"  ("+iface.HardwareAddr.String()+"."+iface.Name+") ", zap.Error(err))
	}
}
