.DEFAULT_GOAL := run

build-mac:
	GOOS=darwin go build -o ssh-tail-mac main.go

build-linux:
	GOOS=linux go build -o ssh-tail-linux main.go

build-windows:
	GOOS=windows go build -o ssh-tail.exe main.go

build: build-mac build-linux build-windows

run:
	go run main.go
