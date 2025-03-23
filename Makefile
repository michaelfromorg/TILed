.PHONY: build clean test test-coverage install run-init run-add run-commit run-log run-push lint help deps

BINARY_NAME=tilcli
COVERAGE_FILE=coverage.out
GO_FILES=$(shell find . -name "*.go" -not -path "./vendor/*")

all: build

help:
	@echo "Available commands:"
	@echo "  make build          - Build the TIL binary"
	@echo "  make deps           - Download dependencies"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make test           - Run all tests"
	@echo "  make test-verbose   - Run all tests with verbose output"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make install        - Install TIL to GOPATH/bin"
	@echo "  make lint           - Run the linter"
	@echo "  make run-init       - Run 'til init'"
	@echo "  make run-add FILE=x - Run 'til add FILE'"
	@echo "  make run-commit     - Run 'til commit -m \"message\"'"
	@echo "  make run-log        - Run 'til log'"
	@echo "  make run-push       - Run 'til push'"

deps:
	go mod tidy
	go mod download

build: deps
	go build -o $(BINARY_NAME) -v

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)

test:
	go test ./...

test-verbose:
	go test -v ./...

test-coverage:
	go test -coverprofile=$(COVERAGE_FILE) ./...
	go tool cover -html=$(COVERAGE_FILE)

install:
	go install

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Example usage commands
run-init: build
	./$(BINARY_NAME) init

run-add: build
	@if [ "$(FILE)" = "" ]; then \
		echo "Error: Please specify a file with FILE=<filename>"; \
		echo "Example: make run-add FILE=example.txt"; \
		exit 1; \
	fi
	./$(BINARY_NAME) add $(FILE)

run-commit: build
	@if [ "$(MESSAGE)" = "" ]; then \
		MESSAGE="Learned something new today"; \
	fi
	./$(BINARY_NAME) commit -m "$(MESSAGE)"

run-log: build
	@if [ "$(NUM)" = "" ]; then \
		./$(BINARY_NAME) log; \
	else \
		./$(BINARY_NAME) log -n $(NUM); \
	fi

run-push: build
	./$(BINARY_NAME) push
