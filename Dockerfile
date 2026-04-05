# -------------------------------
# Stage 1: Development Build
# -------------------------------
FROM golang:alpine AS dev

WORKDIR /app

# Install necessary tools
RUN apk add --no-cache git curl

# Set GOBIN for cleaner binary installs
ENV GOBIN=/usr/local/bin

# Install Air for hot reloading
RUN go install github.com/air-verse/air@v1.52.3

# Copy go mod files and download dependencies
COPY go.mod go.sum* ./
RUN go mod download || true

# Copy the rest of the source code
COPY . .

# Set Air as default command
CMD ["air"]

# -------------------------------
# Stage 2: Builder Stage
# -------------------------------
FROM golang:alpine AS builder

# Build settings for static binary
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

RUN apk add --no-cache git curl

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

# Build the application binary
RUN go build -o file-service ./cmd/main.go

# -------------------------------
# Stage 3: Production Image
# -------------------------------
FROM gcr.io/distroless/static-debian12 AS production

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/file-service .

LABEL org.opencontainers.image.source="https://github.com/dansonserge/DaFileService"
LABEL maintainer="sergedanson@gmail.com"

# Run the binary
CMD ["./file-service"]
