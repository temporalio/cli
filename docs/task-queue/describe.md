---
id: describe
title: temporal task-queue describe
sidebar_label: describe
description: Temporal CLI operation for ....
tags:
	- cli
---

### describe

Describes the Workers that have recently polled on this Task Queue

    The Server records the last time of each poll request.
    
    Poll requests can last up to a minute, so a LastAccessTime under a minute is normal.
    If it's over a minute, then likely either the Worker is at capacity (all Workflow and Activity slots are full) or it has shut down.
    Once it has been 5 minutes since the last poll request, the Worker is removed from the list.
    
    RatePerSecond is the maximum Activities per second the Worker will execute.

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

**--env**
Name of the environment to read environmental variables from. (default: default)

**--fields**
customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--output**
Alias: **-o**
format output as: table, json, card. (default: table)

**--task-queue**
Alias: **-t**
Task Queue name.

**--task-queue-type**
Task Queue type [workflow|activity] (default: workflow)

**--time-format**
format time as: relative, iso, raw. (default: relative)

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

