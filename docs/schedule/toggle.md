---
id: toggle
title: temporal schedule toggle
sidebar_label: toggle
description: Pauses or unpauses a schedule.
tags:
	- cli
---


The `temporal schedule toggle` command can pause and unpause a [Schedule](/workflows#schedule).

Toggling a Schedule requires a reason to be entered on the command line.
Use `--reason` to note the issue leading to the pause or unpause.

Schedule toggles are passed in this format:
` temporal schedule toggle --sid 'your-schedule-id' --pause --reason "paused because the database is down"`
`temporal schedule toggle --sid 'your-schedule-id' --unpause --reason "the database is back up"`

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

- [--pause](/cmd-options/pause)

- [--reason](/cmd-options/reason)

- [--schedule-id](/cmd-options/schedule-id)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

- [--unpause](/cmd-options/unpause)

