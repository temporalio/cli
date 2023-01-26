---
id: update
title: temporal schedule update
sidebar_label: update
description: Updates a schedule with a new definition (full replacement, not patch).
tags:
	- cli
---


The `temporal schedule update` command updates an existing [Schedule](/workflows#schedule).

Like `temporal schedule create`, updated Schedules need to follow a certain format:
```
temporal schedule update 			\
--sid 'your-schedule-id' 	\
--cron '3 11 * * Fri' 		\
--wid 'your-workflow-id' 	\
--tq 'your-task-queue' 		\
--type 'YourWorkflowType'
```

Updating a Schedule takes the given options and replaces the entire configuration of the Schedule with what's provided.
If you only change one value of the Schedule, be sure to provide the other unchanged fields to prevent them from being overwritten.

Use the options provided below to change the command's behavior.

## OPTIONS

- [--address](/cmd-options/address)

- [--calendar](/cmd-options/calendar)

- [--catchup-window](/cmd-options/catchup-window)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--cron](/cmd-options/cron)

- [--end-time](/cmd-options/end-time)

- [--env](/cmd-options/env)

- [--execution-timeout](/cmd-options/execution-timeout)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--input](/cmd-options/input)
Pass "null" for null values.

- [--input-file](/cmd-options/input-file)
If there are multiple JSON files, concatenate them and separate by space or newline.
Input from the command line will overwrite file input.

- [--interval](/cmd-options/interval)

- [--jitter](/cmd-options/jitter)

- [--max-field-length](/cmd-options/max-field-length)

- [--memo](/cmd-options/memo)

- [--memo-file](/cmd-options/memo-file)

- [--namespace](/cmd-options/namespace)

- [--notes](/cmd-options/notes)

- [--overlap-policy](/cmd-options/overlap-policy)

- [--pause](/cmd-options/pause)

- [--pause-on-failure](/cmd-options/pause-on-failure)

- [--remaining-actions](/cmd-options/remaining-actions)

- [--run-timeout](/cmd-options/run-timeout)

- [--schedule-id](/cmd-options/schedule-id)

- [--search-attribute](/cmd-options/search-attribute)

- [--start-time](/cmd-options/start-time)

- [--task-queue](/cmd-options/task-queue)

- [--task-timeout](/cmd-options/task-timeout)

- [--time-zone](/cmd-options/time-zone)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

- [--workflow-id](/cmd-options/workflow-id)

- [--workflow-type](/cmd-options/workflow-type)

