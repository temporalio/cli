---
id: trigger
title: temporal schedule trigger
sidebar_label: trigger
description: Triggers an immediate action.
tags:
	- cli
---


The `temporal schedule trigger` command triggers an immediate action with a given [Schedule](/workflows#schedule).
By default, this action is subject to the Overlap Policy of the Schedule.

`temporal schedule trigger` can be used to start a Workflow Run immediately.
`temporal schedule trigger --sid 'your-schedule-id'`

The Overlap Policy of the Schedule can be overridden as well.
`temporal schedule trigger --sid 'your-schedule-id' --overlap-policy 'AllowAll'`

Use the options provided below to change this command's behavior.

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--env](/cmd-options/env)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--namespace](/cmd-options/namespace)

- [--overlap-policy](/cmd-options/overlap-policy)

- [--schedule-id](/cmd-options/schedule-id)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

