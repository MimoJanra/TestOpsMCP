# Multi-stage build for Allure MCP Server
FROM golang:1.27-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy source code
COPY . .

# Build the binary (explicit output name without .exe for Linux)
RUN go build -o bin/server ./cmd/server

# Runtime stage - minimal image
FROM alpine:3.21

# Install ca-certificates for HTTPS connections to Allure
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/server /app/server

# Create non-root user
RUN addgroup -g 1000 mcp && \
    adduser -D -u 1000 -G mcp mcp
USER mcp

# Default to HTTP mode
EXPOSE 3000

# Health check (HTTP mode: wget /health; stdio mode: no-op)
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO /dev/null http://localhost:3000/health || exit 1

ENTRYPOINT ["/app/server"]
CMD ["--http"]
