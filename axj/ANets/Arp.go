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
		PassDrt:  60,
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
}

type AddrIp struct {
	ip       string
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
	var addrIp *AddrIp = nil
	if val, _ := that.addrMap.Load(addr); val != nil {
		addrIp, _ = val.(*AddrIp)
		if addrIp != nil && addrIp.passTime > now {
			return addrIp.ip
		}
	}

	// 扫描开启
	if !that.scanning && (addrIp == nil || that.scanTime <= now) {
		Util.GoSubmit(that.scanRun)
	}

	if addrIp != nil {
		return addrIp.ip
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
				return addrIp.ip
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

	that.scanning = true
	table := arp.Table()
	now := time.Now().UnixNano()
	that.scanTime = now + that.PassDrt
	for ip, addr := range table {
		that.addrMap.Store(sAddr(addr), &AddrIp{
			ip:       ip,
			passTime: now + that.PassDrt,
		})
	}

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
