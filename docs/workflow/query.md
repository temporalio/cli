---
id: query
title: temporal workflow query
sidebar_label: query
description: Query a Workflow Execution.
tags:
	- cli
---


The `temporal workflow query` command sends a [Query](/workflows#query) to a [Workflow Execution](/workflows#workflow-execution).

Queries can retrieve all or part of the Workflow state within given parameters.
Queries can also be used on completed [Workflows](/workflows#workflow-execution).

Use the command options listed below to change the information returned by this command.
Make sure to write the command as follows:
`temporal workflow query [command options] [arguments]`

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--env](/cmd-options/env)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--input](/cmd-options/input)

- [--input-file](/cmd-options/input-file)
If there are multiple JSON, concatenate them and separate by space or newline.
Input from the command line will overwrite file input.

- [--namespace](/cmd-options/namespace)

- [--reject-condition](/cmd-options/reject-condition)

- [--run-id](/cmd-options/run-id)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

- [--type](/cmd-options/type)

- [--workflow-id](/cmd-options/workflow-id)

