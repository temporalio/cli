---
id: fail
title: temporal activity fail
sidebar_label: fail
description: Temporal CLI operation for ....
tags:
	- cli
---

### fail

Fails an [Activity](https://docs.temporal.io/activities).

**--activity-id**
Identifies the Activity to fail.

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

**--detail**
Detail to fail the Activity.

**--env**
Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--identity**
Specify the operator's identity.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--reason**
Reason to fail the Activity.

**--run-id**
Alias: **-r**
Identifies the current Workflow Run.

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
Identifies the Workflow that the Activity is running on.

