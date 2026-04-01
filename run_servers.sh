#!/bin/bash

echo "========================================"
echo "Mir2-Go 启动脚本"
echo "========================================"

cd "$(dirname "$0")"

if [ ! -d "bin" ]; then
    echo "错误: 请先运行 init.sh 初始化项目"
    exit 1
fi

CONFIG=${1:-config.yaml}

echo "使用配置文件: $CONFIG"
echo ""

echo "启动 DBServer..."
./bin/dbsrv -config $CONFIG > logs/dbsrv.log 2>&1 &
DB_PID=$!
echo "  PID: $DB_PID"

sleep 1

echo "启动 LoginSrv..."
./bin/loginsrv -config $CONFIG > logs/loginsrv.log 2>&1 &
LOGIN_PID=$!
echo "  PID: $LOGIN_PID"

sleep 1

echo "启动 LoginGate..."
./bin/logingate -config $CONFIG > logs/logingate.log 2>&1 &
LG_PID=$!
echo "  PID: $LG_PID"

sleep 1

echo "启动 SelGate..."
./bin/selgate -config $CONFIG > logs/selgate.log 2>&1 &
SG_PID=$!
echo "  PID: $SG_PID"

sleep 1

echo "启动 RunGate..."
./bin/rungate -config $CONFIG > logs/rungate.log 2>&1 &
RG_PID=$!
echo "  PID: $RG_PID"

sleep 1

echo "启动 M2Server..."
./bin/m2server -config $CONFIG > logs/m2server.log 2>&1 &
M2_PID=$!
echo "  PID: $M2_PID"

echo ""
echo "========================================"
echo "所有服务器已启动!"
echo "========================================"
echo ""
echo "进程PID:"
echo "  DBServer:    $DB_PID"
echo "  LoginSrv:    $LOGIN_PID"
echo "  LoginGate:   $LG_PID"
echo "  SelGate:     $SG_PID"
echo "  RunGate:     $RG_PID"
echo "  M2Server:    $M2_PID"
echo ""
echo "查看日志: tail -f logs/*.log"
echo ""
echo "停止服务器: ./stop_servers.sh"

echo "$DB_PID $LOGIN_PID $LG_PID $SG_PID $RG_PID $M2_PID" > .server_pids