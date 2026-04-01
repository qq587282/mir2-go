#!/bin/bash

echo "========================================"
echo "Mir2-Go 停止脚本"
echo "========================================"

if [ -f .server_pids ]; then
    PIDS=$(cat .server_pids)
    
    for pid in $PIDS; do
        if kill -0 $pid 2>/dev/null; then
            echo "停止进程 $pid..."
            kill $pid 2>/dev/null
        fi
    done
    
    rm -f .server_pids
fi

echo ""
echo "所有服务器已停止!"
echo ""