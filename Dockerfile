# syntax=docker/dockerfile:1

FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

# Install CA certificates and create non-root user
RUN apk add --no-cache ca-certificates && \
    adduser -u 1000 -D temporal

# Copy the appropriate binary for target architecture
ARG TARGETARCH
COPY dist/nix_linux_${TARGETARCH}/temporal /temporal

# Run as non-root user temporal
USER temporal:temporal

ENTRYPOINT ["/temporal"]