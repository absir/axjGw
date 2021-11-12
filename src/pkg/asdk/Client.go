package asdk

import (
	"axj/ANet"
	"axj/Kt/KtUnsafe"
)

const (
	CONN  = 0 // 开始连接
	OPEN  = 1 // 连接开启
	LOOP  = 2 // 可以通讯
	CLOSE = 3 // 连接关闭
	ERROR = 4 // 连接错误
	KICK  = 5 // 被剔
)

type Opt interface {
	// 授权数据
	loginData(adapter *Adapter) []byte
	// 推送数据处理 !uri && !data && fid 为 fid编号消息发送失败
	onPush(uri string, data []byte, fid int64)
	// 推送消息管道通知 gid 管道编号 connVer 推送消息时，连接版本，调用逻辑服务器Disc方法，附加验证 continues 为发送推送数据时，附加通知
	// 可以在附加消息逻辑 检测当前gid管道 是否监听， 不监听可调用逻辑服务器Disc方法， 防止之前调用逻辑服务器Disc可以未成功的情况
	onLast(gid string, connVer int32, continues bool)
	// 监听client连接状态编号
	/*
	   gw.state
	   state: {
	       CONN: 0, // 开始连接
	       OPEN: 1, // 连接开启
	       LOOP: 2, // 可以通讯
	       CLOSE: 3, // 连接关闭
	       ERROR: 4, // 连接错误
	       KICK: 5, // 被剔
	   },
	*/
	onState(adapter *Adapter, state int, data []byte)
	// 载入缓存，路由压缩字典
	loadStorage(name string) string
	// 保存缓存
	saveStorage(name string, value string)
}

type Client struct {
	opt        Opt
	addr       string
	adapter    *Adapter
	processor  *ANet.Processor
	uriMapUriI map[string]int32
	uriIMapUri map[int32]string
	uriMapHash string
}

type Adapter struct {
	conn     ANet.Conn
	decryKey []byte
}

func NewClient(addr string, opt Opt) *Client {
	that := new(Client)
	that.opt = opt
	that.addr = addr
	return that
}

func (that *Client) setUriMapUriI(uriMapJson string, uriMapHash string) {
	// uriMapUriI map[string]int32
	return
}

func (that *Client) onError(adapter *Adapter, err error) bool {
	if err == nil {
		return false
	}

	return true
}

func (that *Client) onRep(adapter *Adapter, repI int32, data []byte) {

}

func (that *Client) Conn() {
	if that.adapter == nil {
		//conn, err := net.Dial("tcp", that.addr)
		//if onError(nil, err) {
		//	return
		//}

	}
}

func (that *Client) loop(adapter *Adapter) {
	for adapter == that.adapter {
		err, req, uri, uriI, pid, data := that.processor.ReqPId(adapter.conn, adapter.decryKey)
		if that.onError(adapter, err) {
			break
		}

		println(pid)
		if req < ANet.REQ_ONEWAY {
			// 非返回值 才需要路由压缩解密
			if uri == "" && uriI > 0 {
				// 路由压缩解密
				uri = that.uriIMapUri[uriI]
			}
		}

		if req > ANet.REQ_ONEWAY {
			that.onRep(adapter, req, data)
		}

		switch req {
		case ANet.REQ_BEAT:
			break
		case ANet.REQ_KEY:
			adapter.decryKey = data
			break
		case ANet.REQ_ROUTE:
			that.setUriMapUriI(KtUnsafe.BytesToString(data), uri)
			break
		case ANet.REQ_KICK:
			that.opt.onState(adapter, KICK, data)
			break

		}
	}
}
