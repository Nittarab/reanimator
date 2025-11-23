# Development stage
FROM golang:1.21-alpine AS development

WORKDIR /app

# Install development tools
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Development mode - will be overridden by docker-compose
CMD ["go", "run", "cmd/server/main.go"]

# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the server application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o incident-service cmd/server/main.go

# Build the migration tool
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrate/main.go

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/incident-service .
COPY --from=builder /app/migrate .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose ports (8080 for API, 9090 for metrics)
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

CMD ["./incident-service"]
