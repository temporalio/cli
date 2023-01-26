---
id: backfill
title: temporal schedule backfill
sidebar_label: backfill
description: Backfills a past time range of actions.
tags:
	- cli
---


The `temporal schedule backfill` command executes Actions ahead of their specified time range.
Backfilling can be used to fill in [Workflow Runs](/workflows#run-id) from a time period when the Schedule was paused, or from before the Schedule was created.

```
temporal schedule backfill --sid 'your-schedule-id' \
--overlap-policy 'BufferAll' 				\
--start-time '2022-05-0101T00:00:00Z'		\
--end-time '2022-05-31T23:59:59Z'
```

Use the options provided below to change this command's behavior.

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--end-time](/cmd-options/end-time)

- [--env](/cmd-options/env)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--namespace](/cmd-options/namespace)

- [--overlap-policy](/cmd-options/overlap-policy)

- [--schedule-id](/cmd-options/schedule-id)

- [--start-time](/cmd-options/start-time)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

