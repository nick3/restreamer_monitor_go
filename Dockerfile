# Build stage
FROM golang:1.21-alpine AS builder

# Install git for version info
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN VERSION=$(git describe --always --dirty 2>/dev/null || echo "unknown") && \
    BUILD_TIME=$(date -u '+%Y-%m-%d %H:%M:%S') && \
    GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") && \
    GO_VERSION=$(go version | cut -d ' ' -f 3) && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
    -ldflags "-X 'main.Version=${VERSION}' \
              -X 'main.BuildTime=${BUILD_TIME}' \
              -X 'main.GitCommit=${GIT_COMMIT}' \
              -X 'main.GoVersion=${GO_VERSION}' \
              -w -s" \
    -o /app/RestreamerMonitor ./main/main.go

# Runtime stage
FROM alpine:latest

# Install ffmpeg for relay functionality and ca-certificates for HTTPS
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -u 1000 appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/RestreamerMonitor /app/RestreamerMonitor

# Create config directory for volume mount
RUN mkdir -p /app/config && chown -R appuser:appuser /app

USER appuser

# Default config path
ENV CONFIG_PATH=/app/config/config.json

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD pgrep -x RestreamerMonitor || exit 1

ENTRYPOINT ["/app/RestreamerMonitor"]
CMD ["monitor", "-c", "/app/config/config.json"]
