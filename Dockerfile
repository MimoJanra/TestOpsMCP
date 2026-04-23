# Multi-stage build for Allure MCP Server
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git make

# Copy source code
COPY . .

# Build the binary
RUN make build

# Runtime stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS connections to Allure
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bin/server.exe /app/server.exe

# Create non-root user
RUN addgroup -g 1000 mcp && \
    adduser -D -u 1000 -G mcp mcp
USER mcp

# Default to HTTP mode
EXPOSE 3000

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD test -S /proc/net/unix || exit 1

ENTRYPOINT ["/app/server.exe"]
CMD ["--http"]
