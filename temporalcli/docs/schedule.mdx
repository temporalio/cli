---
id: schedule
title: temporal schedule
sidebar_label: temporal schedule
description: Temporal's Schedule commands allow users to create, update, and manage Workflow Executions seamlessly for automation, supporting commands for creation, backfill, deletion, and more.
toc_max_heading_level: 4
keywords:
  - backfill
  - cli reference
  - command-line-interface-cli
  - schedule
  - schedule backfill
  - schedule create
  - schedule delete
  - schedule describe
  - schedule list
  - schedule toggle
  - schedule trigger
  - schedule update
  - temporal cli
  - updates
tags:
  - backfill
  - cli-reference
  - command-line-interface-cli
  - schedule
  - schedule-backfill
  - schedule-create
  - schedule-delete
  - schedule-describe
  - schedule-list
  - schedule-toggle
  - schedule-trigger
  - schedule-update
  - temporal-cli
  - updates
---

## backfill

Batch-execute actions that would have run during a specified time interval.
Use this command to fill in Workflow runs from when a Schedule was paused,
before a Schedule was created, from the future, or to re-process a previously
executed interval.

Backfills require a Schedule ID and the time period covered by the request.
It's best to use the `BufferAll` or `AllowAll` policies to avoid conflicts
and ensure no Workflow Executions are skipped.

For example:

```
temporal schedule backfill \
    --schedule-id "YourScheduleId" \
    --start-time "2022-05-01T00:00:00Z" \
    --end-time "2022-05-31T23:59:59Z" \
    --overlap-policy BufferAll
```

The policies include:

* **AllowAll**: Allow unlimited concurrent Workflow Executions. This
  significantly speeds up the backfilling process on systems that support
  concurrency. You must ensure running Workflow Executions do not interfere
  with each other.
* **BufferAll**: Buffer all incoming Workflow Executions while waiting for
  the running Workflow Execution to complete.
* **Skip**: If a previous Workflow Execution is still running, discard new
  Workflow Executions.
* **BufferOne**: Same as 'Skip' but buffer a single Workflow Execution to be
  run after the previous Execution completes. Discard other Workflow
  Executions.
* **CancelOther**: Cancel the running Workflow Execution and replace it with
  the incoming new Workflow Execution.
* **TerminateOther**: Terminate the running Workflow Execution and replace
  it with the incoming new Workflow Execution.

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## end-time

Backfill end time.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## output

Non-logging data output format.

## start-time

Backfill start time.

## time-format

Time format.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## create

Create a new Schedule on the Temporal Service. A Schedule automatically starts
new Workflow Executions at the times you specify.

For example:

```
  temporal schedule create \
    --schedule-id "YourScheduleId" \
    --calendar '{"dayOfWeek":"Fri","hour":"3","minute":"30"}' \
    --workflow-id YourBaseWorkflowIdName \
    --task-queue YourTaskQueue \
    --type YourWorkflowType
```

Schedules support any combination of `--calendar`, `--interval`, and `--cron`:

* Shorthand `--interval` strings.
  For example: 45m (every 45 minutes) or 6h/5h (every 6 hours, at the top of
  the 5th hour).
* JSON `--calendar`, as in the preceding example.
* Unix-style `--cron` strings and robfig declarations
  (@daily/@weekly/@every X/etc).
  For example, every Friday at 12:30 PM: `30 12 * * Fri`.

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## calendar

Calendar specification in JSON. For example: `{"dayOfWeek":"Fri","hour":"17","minute":"5"}`.

## catchup-window

Maximum catch-up time for when the Service is unavailable.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## cron

Calendar specification in cron string format. For example: `"30 12 * * Fri"`.

## end-time

Schedule end time.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## execution-timeout

Fail a WorkflowExecution if it lasts longer than `DURATION`. This time-out includes retries and ContinueAsNew tasks.

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## input

Input value. Use JSON content or set --input-meta to override. Can't be combined with --input-file. Can be passed multiple times to pass multiple arguments.

## input-base64

Assume inputs are base64-encoded and attempt to decode them.

