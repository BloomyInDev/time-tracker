BINARY := time-tracker

.PHONY: build build-cli run dev test clean generate

generate:
	go tool templ generate

build: generate
	go build -o bin/$(BINARY) ./cmd/server

build-cli:
	go build -o bin/cli ./cmd/cli

run: build
	./bin/$(BINARY)

test:
	go test ./...

dev:
	go tool templ generate --watch --proxy="http://localhost:8080" --proxyport=7331 --open-browser=false & \
	trap 'kill %1' EXIT; \
	go tool air

clean:
	rm -rf bin tmp *.db
