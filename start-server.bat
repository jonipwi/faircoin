@echo off
cd /d "%~dp0backend"
echo Building FairCoin server...
go build -o bin/faircoin.exe cmd/server/main.go
if %ERRORLEVEL% EQU 0 (
    echo Build successful! Starting server...
    echo.
    echo Admin Dashboard will be available at: http://localhost:8080/admin
    echo Default Admin Login: admin / admin123
    echo.
    .\bin\faircoin.exe
) else (
    echo Build failed!
    pause
)