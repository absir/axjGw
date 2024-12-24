#!/bin/sh
cd `dirname $0`
args=$*

mkdir -p src/bin

cd src/bin
CGO_ENABLED=0; export CGO_ENABLED
echo "args: $args"

case "$args" in
    *linux*)
        GOOS=linux; export GOOS
        GOARCH=amd64; export GOARCH
        go build -tags wsN -o ./proxy-nps-linux ../cmd/proxy-nps/ProxyNps.go
        ;;
    *win*)
        GOOS=windows; export GOOS
        GOARCH=amd64; export GOARCH
        go build -tags wsN -o ./proxy-nps-win.exe ../cmd/proxy-nps/ProxyNps.go
        ;;
esac