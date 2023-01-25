---
id: search-attribute
title: temporal operator search-attribute
sidebar_label: search-attribute
description: Removes custom search attribute metadata only (Elasticsearch index schema is not modified).
tags:
	- cli
---


The `temporal operator search-attribute remove` command removes custom [Search Attribute](/visibility#search-attribute) metadata.
This command does not remove custom Search Attributes from Elasticsearch.
The index schema is not modified.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`temporal operator search-attribute remove [command options] [arguments]`

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

**--name**
Search attribute name.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

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

**--yes**
Alias: **-y**
Confirm all prompts.

