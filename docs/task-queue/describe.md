---
id: describe
title: temporal task-queue describe
sidebar_label: describe
description: Describes the Workers that have recently polled on this Task Queue.
tags:
	- cli
---


The `temporal task-queue describe` command provides [poller](/applcation-development/worker-performance#poller-count) information for a given [Task Queue](/tasks#task-queue).

The [Server](/clusters#temporal-server) records the last time of each poll request.
Should `LastAccessTime` exceeds one minute, it's likely that the Worker is at capacity (all Workflow and Activity slots are full) or that the Worker has shut down.
[Workers](/workers) are removed if 5 minutes have passed since the last poll request.

Use the options listed below to modify what this command returns.
Make sure to write the command as follows:
`temporal task-queue describe [command options] [arguments]`

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

**--env**
Name of the environment to read environmental variables from. (default: default)

**--fields**
Customize fields to print. Set to 'long' to automatically print more of main fields.

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
Format time as: relative, iso, raw. (default: relative)

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

