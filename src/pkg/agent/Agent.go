package agent

import (
	"axj/ANet"
	"time"
)

const (
	REQ_CONN     = ANet.REQ_ROUTE + 1
	REQ_RULES    = REQ_CONN + 1
	REQ_ON_RULE  = REQ_RULES + 1
	REQ_CLOSED   = REQ_ON_RULE + 1
	REQ_DIAL     = REQ_CLOSED + 1
	REQ_DIAL_ERR = REQ_DIAL + 1
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
	// 流量上报
	TrafficDrt time.Duration
	// 服务配置
	Cfg map[string]string
}
