@echo off
chcp 65001 >nul
echo ========================================
echo Mir2-Go 初始化脚本
echo ========================================
setlocal

cd /d "%~dp0"

echo.
echo [1/4] 检查Go环境...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo 错误: 未安装Go语言
    echo 请从 https://golang.org 下载安装
    pause
    exit /b 1
)

for /f "tokens=2" %%i in ('go version') do set GO_VERSION=%%i
echo   Go版本: %GO_VERSION% ✓

echo.
echo [2/4] 下载依赖...
go mod download
if %errorlevel% neq 0 (
    echo 错误: 依赖下载失败
    pause
    exit /b 1
)
echo   依赖下载完成 ✓

echo.
echo [3/4] 编译项目...
if not exist bin mkdir bin

go build -o bin\logingate.exe .\cmd\logingate
go build -o bin\selgate.exe .\cmd\selserver
go build -o bin\rungate.exe .\cmd\rungate
go build -o bin\m2server.exe .\cmd\m2server
go build -o bin\loginsrv.exe .\cmd\loginsrv
go build -o bin\dbsrv.exe .\cmd\dbsrv
go build -o bin\testclient.exe .\cmd\testclient

if %errorlevel% neq 0 (
    echo 错误: 编译失败
    pause
    exit /b 1
)
echo   编译完成 ✓

echo.
echo [4/4] 检查配置文件...
if not exist config.yaml (
    echo   创建默认配置文件...
    (
        echo servername: "Mir2 Server"
        echo serverip: "0.0.0.0"
        echo serverport: 7000
        echo.
        echo logingate:
        echo   enable: true
        echo   ip: "0.0.0.0"
        echo   port: 7000
        echo   maxconn: 5000
        echo.
        echo selgate:
        echo   enable: true
        echo   ip: "0.0.0.0"
        echo   port: 7100
        echo   maxconn: 5000
        echo.
        echo rungate:
        echo   enable: true
        echo   ip: "0.0.0.0"
        echo   port: 7200
        echo   maxconn: 10000
        echo.
        echo m2server:
        echo   enable: true
        echo   ip: "0.0.0.0"
        echo   port: 6000
        echo.
        echo loginsrv:
        echo   enable: true
        echo   ip: "0.0.0.0"
        echo   port: 5500
        echo.
        echo dbsrv:
        echo   enable: true
        echo   type: "memory"
        echo   ip: "127.0.0.1"
        echo   port: 3306
        echo   user: "root"
        echo   password: ""
        echo   database: "mir2"
    ) > config.yaml
    echo   配置文件已创建 ✓
) else (
    echo   配置文件已存在 ✓
)

echo.
echo ========================================
echo 初始化完成!
echo ========================================
echo.
echo 启动服务器:
echo   run_servers.bat
echo.
echo 或手动启动各服务器:
echo   bin\dbsrv.exe -config config.yaml
echo   bin\loginsrv.exe -config config.yaml
echo   bin\logingate.exe -config config.yaml
echo   bin\selgate.exe -config config.yaml
echo   bin\rungate.exe -config config.yaml
echo   bin\m2server.exe -config config.yaml
echo.

endlocal
pause