#!/usr/bin/env bash
cliDir="/e/open/axj/axj-cli"

cd `dirname $0`
args=$*

sed -i 's/const Version = .*/const Version = "'$(date +%g.%m%d.%H%M)'"/g' src/pkg/agent/Version.go

if [[ $args =~ "mips" ]];then
./agentBuild.sh mips
mkdir -p bin/mips
rm -rf bin/mips/*
cp src/bin/agent-mips bin/mips/agent
cp install/agent.yml bin/mips/
cp install/mipsRun.sh bin/mips/run.sh
cd bin
rm -rf mips.tar.gz
tar -zcvf mips.tar.gz -C mips .
$cliDir/mnt/mng/deployMng.sh agent-mips mips.tar.gz /opt/mng dev-1
fi

if [[ $args =~ "arm64" ]];then
./agentBuild.sh arm64
mkdir -p bin/arm64
rm -rf bin/arm64/*
cp src/bin/agent-arm64 bin/arm64/agent
cp install/agent.yml bin/arm64/
cp install/arm64Run.sh bin/arm64/run.sh
cd bin
rm -rf arm64.tar.gz
tar -zcvf arm64.tar.gz -C arm64 .
$cliDir/mnt/mng/deployMng.sh agent-arm64 arm64.tar.gz /opt/mng dev-1
fi