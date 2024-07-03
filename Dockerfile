# Use a Go base image for building
FROM golang:1.22-alpine AS builder

# Set a working directory
WORKDIR /app

# Copy the Go modules and sum files first (for caching)
COPY go.mod go.sum ./

# Download Go module dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -o /server -ldflags="-s -w" ./cmd/app

# Use a fresh alpine base for the runtime
FROM alpine:latest

# Install ca-certificates
RUN apk add --no-cache ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /server /server

# Set a default value for the PORT environment variable
ENV PORT 8090

# Expose port $PORT for the application
EXPOSE $PORT

# Command to run the binary
CMD ["/bin/sh", "-c", "/server"]