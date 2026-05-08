# Build stage
<<<<<<< HEAD
FROM golang:1.25-alpine AS builder
=======
FROM golang:1.22-alpine AS builder
>>>>>>> d564cbdf9eec5568190a1581123e6623293413f1

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o masaar ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy binary from builder
COPY --from=builder /build/masaar .

# Copy migrations
COPY migrations ./migrations

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD /app/masaar ping || exit 1

# Run application
CMD ["/app/masaar"]
