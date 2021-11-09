#!/usr/bin/env bash
cliDir="/e/open/axj/axj-cli"

cd `dirname $0`
args=$*

if [[ -z "$args" ]] || [[ $args =~ "2" ]];then
cd src/cmd/gateway
go build -o ./Gateway Gateway.go

if [[ $args =~ "3" ]];then
    rm -rf publish.tar.gz
    tar -zcvf publish.tar.gz Gateway
    $cliDir/mnt/mng/deployMng.sh axj-gateway publish.tar.gz /opt/mng dev-1
fi
fi