---
id: execute
title: temporal workflow execute
sidebar_label: execute
description: Temporal CLI operation for ....
tags:
	- cli
---

### execute

Start a new [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) and prints its progress.

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
Optional Cron Schedule for the Workflow. Cron spec is formatted as:
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
Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks. (default: 0)

**--fields**
Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--id-reuse-policy**
Allows the same Workflow Id to be used in a new Workflow Execution. Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.

**--input**
Alias: **-i**
Optional JSON input to provide to the Workflow.
Pass "null" for null values.

**--input-file**
Passes optional input for the Workflow from a JSON file.
If there are multiple JSON files, concatenate them and separate by space or newline.
Input from the command line will overwrite file input.

**--limit**
Number of items to print. (default: 0)

**--max-field-length**
Maximum length for each attribute field. (default: 0)

**--memo**
Passes a memo in key=value format. Use valid JSON formats for value.

**--memo-file**
Passes a memo as file input, with each line following key=value format. Use valid JSON formats for value.

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

**--run-timeout**
Timeout (in seconds) of a single Workflow run. (default: 0)

**--search-attribute**
Passes Search Attribute in key=value format. Use valid JSON formats for value.

**--task-queue**
Alias: **-t**
Task Queue

**--task-timeout**
Start-to-close timeout for a Workflow Task (in seconds). (default: 10)

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

**--type**
Workflow type name.

**--workflow-id**
Alias: **-w**
Workflow Id

