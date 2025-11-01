# --- Build Stage ---
# We use a specific Go version in a lightweight Alpine container as our build environment.
FROM golang:1.25-alpine AS builder

# Install build dependencies. CGO is needed for pgx.
RUN apk add --no-cache git build-base

WORKDIR /app

# Copy and download dependencies first to leverage Docker's layer caching.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the application.
# We build a statically linked binary to keep the final image small and self-contained.
# The 'templates' are already embedded, so we don't need to copy them.
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags '-w -s' -o /app/server ./main.go

# --- Final Stage ---
# We use a minimal Alpine image for a tiny, secure final container.
FROM alpine:latest

# Add SSL certificates
RUN apk add --no-cache ca-certificates

# Copy the built application binary from the builder stage.
COPY --from=builder /app/server /app/server

# Expose the port our application listens on.
EXPOSE 8080

# Set the command to run when the container starts.
CMD ["/app/server"]
