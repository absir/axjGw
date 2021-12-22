package ANets

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"encoding/binary"
	"github.com/mdlayher/arp"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type ArpScan struct {
	locker    sync.Locker
	iface     *net.Interface
	rcvr      func(ip net.IP, addr net.HardwareAddr)
	err       func(reason string, iface *net.Interface, err error)
	startTime int64
	stopDrt   time.Duration
	addr      *net.IPNet
	client    *arp.Client
	stop      chan struct{}
}

func NewArpScan(iface *net.Interface, rcvr func(ip net.IP, addr net.HardwareAddr), err func(reason string, iface *net.Interface, err error), stopDrt time.Duration) *ArpScan {
	that := new(ArpScan)
	that.locker = new(sync.Mutex)
	that.iface = iface
	that.rcvr = rcvr
	that.err = err
	that.stopDrt = stopDrt
	return that
}

func (that *ArpScan) Start() {
	that.locker.Lock()
	that.startTime = time.Now().UnixNano()
	that.scan()
	that.locker.Unlock()
}

func (that *ArpScan) Stop() {
	that.stopDo(true)
}

func (that *ArpScan) stopDo(locker bool) {
	if locker {
		that.locker.Lock()
	}

	client := that.client
	if client != nil {
		that.addr = nil
		that.client = nil
		client.Close()
	}

	if locker {
		that.locker.Unlock()
	}
}

func (that *ArpScan) stopDrtDo(client *arp.Client) {
	var brk bool
	for {
		brk = true
		time.Sleep(that.stopDrt)
		that.locker.Lock()
		if client == that.client {
			if that.startTime > time.Now().UnixNano()-int64(that.stopDrt) {
				brk = false

			} else {
				that.stopDo(false)
			}
		}

		that.locker.Unlock()
		if brk {
			break
		}
	}
}

func (that *ArpScan) stopClient(client *arp.Client) {
	if client == that.client {
		that.locker.Lock()
		if client == that.client {
			that.stopDo(false)
		}

		that.locker.Unlock()
	}
}

func (that *ArpScan) onErr(reason string, err error) {
	if that.err == nil {
		iface := that.iface
		if err == nil {
			AZap.Warn(reason + "  (" + iface.HardwareAddr.String() + "." + iface.Name + ") ")

		} else {
			AZap.LoggerS.Error(reason+"  ("+iface.HardwareAddr.String()+"."+iface.Name+") ", zap.Error(err))
		}

	} else {
		that.err(reason, that.iface, err)
	}
}

func (that *ArpScan) scan() {
	if that.client == nil {
		var addr *net.IPNet
		if addrs, err := that.iface.Addrs(); err != nil {
			that.onErr("Addrs", err)
			return

		} else {
			for _, a := range addrs {
				if ipnet, ok := a.(*net.IPNet); ok {
					if ip4 := ipnet.IP.To4(); ip4 != nil {
						addr = &net.IPNet{
							IP:   ip4,
							Mask: ipnet.Mask[len(ipnet.Mask)-4:],
						}
						break
					}
				}
			}
		}

		// Sanity-check that the interface has a good address.
		if addr == nil {
			that.onErr("no good IP network found", nil)
			return

		} else if addr.IP[0] == 127 {
			that.onErr("skipping localhost", nil)
			return

		} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
			that.onErr("mask means network is too large", nil)
			return
		}

		that.addr = addr
		client, err := arp.Dial(that.iface)
		if err != nil {
			that.onErr("Dial", err)
			return
		}

		that.client = client
		// 读取返回
		Util.GoSubmit(func() {
			that.readARP(that.client)
		})

		// 自动关闭
		if that.stopDrt > 0 {
			Util.GoSubmit(func() {
				that.stopDrtDo(that.client)
			})
		}
	}

	// 发送查询请求
	if err := that.writeARP(that.client, that.addr); err != nil {
		that.stopDo(false)
		that.onErr("writeARP", err)
	}
}

func (that *ArpScan) readARP(client *arp.Client) {
	for {
		pack, _, err := client.Read()
		if err != nil {
			that.onErr("readARP", err)
			if client == that.client {

			}

			break
		}

		if pack == nil || pack.Operation != arp.OperationReply {
			continue
		}

		if that.rcvr == nil {
			AZap.Info("IP %v is at %v", pack.SenderIP, pack.SenderHardwareAddr)

		} else {
			that.rcvr(pack.SenderIP, pack.SenderHardwareAddr)
		}
	}
}

func (that *ArpScan) writeARP(client *arp.Client, addr *net.IPNet) error {
	num := binary.BigEndian.Uint32(addr.IP)
	mask := binary.BigEndian.Uint32(addr.Mask)
	num &= mask
	var err error = nil
	for mask < 0xffffffff {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], num)
		err = client.Request(net.IP(buf[:]))
		if err != nil {
			break
		}

		mask++
		num++
	}

	return err
}
