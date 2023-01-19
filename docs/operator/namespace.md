---
id: namespace
title: temporal operator namespace
sidebar_label: namespace
description: Temporal CLI operation for ....
tags:
	- cli
---

### namespace

Operations applying to Namespaces.

#### describe

Describe a Namespace by its name or Id.

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

**--namespace-id**
Namespace Id

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

#### list

List all Namespaces.

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

#### register

Registers a new Namespace.

**--active-cluster**
Active cluster name

**--address**
The host and port (formatted as host:port) for the Temporal Frontend Service.

**--cluster**
Cluster name

**--codec-auth**
Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**
Endpoint for a remote Codec Server.

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--data**
Namespace data in a format key=value

**--description**
Namespace description

**--email**
Owner email

**--env**
Name of the environment to read environmental variables from. (default: default)

**--global**
Flag to indicate whether namespace is a global namespace

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--history-archival-state**
Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**
Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--retention**
Workflow Execution retention

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

**--visibility-archival-state**
Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**
Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### update

Updates a Namespace.

**--active-cluster**
Active cluster name

**--address**
The host and port (formatted as host:port) for the Temporal Frontend Service.

**--cluster**
Cluster name

**--codec-auth**
Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**
Endpoint for a remote Codec Server.

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--data**
Namespace data in a format key=value

**--description**
Namespace description

**--email**
Owner email

**--env**
Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--history-archival-state**
Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**
Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--promote-global**
Promote local namespace to global namespace

**--reason**
Reason for the operation

**--retention**
Workflow Execution retention

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

**--visibility-archival-state**
Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**
Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### delete

Deletes an existing Namespace.

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

