GO ?= go

.PHONY: fmt test build lint

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	mkdir -p bin
	$(GO) build -o bin/munch-widget ./cmd/munch-widget

lint:
	golangci-lint run ./...
