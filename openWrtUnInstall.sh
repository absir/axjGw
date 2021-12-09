#!/usr/bin/env bash
cliDir="/e/open/axj2/axj-cli"

cd `dirname $0`

if [[ $1 ==  *help ]];then
  echo 'args $port $ip $user $key'
  exit
fi

export sshBin="/bin/ash"
export crontabStart="true"

$cliDir/mnt/mng/uninstall.sh agent $1 $2 $3 $4