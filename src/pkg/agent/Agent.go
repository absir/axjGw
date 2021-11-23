package agent

import "axj/ANet"

const (
	REQ_CERT  = ANet.REQ_ROUTE + 1
	REQ_CONN  = REQ_CERT + 1
	REQ_RULES = REQ_CONN + 1
)

type RULE struct {
	// 协议
	Proto string
	// 本地连接地址
	Addr string
	// 代理端口
	Port string
}
