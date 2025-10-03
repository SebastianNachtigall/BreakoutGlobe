FROM golang:1.21-alpine AS development

WORKDIR /app

# Install air for hot reloading and git (required for go install)
RUN apk add --no-cache git
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Expose port
EXPOSE 8080

# Use air for hot reloading in development
CMD ["air", "-c", ".air.toml"]

# Production build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/main .

# Declare volume mount point for persistent storage
# This ensures Railway's volume mount takes precedence
VOLUME ["/app/uploads"]

# Add health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

EXPOSE 8080
CMD ["./main"]