# Build stage
FROM golang:1.20-alpine AS builder

WORKDIR /app

# Install build dependencies including certificates for go-mitmproxy
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev

# Copy go mod and sum files
COPY go.mod ./
# Create empty go.sum if it doesn't exist
RUN touch go.sum

# Download dependencies
RUN go mod download
RUN go mod tidy

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mitm-proxy .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates and other runtime dependencies for go-mitmproxy
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from the builder stage
COPY --from=builder /app/mitm-proxy .

# Create directory for certificates
RUN mkdir -p /app/certs

# Expose the proxy port and web interface port
EXPOSE 8080 8081

# Set environment variables
ENV LISTEN_ADDR=:8080
ENV LOGGER_ENDPOINT=http://logger-service/api/logs
ENV WEB_INTERFACE=true
ENV WEB_PORT=8081

# Run the application
CMD ["./mitm-proxy"]