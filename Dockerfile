# Start from the official Golang image for building
FROM golang:1.24 AS builder

WORKDIR /app

# Use Go proxy with direct fallback to handle both TLS timeouts and access restrictions
# This tries proxy.golang.org first, then falls back to direct VCS access
ENV GOPROXY=https://proxy.golang.org,direct
ENV GOSUMDB=sum.golang.org

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o app main.go

# Use a minimal base image for running
FROM alpine:3.18

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy the built binary from builder
COPY --from=builder /app/app ./app

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (change if your app uses a different port)
EXPOSE 8080

# Run the binary
CMD ["./app"]
