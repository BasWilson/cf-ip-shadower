clean:
	go clean

compile: 
	@make clean
	if [ ! -d "./bin" ]; then mkdir ./bin; fi
	@go build -o ./bin/shadow ./cmd/shadow/main.go

compile_linux: 
	@make clean
	if [ ! -d "./bin" ]; then mkdir ./bin; fi
	GOARCH=amd64 GOOS=linux go build -o ./bin/shadow-linux-amd64 ./cmd/shadow/main.go

compile_darwin: 	
	@make clean
	if [ ! -d "./bin" ]; then mkdir ./bin; fi
	GOARCH=arm64 GOOS=darwin go build -o ./bin/shadow-darwin-arm64 ./cmd/shadow/main.go

run: 
	@make clean
	@make compile
	./bin/shadow

dev:
	@go run cmd/shadow/main.go