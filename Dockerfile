# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire project
COPY . .

# Download and extract compressed GeoIP database during build
ARG GEOIP_DOWNLOAD_URL
ENV GEOIP_DOWNLOAD_URL=${GEOIP_DOWNLOAD_URL}

RUN if [ -n "$GEOIP_DOWNLOAD_URL" ]; then \
        echo "Downloading and extracting GeoIP databases..."; \
        chmod +x setup-geoip.sh && \
        ./setup-geoip.sh; \
    else \
        echo "WARNING: GEOIP_DOWNLOAD_URL not set, skipping GeoIP download"; \
    fi

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

# Create geoip directory
RUN mkdir -p pkg/geoip

# Copy geoip database if they exists
COPY --from=builder /app/pkg/geoip/*.mmdb ./pkg/geoip/ 2>/dev/null || true

EXPOSE 8080

CMD ["./main"]