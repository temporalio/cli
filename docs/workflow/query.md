---
id: query
title: temporal workflow query
sidebar_label: query
description: Temporal CLI operation for ....
tags:
	- cli
---

### query

[Query](https://docs.temporal.io/workflows/#query) a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution).

    Queries can retrieve all or part of the Workflow state within given parameters.
    Queries can also be used on completed [Workflows](https://docs.temporal.io/workflows/).

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

**--type**
The query type you want to run.

**--workflow-id**
Alias: **-w**
Workflow Id

