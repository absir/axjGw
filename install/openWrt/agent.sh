#!/bin/sh /etc/rc.common
# Copyright (C) 2006 OpenWrt.org
START=93
start() {
  /opt/agent/serv.sh start
}
stop() {
  /opt/agent/serv.sh stop
}
restart() {
  /opt/agent/serv.sh restart
}