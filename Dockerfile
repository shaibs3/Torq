# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for go get)
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for HTTPS requests and health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /root/

# Copy the binary and health check script from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/healthcheck.sh .

# Make health check script executable and change ownership
RUN chmod +x healthcheck.sh && \
    chown appuser:appgroup /root/main /root/healthcheck.sh

# Switch to non-root user
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ./healthcheck.sh

# Run the binary
ENTRYPOINT ["./main"] 