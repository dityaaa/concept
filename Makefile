BINARY_NAME ?= concept
VERSION ?= $(shell git describe --tags 2>/dev/null | cut -c 2-)
OUTPUT_DIR ?= ./build

build:
	mkdir ${OUTPUT_DIR}
	GOARCH=amd64 GOOS=linux go build -o ./build/${BINARY_NAME}-linux-amd64 ./cli/main.go
	GOARCH=amd64 GOOS=windows go build -o ./build/${BINARY_NAME}-windows-amd64.exe ./cli/main.go

clean:
	go clean
	rm -rf ${OUTPUT_DIR}
