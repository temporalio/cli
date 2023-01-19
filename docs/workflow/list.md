---
id: list
title: temporal workflow list
sidebar_label: list
description: Temporal CLI operation for ....
tags:
	- cli
---

### list

List Workflow Executions based on a Query.

By default, this command lists up to 10 closed Workflow Executions.

**--address**
The host and port (formatted as host:port) for the Temporal Frontend Service.

**--archived**
List archived Workflow Executions (EXPERIMENTAL)

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
customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--limit**
number of items to print (default: 0)

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager**
Alias: **-P**
disable interactive pager

**--output**
Alias: **-o**
format output as: table, json, card. (default: table)

**--pager**
pager to use: less, more, favoritePager..

**--query**
Alias: **-q**
Filter results using SQL like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for details

**--time-format**
format time as: relative, iso, raw. (default: relative)

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

