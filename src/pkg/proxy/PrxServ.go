package proxy

import (
	"axj/ANet"
	"axj/APro"
	"axj/Thrd/AZap"
	"axj/Thrd/Util"
	"axjGW/pkg/agent"
	"bytes"
	"errors"
	"go.uber.org/zap"
	"net"
	"strconv"
)

type prxServ struct {
	Name  string
	Proto PrxProto
	cid   int64
	pAddr string
}

func StartServ(name string, addr string, proto PrxProto) *prxServ {
	val, _ := PrxServMng.servMap.Load(addr)
	if val != nil {
		that, _ := val.(*prxServ)
		return that
	}

	that := new(prxServ)
	that.Name = name
	that.Proto = proto
	serv, err := net.Listen("tcp", addr)
	if err != nil {
		AZap.Logger.Error("prxServ Start err", zap.Error(err))
	}

	AZap.Logger.Info("prxServ Start " + addr)
	Util.GoSubmit(func() {
		for !APro.Stopped {
			conn, err := serv.Accept()
			if err != nil {
				if APro.Stopped {
					return
				}

				AZap.Logger.Warn("prxServ Accept Err", zap.Error(err))
				continue
			}

			that.accept(conn.(*net.TCPConn))
		}
	})

	return that
}

func (that *prxServ) Update(proto PrxProto, cid int64, pAddr string) {
	if that.Proto != proto {
		AZap.Logger.Warn("prxServ Update err " + that.Proto.Name() + " => " + proto.Name())
		return
	}

	if cid > 0 && pAddr != "" {
		that.cid = cid
		that.pAddr = pAddr
		AZap.Logger.Info("prxServ Update " + strconv.FormatInt(cid, 10) + " => " + pAddr)
	}
}

func (that *prxServ) accept(conn *net.TCPConn) {
	buff := make([]byte, that.Proto.ReadBufferSize())
	buffer := &bytes.Buffer{}
	name := ""
	Util.GoSubmit(func() {
		for {
			size, err := conn.Read(buff)
			if err != nil || size <= 0 {
				PrxMng.closeConn(conn, err)
				return
			}

			ok := false
			ok, err = that.Proto.ReadServerName(buffer, buff[:size], &name)
			if err != nil {
				PrxMng.closeConn(conn, err)
				return
			}

			if ok {
				break
			}
		}

		// 解析代理连接成功
		client, pAddr := that.clientPAddr(name, that.Proto)
		if client == nil || pAddr == "" {
			PrxMng.closeConn(conn, errors.New("NO CLIENT RULE "+that.Name+" - "+name))
			return
		}

		id, adap := PrxMng.adapOpen(that.Proto, conn, buff)
		err := client.Get().Rep(true, agent.REQ_CONN, pAddr, id, buffer.Bytes(), false, false, 0)
		if err != nil {
			adap.Close(err)
			return
		}
	})
}

// 获取代理客户端，代理地址
func (that *prxServ) clientPAddr(name string, proto PrxProto) (ANet.Client, string) {
	if that.cid > 0 && that.pAddr != "" {
		// 指定代理
		client := PrxServMng.Manager.Client(that.cid)
		if client == nil {
			return nil, ""
		}

		return client, that.pAddr
	}

	return nil, ""
}
