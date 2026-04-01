#!/bin/bash

set -e

echo "Building Mir2-Go Server..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo "Creating bin directory..."
mkdir -p bin

build_server() {
    local name=$1
    local path=$2
    echo "Building $name..."
    go build -o "bin/${name}" "$path"
    if [ $? -ne 0 ]; then
        echo "Failed to build $name"
        exit 1
    fi
}

build_server "logingate" "./cmd/logingate"
build_server "selgate" "./cmd/selserver"
build_server "rungate" "./cmd/rungate"
build_server "m2server" "./cmd/m2server"
build_server "loginsrv" "./cmd/loginsrv"
build_server "dbsrv" "./cmd/dbsrv"

echo ""
echo "========================================"
echo "Build completed successfully!"
echo "========================================"
echo ""
echo "Binaries in bin/ directory:"
ls -la bin/
echo ""
