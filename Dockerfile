FROM golang:1.24-alpine AS base

WORKDIR /app
ENV GOTOOLCHAIN=auto
RUN apk add --no-cache git curl

# -------------------------------
# Stage 1: Development Build
# -------------------------------
FROM base AS dev

RUN go install github.com/air-verse/air@v1.52.3

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

CMD ["air"]

# -------------------------------
# Stage 2: Builder Stage
# -------------------------------
FROM base AS builder

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

COPY go.mod go.sum* ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .

# Build the application binary with caching
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    go build -o file-service ./cmd/main.go

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
