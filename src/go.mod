module axjGW

go 1.16

require (
	axj v0.0.0
	github.com/apache/thrift v0.15.0
	github.com/hashicorp/golang-lru v0.5.4
	go.etcd.io/etcd/api/v3 v3.5.0
	go.etcd.io/etcd/client/v3 v3.5.0
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20210913180222-943fd674d43e
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	google.golang.org/grpc v1.42.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.15
)

replace axj v0.0.0 => ./../axj
