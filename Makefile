all: build

build:
	go build -v github.com/betalo-sweden/await

deps:
	go get -u github.com/golang/dep
	dep ensure

test:
	go test -v

rel:
	GOOS=darwin GOARCH=amd64 go build -o await-darwin-amd64 github.com/betalo-sweden/await
	GOOS=linux  GOARCH=amd64 go build -o await-linux-amd64  github.com/betalo-sweden/await

.PHONY: all build rel test
