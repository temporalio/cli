FROM --platform=$BUILDARCH scratch AS dist
COPY ./dist/nix_linux_amd64_v1/temporal /dist/amd64/temporal
COPY ./dist/nix_linux_arm64/temporal /dist/arm64/temporal

FROM alpine:3.22
ARG TARGETARCH
COPY --from=dist /dist/$TARGETARCH/temporal /usr/local/bin/temporal

ENTRYPOINT ["temporal"]
