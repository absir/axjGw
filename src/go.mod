module axjGW

go 1.16

require (
	axj v0.0.0
	github.com/apache/thrift v0.14.2
	github.com/smallnest/rpcx v1.6.11
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	gw v0.0.0
)

replace (
	axj v0.0.0 => ./../axj
	gw v0.0.0 => ./../gen-go/gw
)
