---
id: reset
title: temporal workflow reset
sidebar_label: reset
description: Temporal CLI operation for ....
tags:
	- cli
---

### reset

Resets a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) by Event Id or reset type.

A reset allows the Workflow to be resumed from a certain point without losing your parameters or [Event History](https://docs.temporal.io/workflows/#event-history).

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

**--event-id**
The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others.

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--reapply-type**
Event types to reapply after the reset point: , Signal, None. (default: All)

**--reason**
Reason to reset.

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
Event type to which you want to reset: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

**--workflow-id**
Alias: **-w**
Workflow Id

