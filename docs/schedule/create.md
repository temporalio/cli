---
id: create
title: temporal schedule create
sidebar_label: create
description: Create a new schedule.
tags:
	- cli
---


The `temporal schedule create` command creates a new [Schedule](/workflows#schedule).
Newly created Schedules return a Schedule ID to be used in other Schedule commands.

Schedules need to follow a format like the example shown here:
```
temporal schedule create \
--sid 'your-schedule-id' \
--cron '3 11 * * Fri' \
--wid 'your-workflow-id' \
--tq 'your-task-queue' \
--type 'YourWorkflowType'
```

Any combination of `--cal`, `--interval`, and `--cron` is supported.
Actions will be executed at any time specified in the Schedule.

Use the options provided below to change the command's behavior.

## OPTIONS

**--address**
The host and port (formatted as host:port) for the Temporal Frontend Service.

**--calendar**
Calendar specification in JSON, e.g. {"dayOfWeek":"Fri","hour":"17","minute":"5"}

**--catchup-window**
Maximum allowed catch-up time if server is down.

**--codec-auth**
Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**
Endpoint for a remote Codec Server.

**--color**
when to use color: auto, always, never. (default: auto)

**--context-timeout**
An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--cron**
Calendar specification as cron string, e.g. "30 2 * * 5" or "@daily".

**--end-time**
Overall schedule end time.

**--env**
Name of the environment to read environmental variables from. (default: default)

**--execution-timeout**
Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks. (default: 0)

**--grpc-meta**
Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format.

**--input**
Alias: **-i**
Optional JSON input to provide to the Workflow.
Pass "null" for null values.

**--input-file**
Passes optional input for the Workflow from a JSON file.
If there are multiple JSON files, concatenate them and separate by space or newline.
Input from the command line will overwrite file input.

**--interval**
Interval duration, e.g. 90m, or 90m/13m to include phase offset.

**--jitter**
Jitter duration.

**--max-field-length**
Maximum length for each attribute field. (default: 0)

**--memo**
Set a memo on a schedule. Format: key=value. Use valid JSON formats for value.

**--memo-file**
Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value.

**--namespace**
Alias: **-n**
Identifies a Namespace in the Temporal Workflow. (default: default)

**--notes**
Initial value of notes field.

**--overlap-policy**
Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll.

**--pause**
Initial value of paused state.

**--pause-on-failure**
Pause schedule after any workflow failure.

**--remaining-actions**
Total number of actions allowed. (default: 0)

**--run-timeout**
Timeout (in seconds) of a single Workflow run. (default: 0)

**--schedule-id**
Alias: **-s**
Schedule Id

**--search-attribute**
Set Search Attribute on a schedule. Format: key=value. Use valid JSON formats for value.

**--start-time**
Overall schedule start time.

**--task-queue**
Alias: **-t**
Task Queue

**--task-timeout**
Start-to-close timeout for a Workflow Task (in seconds). (default: 10)

**--time-zone**
Time zone (IANA name).

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

**--workflow-id**
Alias: **-w**
Workflow Id

**--workflow-type**
Workflow type name.

