FROM golang:1.17 AS builder

 RUN go install github.com/temporalio/cli/cmd/temporal@latest

 FROM gcr.io/distroless/base-debian11

 COPY --from=builder /go/bin/temporal /

 EXPOSE 7233
 ENTRYPOINT ["/temporal", "server", "start-dev", "-n", "default", "--ip" , "0.0.0.0"]
