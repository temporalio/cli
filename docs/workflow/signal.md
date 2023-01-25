---
id: signal
title: temporal workflow signal
sidebar_label: signal
description: Signal Workflow Execution by Id or List Filter.
tags:
	- cli
---


The `temporal workflow signal` command is used to [Signal](/workflows#signal) a [Workflow Execution](/workflows#workflow-execution) by ID or [List Filter](/visibility#list-filter).

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`temporal workflow signal [command options] [arguments]`

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

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--input**
Alias: **-i**
Input for the signal (JSON).

**--input-file**
Input for the signal from file (JSON).

**--name**
Signal Name

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--query**
Alias: **-q**
Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.

**--reason**
Reason for signaling with List Filter.

**--run-id**
Alias: **-r**
Run Id

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

**--workflow-id**
Alias: **-w**
Signal Workflow Execution by Id.

**--yes**
Alias: **-y**
Confirm all prompts.

