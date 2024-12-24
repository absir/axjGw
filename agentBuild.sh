#!/bin/sh
cd `dirname $0`
args=$*

mkdir -p src/bin

cd src/bin
CGO_ENABLED=0; export CGO_ENABLED
echo "args: $args"

case "$args" in
    *mips*)
        GOOS=linux; export GOOS
        GOARCH=mipsle; export GOARCH
        GOMIPS=softfloat; export GOMIPS
        go build -tags wsN,httpN -o ./agent-mips ../cmd/agent/Agent.go
        ;;
    *arm0*)
        GOOS=linux; export GOOS
        GOARCH=arm; export GOARCH
        go build -tags wsN,httpN -o ./agent-arm0 ../cmd/agent/Agent.go
        ;;
    *arm64*)
        GOOS=linux; export GOOS
        GOARCH=arm64; export GOARCH
        go build -tags wsN,httpN -o ./agent-arm64 ../cmd/agent/Agent.go
        ;;
    *arm5*)
        GOOS=linux; export GOOS
        GOARCH=arm; export GOARCH
        GOARM=5; export GOARM
        go build -tags wsN,httpN -o ./agent-arm5 ../cmd/agent/Agent.go
        ;;
    *arm7*)
        GOOS=linux; export GOOS
        GOARCH=arm; export GOARCH
        GOARM=7; export GOARM
        go build -tags wsN,httpN -o ./agent-arm7 ../cmd/agent/Agent.go
        ;;
    *linux*)
        GOOS=linux; export GOOS
        GOARCH=amd64; export GOARCH
        go build -tags wsN -o ./agent-linux ../cmd/agent/Agent.go
        ;;
    *win*)
        GOOS=windows; export GOOS
        GOARCH=amd64; export GOARCH
        go build -tags wsN -o ./agent-win.exe ../cmd/agent/Agent.go
        mkdir ../../bin/win
        rm -rf ../../bin/win/*
        cp -rf agent-win.exe ../../bin/win/
        cp -rf ../../install/agent.yml ../../bin/win/
        cp -rf ../../install/win/* ../../bin/win/
        ;;
esac