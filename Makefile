.PHONY: build test lint clean fmt tidy proto

GO ?= go

build:
	$(GO) build ./...

test:
	$(GO) test -race -count=1 -timeout 60s ./...

lint:
	golangci-lint run --timeout 120s ./...

clean:
	rm -f gen/muxcore/indexer/v1/*.go

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

proto:
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		-I proto proto/muxcore/indexer/v1/indexer.proto

ci: lint test build
