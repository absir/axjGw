#!/usr/bin/env bash
cliDir="/e/open/axj/axj-cli"

cd `dirname $0`
args=$*

if [[ $args =~ "mips" ]];then
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