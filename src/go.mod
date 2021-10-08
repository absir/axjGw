module axjGW

go 1.16

require (
	axj v0.0.0
	github.com/apache/thrift v0.14.2
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20210913180222-943fd674d43e
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.15
	gw v0.0.0
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
)

replace (
	axj v0.0.0 => ./../axj
	gw v0.0.0 => ./../gen-go/gw
)
