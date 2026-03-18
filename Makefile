VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
LDFLAGS  = -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"
BINARY   = mcp-obtrace

.PHONY: build run run-sse run-streamable-http test lint clean image

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/mcp-obtrace

run: build
	./bin/$(BINARY)

run-sse: build
	./bin/$(BINARY) --transport sse --addr :8000

run-streamable-http: build
	./bin/$(BINARY) --transport streamable-http --addr :8000

test:
	go test ./...

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ dist/

image:
	docker build -t obtrace-mcp:$(VERSION) .
