.PHONY: all build check clean deps dist fmt fmt-check help install lint run-add run-commit run-init run-log run-push test test-coverage test-race test-verbose vet

BINARY_NAME := til
COVERAGE_FILE := coverage.out
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X github.com/michaelfromorg/tiled/cmd.Version=$(VERSION)

all: build

help:
	@echo "Available commands:"
	@echo "  make build          - Build the til binary"
	@echo "  make check          - Run formatting, vet, unit, integration, and race checks"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Download dependencies"
	@echo "  make dist           - Build release archives and SHA-256 checksums"
	@echo "  make fmt            - Format Go source files"
	@echo "  make install        - Install til to GOPATH/bin"
	@echo "  make test           - Run all tests"
	@echo "  make test-coverage  - Write coverage.out and print a coverage summary"
	@echo "  make test-race      - Run all tests with the race detector"
	@echo "  make vet            - Run go vet"

deps:
	go mod download

build:
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/til

dist:
	VERSION="$(VERSION)" DIST_DIR="dist" bash ./scripts/build-release.sh

clean:
	go clean
	rm -rf dist
	rm -f $(BINARY_NAME) $(COVERAGE_FILE)

fmt:
	gofmt -w ./cmd ./internal ./test

fmt-check:
	@files="$$(gofmt -l ./cmd ./internal ./test)"; \
	if [ -n "$$files" ]; then \
		echo "The following files need gofmt:"; \
		echo "$$files"; \
		exit 1; \
	fi

vet:
	go vet ./...

lint: fmt-check vet

test:
	go test ./...

test-verbose:
	go test -v ./...

test-race:
	go test -race ./...

test-coverage:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -func=$(COVERAGE_FILE)

check: fmt-check vet test test-race build

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/til

run-init: build
	./$(BINARY_NAME) init

run-add: build
	@if [ -z "$(FILE)" ]; then \
		echo "Usage: make run-add FILE=example.txt"; \
		exit 1; \
	fi
	./$(BINARY_NAME) add "$(FILE)"

run-commit: build
	./$(BINARY_NAME) commit -m "$(or $(MESSAGE),Learned something new today)"

run-log: build
	./$(BINARY_NAME) log -n "$(or $(NUM),10)"

run-push: build
	./$(BINARY_NAME) push
