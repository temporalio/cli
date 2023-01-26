---
id: list
title: temporal workflow list
sidebar_label: list
description: List Workflow Executions based on a Query.
tags:
	- cli
---


The `temporal workflow list` command provides a list of [Workflow Executions](/workflows#workflow-execution) that meet the criteria of a given [Query](/workflows#query).
By default, this command returns a list of up to 10 closed Workflow Executions.

Use the command options listed below to change the information returned by this command.
Make sure to write the command as follows:
`temporal workflow list [command options] [arguments]`

## OPTIONS

- [--address](/cmd-options/address)

- [--archived](/cmd-options/archived)
Currently an experimental feature.

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--env](/cmd-options/env)

- [--fields](/cmd-options/fields)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--limit](/cmd-options/limit)

- [--namespace](/cmd-options/namespace)

- [--no-pager](/cmd-options/no-pager)

- [--output](/cmd-options/output)

- [--pager](/cmd-options/pager)
Options: less, more, favoritePager.

- [--query](/cmd-options/query)

- [--time-format](/cmd-options/time-format)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

