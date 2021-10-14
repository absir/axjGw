package cluster

import (
	"axj/Kt/Kt"
	"axj/Kt/KtStr"
	"github.com/apache/thrift/lib/go/thrift"
	"sync"
)

type Peer struct {
	Node
	socket *thrift.TSocket
}

func NewPeer(node Node) *Peer {
	peer := new(Peer)
	peer.Node = node
	peer.socket = nil
	return peer
}

func (that Peer) Client(pubs []KtStr.Matcher, conf *thrift.TConfiguration) *thrift.TSocket {
	if that.socket == nil {
		var err error = nil
		if KtStr.Matchers(pubs, that.addr, false) {
			that.socket, err = thrift.NewTSocketConf(that.addrPub, conf)

		} else {
			that.socket, err = thrift.NewTSocketConf(that.addr, conf)
		}

		Kt.Panic(err)
	}

	return that.socket
}

type Peers struct {
	Peers map[Id]Peer
	mutex sync.RWMutex
}

func NewPeers() *Peers {
	peers := new(Peers)
	peers.Peers = map[Id]Peer{}
	peers.mutex = sync.RWMutex{}
	return peers
}
