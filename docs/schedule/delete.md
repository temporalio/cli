---
id: delete
title: temporal schedule delete
sidebar_label: delete
description: Deletes a schedule.
tags:
	- cli
---


The `temporal schedule delete` command deletes a [Schedule](/workflows#schedule).
Deleting a Schedule does not affect any [Workflows](/workflows) started by the Schedule.

[Workflow Executions](/workflows#workflow-execution) started by Schedules can be cancelled or terminated like other Workflow Executions.
However, Workflow Executions started by a Schedule can be identified by their [Search Attributes](/visibility#search-attribute), making them targetable by batch command for termination.

`temporal schedule delete --sid 'your-schedule-id' [command options] [arguments]`

Use the options below to change the behavior of this command.

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--env](/cmd-options/env)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--namespace](/cmd-options/namespace)

- [--schedule-id](/cmd-options/schedule-id)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

