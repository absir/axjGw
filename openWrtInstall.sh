#!/usr/bin/env bash
cliDir="/e/open/axj2/axj-cli"

cd `dirname $0`

if [[ $1 ==  *help ]];then
  echo 'args $agent $port $ip $user $key'
  exit
fi

chmod +x $1

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

ssh $sPort $rUser@$rIp -i ~/.ssh/$rKey "$sshBin" << remotessh

chmod +x /opt/agent/agent
if [[ -f /etc/storage/started_script.sh ]];then
  mv /opt/agent/agent /etc/storage/agent/agent
  /etc/storage/agent/serv.sh start
  /sbin/mtd_storage.sh save
  exit
fi

/opt/agent/serv.sh start

exit
remotessh
