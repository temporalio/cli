# syntax=docker/dockerfile:1

FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412
ARG TARGETARCH
RUN apk add --no-cache ca-certificates
COPY ./dist/nix_linux_${TARGETARCH}_*/temporal /usr/local/bin/temporal
RUN adduser -u 1000 -D temporal
USER temporal

ENTRYPOINT ["temporal"]
