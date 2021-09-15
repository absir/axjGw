package cluster

type Id int16

type Node struct {
	// 节点地址
	addr string
	// 外部地址
	addrPub string
	// 服务属性
	servAtt string
}
