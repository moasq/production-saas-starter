# Builder stage
FROM golang:1.25-alpine3.20 AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application with additional flags for production
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /main ./src/main/main.go

# Final stage - using Alpine for smaller image with necessary system files
FROM alpine:3.20

# Install necessary packages and clean up
RUN apk add --no-cache ca-certificates tzdata && \
    rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the binary from builder
COPY --from=builder /main /app/main

# Set proper permissions
RUN chown -R appuser:appgroup /app && \
    chmod +x /app/main

# Use non-root user
USER appuser

# Image metadata
LABEL org.opencontainers.image.title="B2B SaaS Starter Backend" \
      org.opencontainers.image.description="Go backend for B2B SaaS Starter" \
      org.opencontainers.image.vendor="B2B SaaS Starter" \
      org.opencontainers.image.version="1.0.0" \
      org.opencontainers.image.source="https://github.com/yourusername/b2b-saas-starter"

# Expose the port your app runs on
EXPOSE 8080

# Command to run the application
ENTRYPOINT ["/app/main"]