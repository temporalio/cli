ARG BASE_SERVER_IMAGE=temporalio/base-server:1.2.0
ARG GOPROXY

##### builder #####
FROM golang:1.18-alpine3.14 AS temporal-cli-builder

RUN apk add --update --no-cache \
    make \
    git

WORKDIR /home/temporal-cli-builder

# pre-build dependecies to improve subsequent build times
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

##### temporal CLI #####
FROM ${BASE_SERVER_IMAGE} AS ui-server
WORKDIR /etc/temporal

COPY --from=temporal-cli-builder /home/temporal-cli-builder/temporal /usr/local/bin

EXPOSE 7233
EXPOSE 8233

# Keep the container running.
ENTRYPOINT ["/temporal", "server", "start-dev", "-n", "default", "--ip" , "0.0.0.0"]
