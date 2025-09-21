#!/bin/bash

echo "Building FairCoin Backend..."

cd backend

echo "Copying environment file..."
if [ ! -f .env ]; then
    cp .env.example .env
    echo "Created .env file. Please update the configuration as needed."
fi

echo "Installing Go dependencies..."
go mod tidy

echo "Building the application..."
mkdir -p bin
go build -o bin/faircoin cmd/server/main.go

echo ""
echo "Build complete! To start the server:"
echo "  cd backend"
echo "  ./bin/faircoin"
echo ""
echo "Or use the development server:"
echo "  cd backend"
echo "  go run cmd/server/main.go"
echo ""
echo "Frontend is available at: http://localhost:8080"
echo "API documentation: http://localhost:8080/health"
echo ""