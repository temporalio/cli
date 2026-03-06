FROM alpine:3.23

ARG TARGETARCH

RUN apk add --no-cache ca-certificates tzdata
COPY dist/${TARGETARCH}/temporal /usr/local/bin/temporal
RUN adduser -u 1000 -D temporal
USER temporal

ENTRYPOINT ["temporal"]
