# Multi-stage Dockerfile for UseKuro
# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o usekuro \
    ./cmd/usekuro/

# Runtime stage
FROM scratch

# Copy timezone data and certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/usekuro /usekuro

# Copy example mocks (optional)
COPY --from=builder /app/examples /examples
COPY --from=builder /app/mocks /mocks

# Create non-root user
USER 65534:65534

# Expose common ports for mocks
EXPOSE 8080 8798 9090 2022

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usekuro", "validate", "/examples/http_api.kuro"] || exit 1

# Default command
ENTRYPOINT ["/usekuro"]
CMD ["web", "8798"]

# Labels for metadata
LABEL maintainer="UseKuro Team <hello@usekuro.com>"
LABEL org.opencontainers.image.title="UseKuro"
LABEL org.opencontainers.image.description="Mock any protocol like a master. No coding required."
LABEL org.opencontainers.image.url="https://usekuro.com"
LABEL org.opencontainers.image.source="https://github.com/usekuro/kuro"
LABEL org.opencontainers.image.vendor="UseKuro"
LABEL org.opencontainers.image.licenses="MIT"
