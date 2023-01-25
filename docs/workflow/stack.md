---
id: stack
title: temporal workflow stack
sidebar_label: stack
description: Query a Workflow Execution with __stack_trace as the query type.
tags:
	- cli
---


The `temporal workflow stack` command queries a [Workflow Execution](/workflows#workflow-execution) with `--stack-trace` as the [Query](/workflows#stack-trace-query) type.
Returning the stack trace of all the threads owned by a Workflow Execution can be great for troubleshooting in production.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`temporal workflow stack [command options] [arguments]`

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
Optional query input, in JSON format. For multiple parameters, concatenate them and separate by space.

**--input-file**
Passes optional Query input from a JSON file.
If there are multiple JSON, concatenate them and separate by space or newline.
Input from the command line will overwrite file input.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--reject-condition**
Optional flag for rejecting Queries based on Workflow state. Valid values are "not_open" and "not_completed_cleanly".

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
Workflow Id

