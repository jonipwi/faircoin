@echo off
echo Building FairCoin Backend...

cd backend

echo Copying environment file...
if not exist .env (
    copy .env.example .env
    echo Created .env file. Please update the configuration as needed.
)

echo Installing Go dependencies...
go mod tidy

echo Building the application...
go build -o bin/faircoin.exe cmd/server/main.go

echo.
echo Build complete! To start the server:
echo   cd backend
echo   ./bin/faircoin.exe
echo.
echo Or use the development server:
echo   cd backend  
echo   go run cmd/server/main.go
echo.
echo Frontend is available at: http://localhost:8080
echo API documentation: http://localhost:8080/health
echo.

pause