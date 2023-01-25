---
id: delete
title: temporal schedule delete
sidebar_label: delete
description: Deletes a schedule.
tags:
	- cli
---


The `temporal schedule delete` command deletes a [Schedule](/workflows#schedule).
Deleting a Schedule does not affect any [Workflows](/workflows) started by the Schedule.

[Workflow Executions](/workflows#workflow-execution) started by Schedules can be cancelled or terminated like other Workflow Executions.
However, Workflow Executions started by a Schedule can be identified by their [Search Attributes](/visibility#search-attribute), making them targetable by batch command for termination.

`temporal schedule delete --sid 'your-schedule-id' [command options] [arguments]`

Use the options below to change the behavior of this command.

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

**--schedule-id**
Alias: **-s**
Schedule Id

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

