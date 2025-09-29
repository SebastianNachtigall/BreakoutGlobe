# This Dockerfile is for Railway deployment of the backend
FROM golang:1.21-alpine AS builder

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
WORKDIR /root/

COPY --from=builder /app/main .

# Add health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:$PORT/health || exit 1

# Railway provides PORT environment variable
EXPOSE $PORT

CMD ["./main"]