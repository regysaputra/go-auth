# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Verify the database files exist
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
    fi

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

# Copy entire geoip directory (including any .mmdb files if they exist)
COPY --from=builder /app/pkg ./pkg

EXPOSE 8080

CMD ["./main"]