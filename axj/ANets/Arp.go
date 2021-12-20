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
	ScanDrt  int64         // arp重启
	DelayDrt time.Duration // arp扫描间隔
	PassDrt  int64         // arp缓存过期
	CheckDrt time.Duration // 检查时间
	Timeout  time.Duration // 获取超时
	Debug    bool
	locker   sync.Locker
	addrMap  *sync.Map
	scanning bool
	scanPass int64
}

func (that *config) Write(p []byte) (n int, err error) {
	AZap.Error(KtUnsafe.BytesToString(p))
	n = len(p)
	return
}

var Config *config

func init() {
	Config = &config{
		ScanDrt:  600,
		DelayDrt: 60,
		PassDrt:  30,
		CheckDrt: 100,
		Timeout:  200,
		Debug:    true,
	}

	APro.SubCfgBind("arp", Config)
	Config.ScanDrt *= int64(time.Second)
	Config.DelayDrt *= time.Second
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
	if val, _ := that.addrMap.Load(addr); val != nil {
		addrIp, _ := val.(*AddrIp)
		if addrIp != nil && addrIp.passTime > now {
			return addrIp.ip
		}
	}

	if that.scanPass <= now {
		that.scanStop()
		that.scanPass = now + that.ScanDrt
	}

	if !that.scanning {
		Util.GoSubmit(that.scanLoop)
	}

	if timeout <= 0 {
		timeout = that.Timeout
	}

	var addrIp *AddrIp = nil
	for ; timeout > 0; timeout -= that.CheckDrt {
		time.Sleep(that.CheckDrt)
		if val, _ := that.addrMap.Load(addr); val != nil {
			addrIp, _ = val.(*AddrIp)
			if addrIp != nil && (addrIp.passTime > now || that.PassDrt <= 0) {
				return addrIp.ip
			}
		}
	}

	if addrIp != nil {
		return addrIp.ip
	}

	return ""
}

func (that *config) scanStop() {
	that.locker.Lock()
	that.scanning = false
	that.locker.Unlock()
}

func (that *config) scanLoop() {
	that.locker.Lock()
	if that.scanning {
		that.locker.Unlock()
		return
	}

	that.addrMap = new(sync.Map)
	for ip, addr := range arp.Table() {
		that.addrMap.Store(sAddr(addr), &AddrIp{
			ip:       ip,
			passTime: time.Now().UnixNano() + that.PassDrt,
		})
	}

	that.scanning = true
	that.locker.Unlock()
}
