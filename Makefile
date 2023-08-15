VERSION=$(shell git describe --tags --abbrev=0)

clean:
	go clean -testcache
test: 
	go test
build:
	go build -ldflags "-X anvil.VERSION=${VERSION}" -o bin/anvil cmd/cli/main.go
build-docs:
	go doc
