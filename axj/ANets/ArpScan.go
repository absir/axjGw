package ANets

import (
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"go.uber.org/zap"
	"net"
	"sync"
	"time"
)

type ArpScan struct {
	locker    sync.Locker
	iface     *net.Interface
	rcvr      func(arp *layers.ARP)
	err       func(iface *net.Interface, err error)
	stopDrt   time.Duration
	addr      *net.IPNet
	handle    *pcap.Handle
	stop      chan struct{}
	startTime int64
}

func NewArpScan(iface *net.Interface, rcvr func(arp *layers.ARP), err func(iface *net.Interface, err error), stopDrt time.Duration) *ArpScan {
	that := new(ArpScan)
	that.locker = new(sync.Mutex)
	that.iface = iface
	that.rcvr = rcvr
	that.err = err
	that.stopDrt = stopDrt
	return that
}

func (that *ArpScan) Start() {
	var err error = nil
	that.locker.Lock()
	that.startTime = time.Now().UnixNano()
	err = that.scan(that.iface)
	that.locker.Unlock()
	if err != nil {
		if that.err == nil {
			AZap.Logger.Error("Scan Err", zap.Error(err))

		} else {
			that.err(that.iface, err)
		}
	}
}

func (that *ArpScan) Stop() {
	that.stopDo(true)
}

func (that *ArpScan) stopDo(locker bool) {
	if locker {
		that.locker.Lock()
	}

	if that.stop != nil {
		close(that.stop)
		that.handle.Close()
		that.stop = nil
		that.handle = nil
		that.addr = nil
	}

	if locker {
		that.locker.Unlock()
	}
}

func (that *ArpScan) stopDrtDo(handle *pcap.Handle) {
	for {
		time.Sleep(that.stopDrt)
		that.locker.Lock()
		if that.handle == handle {
			if that.startTime < time.Now().UnixNano()-int64(that.stopDrt) {
				that.stopDo(false)
				break
			}

		} else {
			break
		}

		that.locker.Unlock()
	}
}

// scan scans an individual interface's local network for machines using ARP requests/replies.
//
// scan loops forever, sending packets out regularly.  It returns an error if
// it's ever unable to write a packet.
func (that *ArpScan) scan(iface *net.Interface) error {
	// We just look for IPv4 addresses, so try to find if the interface has one.
	if that.stop == nil {
		var addr *net.IPNet
		if addrs, err := iface.Addrs(); err != nil {
			return err
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
			return errors.New("no good IP network found")
		} else if addr.IP[0] == 127 {
			return errors.New("skipping localhost")
		} else if addr.Mask[0] != 0xff || addr.Mask[1] != 0xff {
			return errors.New("mask means network is too large")
		}
		//log.Printf("Using network range %v for interface %v", addr, iface.Name)

		// Open up a pcap handle for packet reads/writes.
		handle, err := pcap.OpenLive(iface.Name, 65536, true, pcap.BlockForever)
		if err != nil {
			return err
		}

		that.addr = addr
		that.handle = handle
		that.stop = make(chan struct{})
		// Start up a goroutine to read in packet data.
		Util.GoSubmit(func() {
			that.readARP(that.handle, iface, that.stop)
		})

		// 自动关闭
		if that.stopDrt > 0 {
			Util.GoSubmit(func() {
				that.stopDrtDo(that.handle)
			})
		}
	}

	// Write our scan packets out to the handle.
	if err := that.writeARP(that.handle, iface, that.addr); err != nil {
		that.stopDo(false)
		if that.err == nil {
			AZap.Error("error writing packets on %v: %v", iface.Name, err)

		} else {
			that.err(iface, err)
		}

		return err
	}

	return nil
}

// readARP watches a handle for incoming ARP responses we might care about, and prints them.
//
// readARP loops until 'stop' is closed.
func (that *ArpScan) readARP(handle *pcap.Handle, iface *net.Interface, stop chan struct{}) {
	src := gopacket.NewPacketSource(handle, layers.LayerTypeEthernet)
	in := src.Packets()
	for {
		var packet gopacket.Packet
		select {
		case <-stop:
			return
		case packet = <-in:
			arpLayer := packet.Layer(layers.LayerTypeARP)
			if arpLayer == nil {
				continue
			}
			arp := arpLayer.(*layers.ARP)
			if arp.Operation != layers.ARPReply || bytes.Equal([]byte(iface.HardwareAddr), arp.SourceHwAddress) {
				// This is a packet I sent.
				continue
			}
			// Note:  we might get some packets here that aren't responses to ones we've sent,
			// if for example someone else sends US an ARP request.  Doesn't much matter, though...
			// all information is good information :)
			if that.rcvr == nil {
				AZap.Info("IP %v is at %v", net.IP(arp.SourceProtAddress), net.HardwareAddr(arp.SourceHwAddress))

			} else {
				that.rcvr(arp)
			}
		}
	}
}

// writeARP writes an ARP request for each address on our local network to the
// pcap handle.
func (that *ArpScan) writeARP(handle *pcap.Handle, iface *net.Interface, addr *net.IPNet) error {
	// Set up all the layers' fields we can.
	eth := layers.Ethernet{
		SrcMAC:       iface.HardwareAddr,
		DstMAC:       net.HardwareAddr{0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		EthernetType: layers.EthernetTypeARP,
	}
	arp := layers.ARP{
		AddrType:          layers.LinkTypeEthernet,
		Protocol:          layers.EthernetTypeIPv4,
		HwAddressSize:     6,
		ProtAddressSize:   4,
		Operation:         layers.ARPRequest,
		SourceHwAddress:   []byte(iface.HardwareAddr),
		SourceProtAddress: []byte(addr.IP),
		DstHwAddress:      []byte{0, 0, 0, 0, 0, 0},
	}
	// Set up buffer and options for serialization.
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}
	// Send one packet for every address.
	for _, ip := range that.ips(addr) {
		arp.DstProtAddress = []byte(ip)
		gopacket.SerializeLayers(buf, opts, &eth, &arp)
		if err := handle.WritePacketData(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

// ips is a simple and not very good method for getting all IPv4 addresses from a
// net.IPNet.  It returns all IPs it can over the channel it sends back, closing
// the channel when done.
func (that *ArpScan) ips(n *net.IPNet) (out []net.IP) {
	num := binary.BigEndian.Uint32([]byte(n.IP))
	mask := binary.BigEndian.Uint32([]byte(n.Mask))
	num &= mask
	for mask < 0xffffffff {
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], num)
		out = append(out, net.IP(buf[:]))
		mask++
		num++
	}
	return
}
