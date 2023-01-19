---
id: execute
title: temporal workflow execute
sidebar_label: execute
description: Temporal CLI operation for ....
tags:
	- cli
---

### execute

Start a new Workflow Execution and prints its progress.

Single quotes('') are used to wrap input as JSON.

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

**--cron**
Optional cron schedule for the Workflow. Cron spec is as following:
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--env**
Name of the environment to read environmental variables from. (default: default)

**--execution-timeout**
Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--fields**
customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--id-reuse-policy**
Configure if the same Workflow Id is allowed for use in new Workflow Execution. Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning

**--input**
Alias: **-i**
Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**
Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--limit**
number of items to print (default: 0)

**--max-field-length**
Maximum length for each attribute field (default: 0)

**--memo**
Pass a memo in a format key=value. Use valid JSON formats for value

**--memo-file**
Pass a memo from a file, where each line follows the format key=value. Use valid JSON formats for value

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

**--run-timeout**
Timeout (in seconds) of a single Workflow run. (default: 0)

**--search-attribute**
Pass Search Attribute in a format key=value. Use valid JSON formats for value

**--task-queue**
Alias: **-t**
Task Queue

**--task-timeout**
Workflow task start to close timeout (seconds) (default: 10)

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

**--type**
Workflow type name.

**--workflow-id**
Alias: **-w**
Workflow Id

