@echo off
chcp 65001 >nul
echo ========================================
echo Mir2-Go 启动脚本
echo ========================================
setlocal

cd /d "%~dp0"

if not exist "bin" (
    echo 错误: 请先运行 init.bat 初始化项目
    pause
    exit /b 1
)

set CONFIG=%1
if "%CONFIG%"=="" set CONFIG=config.yaml

echo 使用配置文件: %CONFIG%
echo.

echo 启动 DBServer...
start "Mir2-DBServer" /b bin\dbsrv.exe -config %CONFIG%
timeout /t 1 /nobreak >nul

echo 启动 LoginSrv...
start "Mir2-LoginSrv" /b bin\loginsrv.exe -config %CONFIG%
timeout /t 1 /nobreak >nul

echo 启动 LoginGate...
start "Mir2-LoginGate" /b bin\logingate.exe -config %CONFIG%
timeout /t 1 /nobreak >nul

echo 启动 SelGate...
start "Mir2-SelGate" /b bin\selgate.exe -config %CONFIG%
timeout /t 1 /nobreak >nul

echo 启动 RunGate...
start "Mir2-RunGate" /b bin\rungate.exe -config %CONFIG%
timeout /t 1 /nobreak >nul

echo 启动 M2Server...
start "Mir2-M2Server" /b bin\m2server.exe -config %CONFIG%

echo.
echo ========================================
echo 所有服务器已启动!
echo ========================================
echo.
echo 停止服务器请运行 stop_servers.bat
echo.

endlocal
pause