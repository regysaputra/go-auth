# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

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

EXPOSE 8080

CMD ["./main"]