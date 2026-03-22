FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X main.version=$(cat image-tag 2>/dev/null || echo dev)" \
    -o /mcp-obtrace ./cmd/mcp-obtrace

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

RUN useradd -r -s /bin/false mcp
USER mcp

COPY --from=builder /mcp-obtrace /usr/local/bin/mcp-obtrace

EXPOSE 8000

ENTRYPOINT ["mcp-obtrace"]
CMD ["--transport", "sse", "--addr", ":8000"]
