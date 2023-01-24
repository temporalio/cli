---
id: list
title: temporal batch list
sidebar_label: list
description: List Batch operation jobs.
tags:
	- cli
---


When used, `temporal batch list` returns all Batch jobs.

Use the command options listed below to change the information returned by this command.
Make sure to write the command in this format:
`temporal batch list [command options] [arguments]`

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

**--fields**
Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--limit**
Number of items to print. (default: 0)

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager**
Alias: **-P**
Disables the interactive pager.

**--output**
Alias: **-o**
format output as: table, json, card. (default: table)

**--pager**
Sets the pager for Temporal CLI to use.
Options: less, more, favoritePager.

**--time-format**
Format time as: relative, iso, raw. (default: relative)

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

