FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -o orchestrator ./cmd/orchestrator

# Final stage
FROM alpine:latest

RUN apk add --no-cache \
    ca-certificates \
    postgresql-client \
    mysql-client \
    tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/orchestrator /app/orchestrator

# Create directories for configs
RUN mkdir -p /app/configs /app/backups

ENTRYPOINT ["/app/orchestrator"]
