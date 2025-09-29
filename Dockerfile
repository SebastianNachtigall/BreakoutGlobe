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

# Railway provides PORT environment variable at runtime
# We can't use $PORT in EXPOSE since it's not available at build time
EXPOSE 8080

CMD ["./main"]