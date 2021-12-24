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
	locker  sync.Locker
	iface   *net.Interface
	rcvr    func(scan *ArpScan, ip net.IP, addr net.HardwareAddr)
	err     func(reason string, iface *net.Interface, err error, ig bool)
	stopDrt time.Duration
	reqTime int64
	addr    *net.IPNet
	client  *arp.Client
	stop    chan struct{}
}

func NewArpScan(iface *net.Interface, rcvr func(scan *ArpScan, ip net.IP, addr net.HardwareAddr), err func(reason string, iface *net.Interface, err error, ig bool), stopDrt time.Duration) *ArpScan {
	that := new(ArpScan)
	that.locker = new(sync.Mutex)
	that.iface = iface
	that.rcvr = rcvr
	that.err = err
	that.stopDrt = stopDrt
	return that
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
	for {
		time.Sleep(that.stopDrt)
		if client != that.client {
			return
		}

		if that.reqTime <= time.Now().UnixNano()-int64(that.stopDrt) {
			break
		}
	}

	that.stopClient(client)
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

func (that *ArpScan) onErr(reason string, err error, ig bool) {
	if that.err == nil {
		iface := that.iface
		if err == nil {
			AZap.Warn(reason + "  (" + iface.HardwareAddr.String() + "." + iface.Name + ") ")

		} else {
			AZap.LoggerS.Error(reason+"  ("+iface.HardwareAddr.String()+"."+iface.Name+") ", zap.Error(err))
		}

	} else {
		that.err(reason, that.iface, err, ig)
	}
}

func (that *ArpScan) conn() *arp.Client {
	client := that.client
	if client == nil {
		var addr *net.IPNet
		if addrs, err := that.iface.Addrs(); err != nil {
			that.onErr("Addrs", err, false)
			return nil

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
			that.onErr("no good IP network found", nil, true)
			return nil

		} else if addr.IP[0] == 127 {
			that.onErr("skipping localhost", nil, true)
			return nil

		} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
			that.onErr(" mask means network is too large", nil, true)
			return nil
		}

		that.addr = addr
		var err error
		client, err = arp.Dial(that.iface)
		if err != nil {
			that.onErr("Dial", err, false)
			return nil
		}

		// 读取返回
		that.client = client
		Util.GoSubmit(func() {
			that.readARP(client)
		})

		// 自动关闭
		if that.stopDrt > 0 {
			Util.GoSubmit(func() {
				that.stopDrtDo(client)
			})
		}

		return client
	}

	return nil
}

func (that *ArpScan) connReq(locker bool) *arp.Client {
	if locker {
		that.locker.Lock()
	}

	client := that.conn()
	if client != nil {
		that.reqTime = time.Now().UnixNano()
	}

	if locker {
		that.locker.Unlock()
	}

	return client
}

func (that *ArpScan) readARP(client *arp.Client) {
	for {
		pack, _, err := client.Read()
		if err != nil {
			that.onErr("readARP", err, false)
			break
		}

		if pack == nil || pack.Operation != arp.OperationReply {
			continue
		}

		if that.rcvr == nil {
			AZap.Info("IP %v is at %v", pack.SenderIP, pack.SenderHardwareAddr)

		} else {
			that.rcvr(that, pack.SenderIP, pack.SenderHardwareAddr)
		}
	}

	that.stopClient(client)
}

func (that *ArpScan) ReqAll() {
	client := that.connReq(true)
	if client == nil {
		return
	}

	// 发送查询请求
	if err := that.writeARPAll(client, that.addr); err != nil {
		that.stopDo(false)
		that.onErr("ReqAll", err, false)
	}
}

func (that *ArpScan) writeARPAll(client *arp.Client, addr *net.IPNet) error {
	num := binary.BigEndian.Uint32(addr.IP)
	mask := binary.BigEndian.Uint32(addr.Mask)
	num &= mask
	var err error = nil
	for mask < 0xffffffff {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], num)
		err = client.Request(buf[:])
		if err != nil {
			break
		}

		mask++
		num++
	}

	return err
}

func (that *ArpScan) ReqIp(ip net.IP) {
	client := that.connReq(true)
	if client == nil {
		return
	}

	// 发送查询请求
	if err := client.Request(ip); err != nil {
		that.stopDo(false)
		that.onErr("ReqIp "+ip.String(), err, false)
	}
}
