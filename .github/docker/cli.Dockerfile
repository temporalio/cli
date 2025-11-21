# syntax=docker/dockerfile:1

ARG ALPINE_IMAGE
ARG BUILDARCH

# Build stage - copy binaries from goreleaser output
FROM --platform=$BUILDARCH scratch AS dist
COPY dist/nix_linux_amd64_v1/temporal /dist/amd64/temporal
COPY dist/nix_linux_arm64_v8.0/temporal /dist/arm64/temporal

# Stage to extract CA certificates and create user files
FROM ${ALPINE_IMAGE} AS certs
RUN apk add --no-cache ca-certificates && \
    adduser -u 1000 -D temporal

# Final stage - minimal scratch-based image
FROM scratch

ARG TARGETARCH

# Copy CA certificates from certs stage
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy passwd and group files for non-root user
COPY --from=certs /etc/passwd /etc/passwd
COPY --from=certs /etc/group /etc/group

# Copy the appropriate binary for target architecture
COPY --from=dist /dist/$TARGETARCH/temporal /temporal

# Run as non-root user temporal (uid 1000)
USER 1000:1000

ENTRYPOINT ["/temporal"]
