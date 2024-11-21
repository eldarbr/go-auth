DISABLED_LINTERS="depguard,paralleltest,execinquery,gochecknoglobals"

ERR_NO_DB_URI='WARNING: The database unit tests will be skipped. Please provide a connection uri to complete tests.\n'

all:

build:
	go build -o bin/ cmd/go-auth/go-auth.go

test:
	@if [ ! -n "$(TEST_DB_URI)" ]; then echo $(ERR_NO_DB_URI); fi
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
	find . -name "*.pprof" -print0 | xargs -0 rm

generate:
