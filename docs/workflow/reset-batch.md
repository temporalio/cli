---
id: reset-batch
title: temporal workflow reset-batch
sidebar_label: reset-batch
description: Reset a batch of Workflow Executions by reset type: LastWorkflowTask, LastContinuedAsNew, FirstWorkflowTask
tags:
	- cli
---


The `temporal workflow reset-batch` command resets a batch of [Workflow Executions](/workflows#workflow-execution) by `resetType`.
Resetting a [Workflow](/workflows) allows the process to resume from a certain point without losing your parameters or [Event History](/workflows#event-history).

Use the options listed below to change reset behavior.
Make sure to write the command as follows:
`temporal workflow reset-batch [command options] [arguments]`

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--dry-run](/cmd-options/dry-run)

- [--env](/cmd-options/env)

- [--exclude-file](/cmd-options/exclude-file)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--input-file](/cmd-options/input-file)

- [--input-parallelism](/cmd-options/input-parallelism)

- [--input-separator](/cmd-options/input-separator)

- [--namespace](/cmd-options/namespace)

- [--non-deterministic](/cmd-options/non-deterministic)

- [--query](/cmd-options/query)

- [--reason](/cmd-options/reason)

- [--skip-base-is-not-current](/cmd-options/skip-base-is-not-current)

- [--skip-current-open](/cmd-options/skip-current-open)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

- [--type](/cmd-options/type)

