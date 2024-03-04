FROM golang:1.22-bookworm as builder

WORKDIR /app

# Copy everything
COPY . ./

# Build
RUN go build -tags protolegacy ./cmd/temporal

# Use slim container for running
FROM debian:bookworm-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
    ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Copy binary
COPY --from=builder /app/temporal /app/temporal

# Set CLI as primary entrypoint
ENTRYPOINT ["/app/temporal"]
