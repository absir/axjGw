# Gw


## 虚拟人网关

### 部署:

```
#!/bin/bash
source /root/.profile
chmod +x wsBuildDeploy.sh
go env -w GOPROXY=https://goproxy.cn
cd src
go mod tidy
cd -
./wsBuildDeploy.sh 2,3
```

### 本地执行

```
cd src
go mod tidy
go run cmd/wserver/Wserver.go
```