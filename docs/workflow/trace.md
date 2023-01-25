---
id: trace
title: temporal workflow trace
sidebar_label: trace
description: Trace progress of a Workflow Execution and its children.
tags:
	- cli
---


The `temporal workflow trace` command tracks the progress of a [Workflow Execution](/workflows#workflow-execution) and any  [Child Workflows](/workflows#child-workflow) it generates.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`temporal workflow trace [command options] [arguments]`

## OPTIONS

**--address**
The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**
Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**
Endpoint for a remote Codec Server.

**--color**
when to use color: auto, always, never. (default: auto)

**--concurrency**
Request concurrency (default: 10)

**--context-timeout**
An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--depth**
Number of Child Workflows to expand, -1 to expand all Child Workflows. (default: -1)

**--env**
Name of the environment to read environmental variables from. (default: default)

**--fold**
Statuses for which Child Workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --no-fold supplied. (default: completed,canceled,terminated)

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-fold**
Disable folding. All Child Workflows within the set depth will be fetched and displayed.

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

