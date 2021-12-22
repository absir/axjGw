cd $(dirname $0)
currDir=$(pwd)

#运行内存限制
ulimit -v 20000

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