ARG BASE_SERVER_IMAGE=temporalio/base-server:1.2.0
ARG GOPROXY

##### tctl builder #####
FROM golang:1.18-alpine3.14 AS tctl-builder

RUN apk add --update --no-cache \
    make \
    git

WORKDIR /home/tctl-builder

# pre-build dependecies to improve subsequent build times
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

##### tctl #####
FROM ${BASE_SERVER_IMAGE} AS ui-server
WORKDIR /etc/temporal

COPY --from=tctl-builder /home/tctl-builder/tctl /usr/local/bin
COPY --from=tctl-builder /home/tctl-builder/tctl-authorization-plugin /usr/local/bin

# Keep the container running.
ENTRYPOINT ["tail", "-f", "/dev/null"]