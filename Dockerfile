# syntax=docker/dockerfile:1

# Build stage to extract CA certificates
FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412 AS certs
RUN apk add --no-cache ca-certificates

# Final distroless stage
FROM gcr.io/distroless/static-debian12:nonroot

# Copy CA certificates from certs stage
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the appropriate binary for target architecture
ARG TARGETARCH
COPY dist/nix_linux_${TARGETARCH}/temporal /temporal

# Set entrypoint
ENTRYPOINT ["/temporal"]