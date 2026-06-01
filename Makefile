BINARY      := pandaprobe
PKG         := github.com/chirpz-ai/pandaprobe-cli
VERSION_PKG := $(PKG)/internal/version

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE      ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
	-X $(VERSION_PKG).Version=$(VERSION) \
	-X $(VERSION_PKG).Commit=$(COMMIT) \
	-X $(VERSION_PKG).Date=$(DATE)

COVER_THRESHOLD := 80

.PHONY: all build install test test-cover cover lint fmt tidy vet snapshot clean

all: build

## build: compile the binary for the current platform
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

## install: install the binary into $GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## test: run all unit tests with the race detector
test:
	go test -race ./...

## test-cover: run tests and print per-package coverage
test-cover:
	go test -race -cover ./...

## cover: enforce a coverage threshold on internal packages
cover:
	@go test -coverprofile=coverage.out ./internal/... >/dev/null
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	echo "internal coverage: $$total% (threshold $(COVER_THRESHOLD)%)"; \
	awk "BEGIN{exit !($$total >= $(COVER_THRESHOLD))}" || { echo "coverage below threshold"; exit 1; }

## lint: run golangci-lint
lint:
	golangci-lint run

## fmt: format the code
fmt:
	gofmt -s -w .

## vet: run go vet
vet:
	go vet ./...

## tidy: tidy go.mod/go.sum
tidy:
	go mod tidy

## snapshot: build a local cross-platform snapshot with GoReleaser
snapshot:
	goreleaser release --snapshot --clean

## clean: remove build artifacts
clean:
	rm -f $(BINARY) coverage.out
	rm -rf dist
