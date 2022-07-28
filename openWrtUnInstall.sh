#!/usr/bin/env bash
cliDir="/e/open/axj/axj-cli"

cd `dirname $0`

if [[ $1 ==  *help ]];then
  echo 'args $port $ip $user $key'
  exit
fi

export sshBin="/bin/ash"
export crontabStart="true"

argi=1
source $cliDir/mnt/mas/_ssh.sh

ssh $sPort $rUser@$rIp -i ~/.ssh/$rKey "$sshBin" << remotessh

/opt/agent/serv.sh stop

exit
remotessh

$cliDir/mnt/mng/uninstall.sh agent $1 $2 $3 $4