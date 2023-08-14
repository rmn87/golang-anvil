mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

VERSION=$(shell git rev-parse --short HEAD)

build: build-cli 
clean:
	go clean -testcache
test: 
	go test -race -coverprofile=coverage.txt -covermode=atomic

build:
	go build -o bin/cli cmd/cli/main.go
run:
	bin/cli
build-docs:
	go doc
