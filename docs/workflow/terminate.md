---
id: terminate
title: temporal workflow terminate
sidebar_label: terminate
description: Terminate Workflow Execution by Id or List Filter.
tags:
	- cli
---


The `temporal workflow terminate` command terminates a [Workflow Execution](/workflows#workflow-execution)

Terminating a running Workflow Execution records a [`WorkflowExecutionTerminated` event](/events#workflowexecutionterminated) as the closing Event in the [Event History](/workflows#event-history).
Any further [Command](/workflows#command) Tasks cannot be scheduled after running this command.

Use the options listed below to change termination behavior.
Make sure to write the command as follows:
`temporal workflow terminate [command options] [arguments]`

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

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--query**
Alias: **-q**
Terminate Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.

**--reason**
Reason for termination.

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
Terminate Workflow Execution by Id.

**--yes**
Alias: **-y**
Confirm all prompts.

