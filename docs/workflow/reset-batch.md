---
id: reset-batch
title: temporal workflow reset-batch
sidebar_label: reset-batch
description: Reset a batch of Workflow Executions by reset type: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew
tags:
	- cli
---


Resetting a Workflow allows the process to resume from a certain point without losing your parameters or [Event History](https://docs.temporal.io/workflows/#event-history).

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

**--dry-run**
Simulate reset without resetting any Workflow Executions.

**--env**
Name of the environment to read environmental variables from. (default: default)

**--exclude-file**
Input file that specifies Workflow Executions to exclude from resetting.

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--input-file**
Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id.

**--input-parallelism**
Number of goroutines to run in parallel. Each goroutine processes one line for every second. (default: 1)

**--input-separator**
Separator for the input file. The default is a tab (	). (default: 	)

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--non-deterministic**
Reset Workflow Execution only if its last Event is WorkflowTaskFailed with a nondeterministic error.

**--query**
Alias: **-q**
Visibility query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query.

**--reason**
Reason for resetting the Workflow Executions.

**--skip-base-is-not-current**
Skip a Workflow Execution if the base Run is not the current Run.

**--skip-current-open**
Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run.

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
Event type to which you want to reset: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

