#!/bin/bash

echo "========================================"
echo "Mir2-Go 初始化脚本"
echo "========================================"

cd "$(dirname "$0")"

echo ""
echo "[1/4] 检查Go环境..."
if ! command -v go &> /dev/null; then
    echo "错误: 未安装Go语言"
    echo "请从 https://golang.org 下载安装"
    exit 1
fi

GO_VERSION=$(go version | grep -oP 'go\d+\.\d+')
echo "  Go版本: $GO_VERSION ✓"

echo ""
echo "[2/4] 下载依赖..."
go mod download
if [ $? -ne 0 ]; then
    echo "错误: 依赖下载失败"
    exit 1
fi
echo "  依赖下载完成 ✓"

echo ""
echo "[3/4] 编译项目..."
go build -o bin/logingate ./cmd/logingate
go build -o bin/selgate ./cmd/selserver
go build -o bin/rungate ./cmd/rungate
go build -o bin/m2server ./cmd/m2server
go build -o bin/loginsrv ./cmd/loginsrv
go build -o bin/dbsrv ./cmd/dbsrv
go build -o bin/testclient ./cmd/testclient

if [ $? -ne 0 ]; then
    echo "错误: 编译失败"
    exit 1
fi
echo "  编译完成 ✓"

echo ""
echo "[4/4] 检查配置文件..."
if [ ! -f config.yaml ]; then
    echo "  创建默认配置文件..."
    cat > config.yaml << 'EOF'
servername: "Mir2 Server"
serverip: "0.0.0.0"
serverport: 7000

logingate:
  enable: true
  ip: "0.0.0.0"
  port: 7000
  maxconn: 5000

selgate:
  enable: true
  ip: "0.0.0.0"
  port: 7100
  maxconn: 5000

rungate:
  enable: true
  ip: "0.0.0.0"
  port: 7200
  maxconn: 10000

m2server:
  enable: true
  ip: "0.0.0.0"
  port: 6000

loginsrv:
  enable: true
  ip: "0.0.0.0"
  port: 5500

dbsrv:
  enable: true
  type: "memory"
  ip: "127.0.0.1"
  port: 3306
  user: "root"
  password: ""
  database: "mir2"
EOF
    echo "  配置文件已创建 ✓"
else
    echo "  配置文件已存在 ✓"
fi

echo ""
echo "========================================"
echo "初始化完成!"
echo "========================================"
echo ""
echo "启动服务器:"
echo "  Windows: start_servers.bat"
echo "  Linux:   ./start_servers.sh"
echo ""
echo "或手动启动各服务器:"
echo "  ./bin/dbsrv -config config.yaml &"
echo "  ./bin/loginsrv -config config.yaml &"
echo "  ./bin/logingate -config config.yaml &"
echo "  ./bin/selgate -config config.yaml &"
echo "  ./bin/rungate -config config.yaml &"
echo "  ./bin/m2server -config config.yaml &"
echo ""