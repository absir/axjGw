package ANets

import (
	"axj/APro"
	"axj/Kt/KtUnsafe"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"github.com/mostlygeek/arp"
	"strings"
	"sync"
	"time"
)

type config struct {
	ScanDrt   int64         // 扫描间隔
	PassDrt   int64         // 过期间隔
	CheckDrt  time.Duration // 检查间隔
	UpdateDrt time.Duration // 更新间隔
	Timeout   time.Duration // 获取超时
	Debug     bool
	locker    sync.Locker
	addrMap   *sync.Map
	scanning  bool
	scanTime  int64
}

func (that *config) Write(p []byte) (n int, err error) {
	AZap.Error(KtUnsafe.BytesToString(p))
	n = len(p)
	return
}

var Config *config

func init() {
	Config = &config{
		ScanDrt:   600,
		PassDrt:   30,
		CheckDrt:  100,
		UpdateDrt: 300,
		Timeout:   5000,
		Debug:     true,
	}

	APro.SubCfgBind("arp", Config)
	Config.ScanDrt *= int64(time.Second)
	Config.PassDrt *= int64(time.Second)
	Config.CheckDrt *= time.Millisecond
	Config.UpdateDrt *= time.Millisecond
	Config.Timeout *= time.Millisecond
	Config.locker = new(sync.Mutex)
	Config.addrMap = new(sync.Map)
}

type AddrIp struct {
	ip       string
	scanTime int64
	passTime int64
}

func sAddr(addr string) string {
	addr = strings.ReplaceAll(addr, ":", "")
	addr = strings.ReplaceAll(addr, "-", "")
	addr = strings.ReplaceAll(addr, "_", "")
	return addr
}

func (that *config) FindIp(addr string, timeout time.Duration) string {
	addr = sAddr(addr)
	now := time.Now().UnixNano()
	var pTime int64 = 0
	if val, _ := that.addrMap.Load(addr); val != nil {
		addrIp, _ := val.(*AddrIp)
		if addrIp != nil {
			if addrIp.passTime > now {
				return addrIp.ip
			}

			if addrIp.scanTime > now {
				// 扫描开启
				that.scanStart()
				return addrIp.ip
			}

			pTime = addrIp.passTime
		}
	}

	// 扫描开启
	that.scanStart()

	// 获取超时
	if timeout <= 0 {
		timeout = that.Timeout
	}

	var addrIp *AddrIp = nil
	var wait time.Duration = 0
	for ; timeout > 0; timeout -= that.CheckDrt {
		time.Sleep(that.CheckDrt)
		wait += that.CheckDrt
		if val, _ := that.addrMap.Load(addr); val != nil {
			addrIp, _ = val.(*AddrIp)
			if addrIp != nil && (wait >= that.UpdateDrt || addrIp.passTime != pTime) {
				return addrIp.ip
			}
		}
	}

	if addrIp != nil {
		return addrIp.ip
	}

	return ""
}

func (that *config) scanStart() {
	if !that.scanning {
		Util.GoSubmit(that.scanRun)
	}
}

func (that *config) scanRun() {
	that.locker.Lock()
	if that.scanning {
		that.locker.Unlock()
		return
	}

	that.scanning = true
	table := arp.Table()
	now := time.Now().UnixNano()
	for ip, addr := range table {
		that.addrMap.Store(sAddr(addr), &AddrIp{
			ip:       ip,
			scanTime: now + that.ScanDrt,
			passTime: now + that.PassDrt,
		})
	}

	// 扫描过期清理
	clearPass := now - that.ScanDrt
	that.addrMap.Range(func(key, value interface{}) bool {
		addrIp, _ := value.(*AddrIp)
		if addrIp == nil || addrIp.passTime <= clearPass {
			that.addrMap.Delete(key)
		}

		return true
	})

	that.scanning = false
	that.locker.Unlock()
}
