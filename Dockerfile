# Build stage
FROM golang:1.18-alpine AS builder

WORKDIR /app

# Install git for fetching Go dependencies
RUN apk add --no-cache git

# Copy dependency files first for layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary with CGO disabled for static linking
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/jsextractor .

# Runtime stage
FROM alpine:3.17

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

# Copy the compiled binary from the builder stage
COPY --from=builder /usr/local/bin/jsextractor /usr/local/bin/jsextractor

# Set the entrypoint
ENTRYPOINT ["jsextractor"]