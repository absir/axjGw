#!/usr/bin/env bash
cliDir="/e/open/axj2/axj-cli"

cd `dirname $0`

if [[ $1 ==  *help ]];then
  echo 'args $port $ip $user $key'
  exit
fi

export sshBin="/bin/ash"

argi=1
source $cliDir/mnt/mas/_ssh.sh

ssh $sPort $rUser@$rIp -i ~/.ssh/$rKey "$sshBin" << remotessh
  /opt/agent/serv.sh stop
  rm -rf /etc/rc.d/S93agent
  rm -rf /etc/init.d/agent
remotessh