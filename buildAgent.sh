cd src/bin
export CGO_ENABLED=0
export GOOS=linux
export GOARCH=arm

go build -tags wsN -o ./agent ../cmd/agent/Agent.go