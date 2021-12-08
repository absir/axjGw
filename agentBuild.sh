#!/usr/bin/env bash
cd `dirname $0`
args=$*

cd src/bin
export CGO_ENABLED=0

if [[ $args =~ "mips" ]];then
export GOOS=linux
export GOARCH=mipsle
export GOMIPS=softfloat
#export GOARM=5
go build -tags wsN -o ./agent-mips ../cmd/agent/Agent.go
fi

if [[ $args =~ "arm5" ]];then
export GOOS=linux
export GOARCH=arm
export GOARM=5
go build -tags wsN -o ./agent-arm5 ../cmd/agent/Agent.go
fi

if [[ $args =~ "arm7" ]];then
export GOOS=linux
export GOARCH=arm
export GOARM=5
go build -tags wsN -o ./agent-arm7 ../cmd/agent/Agent.go
fi

if [[ $args =~ "linux" ]];then
export GOOS=linux
export GOARCH=amd64
go build -tags wsN -o ./agent-linux ../cmd/agent/Agent.go
fi

if [[ $args =~ "win" ]];then
export GOOS=windows
export GOARCH=amd64
go build -tags wsN -o ./agent-win ../cmd/agent/Agent.go
fi