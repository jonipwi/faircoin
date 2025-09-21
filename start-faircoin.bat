@echo off
cd /d "%~dp0backend"

echo Running database migration...
go run cmd/migrate-admin/main.go
echo.

echo Building FairCoin server...
go build -o bin/faircoin.exe cmd/server/main.go
if %ERRORLEVEL% EQU 0 (
    echo Build successful! Starting server...
    echo.
    echo Server will be available at:
    echo - Main site: http://localhost:8080/
    echo - Admin Dashboard: http://localhost:8080/admin  
    echo - Health Check: http://localhost:8080/health
    echo.
    echo Default Admin Login: admin / admin123
    echo.
    .\bin\faircoin.exe
) else (
    echo Build failed!
    pause
)