mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir := $(dir $(mkfile_path))

VERSION=$(shell git rev-parse --short HEAD)

build: build-cli 
clean:
	go clean -testcache
test: 
	go test -race -coverprofile=coverage.txt -covermode=atomic

##########
# CLI
##########

build-cli:
	go build -o bin/cli cmd/cli/main.go
run-cli:
	bin/api
develop-api:
	CompileDaemon -build="go build -o bin/cli cmd/cli/main.go" -command="make run-cli" -exclude-dir=.git
build-docs:
	go doc

