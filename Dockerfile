# ---------- builder ----------
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Install git (if needed for `go get`), ca-certificates for go http clients when building (not strictly necessary here)
RUN apk add --no-cache git

# Use module cache
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags="-s -w" -o /out/main ./cmd/server

# Build the migrate binary (if you have a cmd/migrate)
RUN if [ -d "./cmd/migrate" ]; then \
      CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate; \
    fi

# ---------- runtime ----------
FROM alpine:3.19
# Install required system dependencies (certs, timezone if needed) and drop caches
RUN apk add --no-cache ca-certificates tzdata \
  && update-ca-certificates || true

# Create non-root user
RUN adduser -D -g '' appuser

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /out/main ./main
COPY --from=builder /out/migrate ./migrate

# Ensure proper ownership
RUN chown appuser:appuser /app/main /app/migrate || true
USER appuser

# Mountpoint for optional data (GeoIP). The app expects files under /app/pkg/geoip
VOLUME ["/app/pkg/geoip"]

EXPOSE 8080

# Default command (run main). Use `docker-compose` override if needed.
CMD ["./main"]