## input-file

A path or paths for input file(s). Use JSON content or set --input-meta to override. Can't be combined with --input. Can be passed multiple times to pass multiple arguments.

## input-meta

Input payload metadata as a `KEY=VALUE` pair. When the KEY is "encoding", this overrides the default ("json/plain"). Can be passed multiple times.

## interval

Interval duration. For example, 90m, or 60m/15m to include phase offset.

## jitter

Max difference in time from the specification. Vary the start time randomly within this amount.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## memo

Memo using 'KEY="VALUE"' pairs. Use JSON values.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## notes

Initial notes field value.

## output

Non-logging data output format.

## overlap-policy

Policy for handling overlapping Workflow Executions.

## pause-on-failure

Pause schedule after Workflow failures.

## paused

Pause the Schedule immediately on creation.

## remaining-actions

Total allowed actions. Default is zero (unlimited).

## run-timeout

Fail a Workflow Run if it lasts longer than `DURATION`.

## schedule-id

Schedule ID.

## schedule-memo

Set schedule memo using `KEY="VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={"your": "value"}'. Can be passed multiple times.

## schedule-search-attribute

Set schedule Search Attributes using `KEY="VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={"your": "value"}'. Can be passed multiple times.

## search-attribute

Search Attribute in `KEY=VALUE` format. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={"your": "value"}'. Can be passed multiple times.

## start-time

Schedule start time.

## task-queue

Workflow Task queue.

## task-timeout

Fail a Workflow Task if it lasts longer than `DURATION`. This is the Start-to-close timeout for a Workflow Task.

## time-format

Time format.

## time-zone

Interpret calendar specs with the `TZ` time zone. For a list of time zones, see: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## type

Workflow Type name.

## workflow-id

Workflow ID. If not supplied, the Service generates a unique ID.

## delete

Deletes a Schedule on the front end Service:

```
temporal schedule delete \
    --schedule-id YourScheduleId
```

Removing a Schedule won't affect the Workflow Executions it started that are
still running. To cancel or terminate these Workflow Executions, use `temporal
workflow delete` with the `TemporalScheduledById` Search Attribute instead.

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## output

Non-logging data output format.

## schedule-id

Schedule ID.

## time-format

Time format.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## describe

Show a Schedule configuration, including information about past, current, and
future Workflow runs:

```
temporal schedule describe \
    --schedule-id YourScheduleId
```

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## output

Non-logging data output format.

## schedule-id

Schedule ID.

## time-format

Time format.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## list

Lists the Schedules hosted by a Namespace:

```
temporal schedule list \
    --namespace YourNamespace
```

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## long

Show detailed information.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## output

Non-logging data output format.

## query

Filter results using given List Filter.

## really-long

Show extensive information in non-table form.

## time-format

Time format.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## toggle

Pause or unpause a Schedule by passing a flag with your desired state:

```
temporal schedule toggle \
    --schedule-id "YourScheduleId" \
    --pause \
    --reason "YourReason"
```

and

```
temporal schedule toggle
    --schedule-id "YourScheduleId" \
    --unpause \
    --reason "YourReason"
```

The `--reason` text updates the Schedule's `notes` field for operations
communication. It defaults to "(no reason provided)" if omitted. This field is
also visible on the Service Web UI.

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## output

Non-logging data output format.

## pause

Pause the Schedule.

## reason

Reason for pausing or unpausing the Schedule.

## schedule-id

Schedule ID.

## time-format

Time format.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## unpause

Unpause the Schedule.

## trigger

Trigger a Schedule to run immediately:

```
temporal schedule trigger \
    --schedule-id "YourScheduleId"
```

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## output

Non-logging data output format.

## overlap-policy

Policy for handling overlapping Workflow Executions.

## schedule-id

Schedule ID.

## time-format

Time format.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## update

Update an existing Schedule with new configuration details, including time
specifications, action, and policies:

```
temporal schedule update \
    --schedule-id "YourScheduleId" \
    --workflow-type "NewWorkflowType"
```

Use the following options to change the behavior of this command.

## address

