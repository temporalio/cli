# Intermediate stage to normalize goreleaser output paths
# This copies both architecture binaries and renames them to clean paths,
# allowing the final stage to select the correct binary using TARGETARCH
FROM scratch AS dist
COPY dist/nix_linux_amd64_v1/temporal /dist/amd64/temporal
COPY dist/nix_linux_arm64_v8.0/temporal /dist/arm64/temporal

FROM alpine:3.22@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412

ARG TARGETARCH

RUN apk add --no-cache ca-certificates
COPY --chmod=755 --from=dist /dist/${TARGETARCH}/temporal /usr/local/bin/temporal
RUN adduser -u 1000 -D temporal
USER temporal

ENTRYPOINT ["temporal"]
