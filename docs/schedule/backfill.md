---
id: %s
title: %s
sidebar_label: %s
description: %s
tags:
---

### backfill

Backfills a past time range of actions

**--address**
host:port for Temporal frontend service

**--codec-auth**
Authorization header to set for requests to Codec Server

**--codec-endpoint**
Remote Codec Server Endpoint

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
Optional timeout for context of RPC call in seconds (default: 5)

**--end-time**
Backfill end time

**--env**
Env name to read the client environment variables from (default: default)

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace**
Alias: **-n**
Temporal workflow namespace (default: default)

**--overlap-policy**
Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll

**--schedule-id**
Alias: **-s**
Schedule Id

**--start-time**
Backfill start time

**--tls-ca-path**
Path to server CA certificate

**--tls-cert-path**
Path to x509 certificate

**--tls-disable-host-verification**
Disable tls host name verification (tls must be enabled)

**--tls-key-path**
Path to private key

**--tls-server-name**
Override for target server name

