#!/usr/bin/env bash
cliDir="/e/open/axj/axj-cli"

cd `dirname $0`
args=$*

if [[ -z "$args" ]] || [[ $args =~ "2" ]];then
  mkdir -p src/bin
  cd src/bin
  export CGO_ENABLED=0
  export GOOS=linux
  export GOARCH=amd64
  go build -o ./ws ../cmd/wserver/WServer.go
  rm -rf publish.tar.gz
  tar -zcvf publish.tar.gz ws

if [[ $args =~ "3" ]];then
    $cliDir/mnt/mng/deployMng.sh axj-ws publish.tar.gz /opt/mng dev-1
fi
fi