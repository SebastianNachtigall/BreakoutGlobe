# This Dockerfile is for Railway deployment of the backend
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (required for go install)
RUN apk add --no-cache git

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy backend source code
COPY backend/ .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates wget
WORKDIR /app

COPY --from=builder /app/main .

# Add health check endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Railway provides PORT environment variable at runtime
# We can't use $PORT in EXPOSE since it's not available at build time
EXPOSE 8080

CMD ["./main"]