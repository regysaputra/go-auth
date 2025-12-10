# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install git for go modules
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project (including pre-downloaded GeoIP databases from host)
COPY . .

# Verify the database files exist (for logging only, non-critical)
RUN echo "Checking for GeoIP databases..."; \
    if [ -f pkg/geoip/GeoLite2-City.mmdb ]; then \
        echo "✓ GeoLite2-City.mmdb found"; \
        ls -lh pkg/geoip/GeoLite2-City.mmdb; \
    else \
        echo "⚠ GeoLite2-City.mmdb not found"; \
    fi; \
    if [ -f pkg/geoip/GeoLite2-ASN.mmdb ]; then \
        echo "✓ GeoLite2-ASN.mmdb found"; \
        ls -lh pkg/geoip/GeoLite2-ASN.mmdb; \
    else \
        echo "⚠ GeoLite2-ASN.mmdb not found"; \
    fi || true

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy the public directory with openapi.yml
COPY --from=builder /app/public ./public

# Copy templates directory
COPY --from=builder /app/templates ./templates

# Copy entire pkg directory (includes GeoIP databases)
COPY --from=builder /app/pkg ./pkg

EXPOSE 8080

CMD ["./main"]