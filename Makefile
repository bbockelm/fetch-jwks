BINARY_NAME := fetch-jwks
BIN_DIR := bin

.PHONY: all build test fmt lint clean

all: build

build:
	go build -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/fetch-jwks

test:
	go test ./...

fmt:
	gofmt -w .

lint:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)
