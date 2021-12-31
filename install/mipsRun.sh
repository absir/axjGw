cd $(dirname $0)
currDir=$(pwd)

#运行内存限制 4g路由 ulimit -v ulimit -v can't resolve symbol ‘_ashldi3’
ulimitV=$(ulimit -v 20000 && echo "ok" | wc -l)
if [[ "$ulimitV" -ge 1 ]];then
	ulimit -v 20000
fi

#后台启动
chmod +x agent
if [[ $nohupN -ge 1 ]];then
  $currDir/agent > nohup.out &
else
  nohup $currDir/agent > nohup.out &
fi

#pid保存
pid=$!
echo "$pid" > pid.f