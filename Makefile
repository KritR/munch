GO ?= go
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
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
	install -d $(BINDIR)
	install -m 0755 bin/munch $(BINDIR)/munch

lint:
	golangci-lint run ./...
