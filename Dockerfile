FROM golang:1.21-alpine AS builder

# Install git (needed if your go.mod pulls modules from git)
RUN apk add --no-cache git

# Set working directory inside container
WORKDIR /app

# Copy go.mod and go.sum first (to leverage Docker cache for deps)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app (binary named "app")
RUN go build -o app ./cmd/main.go

FROM alpine:3.18

RUN apk add --no-cache \
    ffmpeg \
    ca-certificates

WORKDIR /app

# Copy binary from build stage
COPY --from=builder /app/app .

# Expose port if needed (example 8080)
EXPOSE 8080

# Run the app
CMD ["./app"]