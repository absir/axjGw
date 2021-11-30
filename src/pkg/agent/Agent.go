package agent

import "axj/ANet"

const (
	REQ_CONN    = ANet.REQ_ROUTE + 1
	REQ_RULES   = REQ_CONN + 1
	REQ_ON_RULE = REQ_RULES + 1
	REQ_CLOSED  = REQ_ON_RULE + 1
)

type RULE struct {
	// 服务名
	Serv string
	// 协议
	Proto string
	// 代理端口
	Port string
	// 本地连接地址
	Addr string
	// 服务配置
	Cfg map[string]string
}
