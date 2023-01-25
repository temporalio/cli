---
id: backfill
title: temporal schedule backfill
sidebar_label: backfill
description: Backfills a past time range of actions.
tags:
	- cli
---


The `temporal schedule backfill` command executes Actions ahead of their specified time range.
Backfilling can be used to fill in [Workflow Runs](/workflows#run-id) from a time period when the Schedule was paused, or from before the Schedule was created.

```
temporal schedule backfill --sid 'your-schedule-id' \
--overlap-policy 'BufferAll' 				\
--start-time '2022-05-0101T00:00:00Z'		\
--end-time '2022-05-31T23:59:59Z'
```

Use the options provided below to change this command's behavior.

## OPTIONS

**--address**
The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**
Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**
Endpoint for a remote Codec Server.

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--end-time**
Backfill end time.

**--env**
Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--overlap-policy**
Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll.

**--schedule-id**
Alias: **-s**
Schedule Id

**--start-time**
Backfill start time.

**--tls-ca-path**
Path to server CA certificate.

**--tls-cert-path**
Path to x509 certificate.

**--tls-disable-host-verification**
Disables TLS host name verification if already enabled.

**--tls-key-path**
Path to private certificate key.

**--tls-server-name**
Provides an override for the target TLS server name.

