---
id:
title:
sidebar_label:
description:
tags:
---


### namespace

Operations applying to Namespaces.

#### describe

Describe a Namespace by its name or Id.

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

**--namespace-id**
Namespace Id

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

#### list

List all Namespaces.

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

#### register

Registers a new Namespace.

**--active-cluster**
Active cluster name

**--address**
host:port for Temporal frontend service

**--cluster**
Cluster name

**--codec-auth**
Authorization header to set for requests to Codec Server

**--codec-endpoint**
Remote Codec Server Endpoint

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
Optional timeout for context of RPC call in seconds (default: 5)

**--data**
Namespace data in a format key=value

**--description**
Namespace description

**--email**
Owner email

**--env**
Env name to read the client environment variables from (default: default)

**--global**
Flag to indicate whether namespace is a global namespace

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--history-archival-state**
Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**
Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace**
Alias: **-n**
Temporal workflow namespace (default: default)

**--retention**
Workflow Execution retention

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

**--visibility-archival-state**
Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**
Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### update

Updates a Namespace.

**--active-cluster**
Active cluster name

**--address**
host:port for Temporal frontend service

**--cluster**
Cluster name

**--codec-auth**
Authorization header to set for requests to Codec Server

**--codec-endpoint**
Remote Codec Server Endpoint

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
Optional timeout for context of RPC call in seconds (default: 5)

**--data**
Namespace data in a format key=value

**--description**
Namespace description

**--email**
Owner email

**--env**
Env name to read the client environment variables from (default: default)

**--grpc-meta**
gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--history-archival-state**
Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**
Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace**
Alias: **-n**
Temporal workflow namespace (default: default)

**--promote-global**
Promote local namespace to global namespace

**--reason**
Reason for the operation

**--retention**
Workflow Execution retention

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

**--visibility-archival-state**
Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**
Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### delete

Deletes an existing Namespace.

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

**--yes**
Alias: **-y**
Confirm all prompts.

