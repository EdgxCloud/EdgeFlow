# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary (main_integrated.go for production)
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static" -w -s' -o edgeflow ./cmd/edgeflow/main_integrated.go

# Frontend build stage
FROM node:18-alpine AS frontend-builder

WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies
RUN npm ci

# Copy frontend source
COPY web/ .

# Build frontend
RUN npm run build

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 edgeflow && \
    adduser -u 1000 -G edgeflow -s /bin/sh -D edgeflow

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/edgeflow .

# Copy frontend build
COPY --from=frontend-builder /app/web/dist ./web/dist

# Copy config
COPY configs/default.yaml ./configs/

# Create data directory
RUN mkdir -p /app/data /app/logs && \
    chown -R edgeflow:edgeflow /app

# Switch to non-root user
USER edgeflow

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Set environment variables
ENV EDGEFLOW_SERVER_HOST=0.0.0.0 \
    EDGEFLOW_SERVER_PORT=8080 \
    EDGEFLOW_DATABASE_PATH=/app/data/edgeflow.db \
    EDGEFLOW_LOGGING_LEVEL=info

# Run
ENTRYPOINT ["./edgeflow"]
