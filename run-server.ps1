cd backend
go mod tidy
go build -o ./bin/demo.exe ./cmd/demo/main.go
go build -o ./bin/faircoin.exe ./cmd/server/main.go

#./bin/faircoin.exe
#./bin/demo.exe
cd ..