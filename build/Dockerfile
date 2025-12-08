# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.25.4 AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies with cache mount
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    go mod download

# Copy source code
COPY . .

# Build the application with cache mounts
# CGO_ENABLED=0 for static binary that works in scratch
RUN --mount=type=cache,target=/go/pkg/mod,id=gomod \
    --mount=type=cache,target=/root/.cache/go-build,id=gobuild \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ledtx .

# Runtime stage
FROM scratch

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/ledtx /app/ledtx

# Run the application
ENTRYPOINT ["/app/ledtx"]
