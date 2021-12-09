#!/usr/bin/env bash
cliDir="/e/open/axj2/axj-cli"

cd `dirname $0`

if [[ $1 ==  *help ]];then
  echo 'args $agent $port $ip $user $key'
  exit
fi

export bashBin="#!\/usr\/bin\/env ash"
export sshBin="/bin/ash"
# 运行内存限制10M
export servEnv="ulimit -v 20000"
# crontabStart
export crontabStart="*/10 * * * *"

$cliDir/mnt/mng/install.sh agent ./agent $2 $3 $4 $5

argi=2
source $cliDir/mnt/mas/_ssh.sh

scp $pPort -i ~/.ssh/$rKey -r $1 $rUser@$rIp:/opt/agent/agent