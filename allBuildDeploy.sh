#!/usr/bin/env bash
cliDir="/e/open/axj/axj-cli"

cd `dirname $0`
args=$*

if [[ -z "$args" ]] || [[ $args =~ "2" ]];then
cd src/cmd/gateway
go build -o ./gateway Gateway.go

if [[ $args =~ "3" ]];then
    rm -rf publish.tar.gz
    tar -zcvf publish.tar.gz gateway
    $cliDir/mnt/mng/deployMng.sh axj-gw publish.tar.gz /opt/mng dev-1
fi
fi

if [[ -z "$args" ]] || [[ $args =~ "5" ]];then
  mkdir -p src/bin
  cd src/bin
  export CGO_ENABLED=0
  export GOOS=linux
  export GOARCH=amd64
  go build -o ./gateway ../cmd/gateway/Gateway.go
  rm -rf publish.tar.gz
  tar -zcvf publish.tar.gz gateway
  $cliDir/mnt/mng/deployMng.sh axj-gw-dev publish.tar.gz /opt/mng dev-1
fi