---
id:
title:
sidebar_label:
description:
tags:
---


### search-attribute

Operations applying to Search Attributes.

#### create

Adds one or more custom Search Attributes.

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

**--name**
Search attribute name.

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

**--type**
Search attribute type: [Text Keyword Int Double Bool Datetime KeywordList]

**--yes**
Alias: **-y**
Confirm all prompts.

#### list

Lists all Search Attributes that can be used in list Workflow Queries.

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

**--output**
Alias: **-o**
format output as: table, json, card. (default: table)

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

#### remove

Removes custom search attribute metadata only (Elasticsearch index schema is not modified).

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

**--name**
Search attribute name.

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

