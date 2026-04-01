@echo off
echo Building Mir2-Go Server...

cd /d %~dp0

echo Creating bin directory...
if not exist bin mkdir bin

echo.
echo Building logingate...
go build -o bin/logingate.exe ./cmd/logingate
if errorlevel 1 goto error

echo.
echo Building selserver...
go build -o bin/selgate.exe ./cmd/selserver
if errorlevel 1 goto error

echo.
echo Building rungate...
go build -o bin/rungate.exe ./cmd/rungate
if errorlevel 1 goto error

echo.
echo Building m2server...
go build -o bin/m2server.exe ./cmd/m2server
if errorlevel 1 goto error

echo.
echo Building loginsrv...
go build -o bin/loginsrv.exe ./cmd/loginsrv
if errorlevel 1 goto error

echo.
echo Building dbsrv...
go build -o bin/dbsrv.exe ./cmd/dbsrv
if errorlevel 1 goto error

echo.
echo ========================================
echo Build completed successfully!
echo ========================================
echo.
echo Binaries in bin/ directory:
dir /b bin\*.exe
echo.
pause
exit /b 0

:error
echo.
echo ========================================
echo Build failed!
echo ========================================
pause
exit /b 1
