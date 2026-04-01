@echo off
chcp 65001 >nul
echo ========================================
echo Mir2-Go 停止脚本
echo ========================================

echo 正在停止所有服务器...
taskkill /F /IM dbsrv.exe >nul 2>nul
taskkill /F /IM loginsrv.exe >nul 2>nul
taskkill /F /IM logingate.exe >nul 2>nul
taskkill /F /IM selgate.exe >nul 2>nul
taskkill /F /IM rungate.exe >nul 2>nul
taskkill /F /IM m2server.exe >nul 2>nul

echo.
echo 所有服务器已停止!
echo.

pause