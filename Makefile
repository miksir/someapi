BIN_NAME := someapi

all: build

build:
	go build -o $(BIN_NAME) cmd/app.go

test:
	go test ./...

clean:
	go clean
	rm -f $(BIN_NAME)

run: build
	./$(BIN_NAME)

debug: build
	./$(BIN_NAME) --debug

.PHONY: all build test clean run
