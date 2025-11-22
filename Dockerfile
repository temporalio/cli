# syntax=docker/dockerfile:1

ARG TARGETARCH

# Build stage - copy binaries from goreleaser output
FROM --platform=$TARGETARCH scratch AS dist
COPY dist/nix_linux_amd64_v1/temporal /dist/amd64/temporal
COPY dist/nix_linux_arm64_v8.0/temporal /dist/arm64/temporal

# Stage to extract CA certificates and create user files
FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 AS certs
RUN apk add --no-cache ca-certificates && \
    adduser -u 1000 -D temporal

# Final scratch stage - completely minimal base
FROM scratch

ARG TARGETARCH

# Copy CA certificates from certs stage
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy passwd and group files for non-root user
COPY --from=certs /etc/passwd /etc/passwd
COPY --from=certs /etc/group /etc/group

# Copy the appropriate binary for target architecture
COPY --from=dist /dist/$TARGETARCH/temporal /temporal

# Run as non-root user temporal
USER temporal:temporal

ENTRYPOINT ["/temporal"]