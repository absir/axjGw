cd $(dirname $0)
currDir=$(pwd)

#运行内存限制
export servEnv="ulimit -v 20000"

#后台启动
nohup $currDir/agent &

#pid保存
pid=$!
echo "$pid" > pid.f