Temporal Service gRPC endpoint.

## api-key

API key for request.

## calendar

Calendar specification in JSON. For example: `{"dayOfWeek":"Fri","hour":"17","minute":"5"}`.

## catchup-window

Maximum catch-up time for when the Service is unavailable.

## codec-auth

Authorization header for Codec Server requests.

## codec-endpoint

Remote Codec Server endpoint.

## color

Output coloring.

## command-timeout

Timeout for the span of a command.

## cron

Calendar specification in cron string format. For example: `"30 12 * * Fri"`.

## end-time

Schedule end time.

## env

Active environment name (`ENV`).

## env-file

Path to environment settings file. (defaults to `$HOME/.config/temporalio/temporal.yaml`).

## execution-timeout

Fail a WorkflowExecution if it lasts longer than `DURATION`. This time-out includes retries and ContinueAsNew tasks.

## grpc-meta

HTTP headers for requests. format as a `KEY=VALUE` pair May be passed multiple times to set multiple headers.

## input

Input value. Use JSON content or set --input-meta to override. Can't be combined with --input-file. Can be passed multiple times to pass multiple arguments.

## input-base64

Assume inputs are base64-encoded and attempt to decode them.

## input-file

A path or paths for input file(s). Use JSON content or set --input-meta to override. Can't be combined with --input. Can be passed multiple times to pass multiple arguments.

## input-meta

Input payload metadata as a `KEY=VALUE` pair. When the KEY is "encoding", this overrides the default ("json/plain"). Can be passed multiple times.

## interval

Interval duration. For example, 90m, or 60m/15m to include phase offset.

## jitter

Max difference in time from the specification. Vary the start time randomly within this amount.

## log-format

Log format. Options are: text, json.

## log-level

Log level. Default is "info" for most commands and "warn" for `server start-dev`.

## memo

Memo using 'KEY="VALUE"' pairs. Use JSON values.

## namespace

Temporal Service Namespace.

## no-json-shorthand-payloads

Raw payload output, even if they are JSON.

## notes

Initial notes field value.

## output

Non-logging data output format.

## overlap-policy

Policy for handling overlapping Workflow Executions.

## pause-on-failure

Pause schedule after Workflow failures.

## paused

Pause the Schedule immediately on creation.

## remaining-actions

Total allowed actions. Default is zero (unlimited).

## run-timeout

Fail a Workflow Run if it lasts longer than `DURATION`.

## schedule-id

Schedule ID.

## schedule-memo

Set schedule memo using `KEY="VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={"your": "value"}'. Can be passed multiple times.

## schedule-search-attribute

Set schedule Search Attributes using `KEY="VALUE` pairs. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={"your": "value"}'. Can be passed multiple times.

## search-attribute

Search Attribute in `KEY=VALUE` format. Keys must be identifiers, and values must be JSON values. For example: 'YourKey={"your": "value"}'. Can be passed multiple times.

## start-time

Schedule start time.

## task-queue

Workflow Task queue.

## task-timeout

Fail a Workflow Task if it lasts longer than `DURATION`. This is the Start-to-close timeout for a Workflow Task.

## time-format

Time format.

## time-zone

Interpret calendar specs with the `TZ` time zone. For a list of time zones, see: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.

## tls

Enable base TLS encryption. Does not have additional options like mTLS or client certs.

## tls-ca-data

Data for server CA certificate. Can't be used with --tls-ca-path.

## tls-ca-path

Path to server CA certificate. Can't be used with --tls-ca-data.

## tls-cert-data

Data for x509 certificate. Can't be used with --tls-cert-path.

## tls-cert-path

Path to x509 certificate. Can't be used with --tls-cert-data.

## tls-disable-host-verification

Disable TLS host-name verification.

## tls-key-data

Private certificate key data. Can't be used with --tls-key-path.

## tls-key-path

Path to x509 private key. Can't be used with --tls-key-data.

## tls-server-name

Override target TLS server name.

## type

Workflow Type name.

## workflow-id

Workflow ID. If not supplied, the Service generates a unique ID.

