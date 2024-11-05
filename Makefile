DISABLED_LINTERS="depguard,paralleltest,execinquery,gochecknoglobals"

all:

build:
	go build -o bin/ cmd/go-auth/go-auth.go

test:
	go test ./... -t-db-uri="$(TEST_DB_URI)"

fmt:
	go fmt ./...

lint:
	go vet ./...
	golangci-lint run --enable-all --color=never --disable=$(DISABLED_LINTERS)

coverage:
	mkdir -p bin
	go test -coverprofile=bin/cover.prof ./... -t-db-uri="$(TEST_DB_URI)"
	go tool cover -html=bin/cover.prof -o bin/coverage.html

clean:
	rm -rf bin

generate:
