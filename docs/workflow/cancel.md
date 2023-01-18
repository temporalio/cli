---
id: %s
title: %s
sidebar_label: %s
description: %s
tags:
---

### cancel

Cancel a Workflow Execution.

    Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History.
    
    After cancellation, the Workflow Execution can perform cleanup work,and a new command task will be scheduled.

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

**--env**
Env name to read the client environment variables from (default: default)

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace**
Alias: **-n**
Temporal workflow namespace (default: default)

**--query**
Alias: **-q**
Cancel Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.

**--reason**
Reason for canceling with List Filter.

**--run-id**
Alias: **-r**
Run Id

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

**--workflow-id**
Alias: **-w**
Cancel Workflow Execution by Id.

**--yes**
Alias: **-y**
Confirm all prompts.

