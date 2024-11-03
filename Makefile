all:

build:

test:

fmt:
	go fmt ./...

lint:
	go vet ./...
	golangci-lint run

generate:
