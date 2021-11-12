package asdk

import "net"

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
	loginData(socket interface{}) []byte
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
	onState(socket interface{}, state int, data []byte)
	// 载入缓存，路由压缩字典
	loadStorage(name string) string
	// 保存缓存
	saveStorage(name string, value string)
}

type Client struct {
	addr   net.Addr
	socket *net.Conn
}

func NewClient(addr string) *Client {
	that := new(Client)

	return that
}
