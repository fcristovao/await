all: build

build:
	go build -v github.com/betalo-sweden/await

test:
	go test -v

rel:
	GOOS=linux go build github.com/betalo-sweden/await

.PHONY: all build rel
