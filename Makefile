GO ?= go
PREFIX ?= /usr/local
VERSION ?= dev
LDFLAGS ?= -X github.com/krithikr/munch/internal/command.Version=$(VERSION)

.PHONY: fmt test build install lint

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

build:
	mkdir -p bin
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/munch ./cmd/munch

install: build
	install -d $(PREFIX)/bin
	install bin/munch $(PREFIX)/bin/munch

lint:
	golangci-lint run ./...
