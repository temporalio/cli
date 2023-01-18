---
id:
title:
sidebar_label:
description:
tags:
---


### reset-batch

Reset a batch of Workflow Executions by reset type: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

>Resetting a Workflow allows the process to resume from a certain point without losing your parameters or Event History.

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

**--dry-run**
Simulate reset without resetting any Workflow Executions.

**--env**
Env name to read the client environment variables from (default: default)

**--exclude-file**
Input file that specifies Workflow Executions to exclude from resetting.

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input-file**
Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id.

**--input-parallelism**
Number of goroutines to run in parallel. Each goroutine processes one line for every second. (default: 1)

**--input-separator**
Separator for the input file. The default is a tab (	). (default: 	)

**--namespace**
Alias: **-n**
Temporal workflow namespace (default: default)

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
Path to server CA certificate

**--tls-cert-path**
Path to x509 certificate

**--tls-disable-host-verification**
Disable tls host name verification (tls must be enabled)

**--tls-key-path**
Path to private key

**--tls-server-name**
Override for target server name

**--type**
Event type to which you want to reset: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

