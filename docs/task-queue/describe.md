---
id: describe
title: temporal task-queue describe
sidebar_label: describe
description: Describes the Workers that have recently polled on this Task Queue.
tags:
	- cli
---


The `temporal task-queue describe` command provides [poller](/applcation-development/worker-performance#poller-count) information for a given [Task Queue](/tasks#task-queue).

The [Server](/clusters#temporal-server) records the last time of each poll request.
Should `LastAccessTime` exceeds one minute, it's likely that the Worker is at capacity (all Workflow and Activity slots are full) or that the Worker has shut down.
[Workers](/workers) are removed if 5 minutes have passed since the last poll request.

Use the options listed below to modify what this command returns.
Make sure to write the command as follows:
`temporal task-queue describe [command options] [arguments]`

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--env](/cmd-options/env)

- [--fields](/cmd-options/fields)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--namespace](/cmd-options/namespace)

- [--output](/cmd-options/output)

- [--task-queue](/cmd-options/task-queue)

- [--task-queue-type](/cmd-options/task-queue-type)

- [--time-format](/cmd-options/time-format)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

