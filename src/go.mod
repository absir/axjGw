module axjGW

go 1.16

require (
	axj v0.0.0
	gitee.com/absir_admin/cmap v1.0.7
	github.com/hashicorp/golang-lru v0.5.4
	github.com/json-iterator/go v1.1.12
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/robfig/cron/v3 v3.0.1
	go.uber.org/zap v1.19.1
	golang.org/x/net v0.0.0-20211118161319-6a13c67c3ce4
	golang.org/x/sys v0.0.0-20210809222454-d867a43fc93e // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/grpc v1.42.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.1.2
	gorm.io/gorm v1.21.15
)

replace axj v0.0.0 => ./../axj
