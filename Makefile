all:

build:
	go build -o bin/ cmd/go-auth/go-auth.go

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	go vet ./...
	golangci-lint run

coverage:
	mkdir -p bin
	go test -coverprofile=bin/cover.prof ./...
	go tool cover -html=bin/cover.prof -o bin/coverage.html

clean:
	rm -rf bin

generate:
