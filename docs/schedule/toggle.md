---
id: toggle
title: temporal schedule toggle
sidebar_label: toggle
description: Pauses or unpauses a schedule.
tags:
	- cli
---


The `temporal schedule toggle` command can pause and unpause a [Schedule](/workflows#schedule).

Toggling a Schedule requires a reason to be entered on the command line.
Use `--reason` to note the issue leading to the pause or unpause.

Schedule toggles are passed in this format:
` temporal schedule toggle --sid 'your-schedule-id' --pause --reason "paused because the database is down"`
`temporal schedule toggle --sid 'your-schedule-id' --unpause --reason "the database is back up"`

Use the options provided below to change this command's behavior.

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

**--pause**
Pauses the schedule.

**--reason**
Free-form text to describe reason for pause/unpause. (default: (no reason provided))

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

**--unpause**
Unpauses the schedule.

