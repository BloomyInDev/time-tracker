BINARY := time-tracker

.PHONY: build run test clean

build:
	go build -o bin/$(BINARY) ./cmd/server

run: build
	./bin/$(BINARY)

test:
	go test ./...

clean:
	rm -rf bin
