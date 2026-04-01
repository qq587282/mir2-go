@echo off
echo Starting Mir2 Go Servers...
echo.

cd /d "%~dp0"

echo [1/6] Starting DBServer...
start "DBServer" bin\dbsrv.exe -config config.yaml
timeout /t 1 /nobreak >nul

echo [2/6] Starting LoginSrv...
start "LoginSrv" bin\loginsrv.exe -config config.yaml
timeout /t 1 /nobreak >nul

echo [3/6] Starting LoginGate...
start "LoginGate" bin\logingate.exe -config config.yaml
timeout /t 1 /nobreak >nul

echo [4/6] Starting SelServer...
start "SelServer" bin\selserver.exe -config config.yaml
timeout /t 1 /nobreak >nul

echo [5/6] Starting RunGate...
start "RunGate" bin\rungate.exe -config config.yaml
timeout /t 1 /nobreak >nul

echo [6/6] Starting M2Server...
start "M2Server" bin\m2server.exe -config config.yaml

echo.
echo All servers started!
echo.
echo Ports:
echo   DBServer   : 5400
echo   LoginSrv   : 15500
echo   LoginGate  : 17000
echo   SelServer  : 17100
echo   RunGate    : 17200
echo   M2Server   : 16000
echo.
pause
