---
id: stack
title: temporal workflow stack
sidebar_label: stack
description: Query a Workflow Execution with __stack_trace as the query type.
tags:
	- cli
---


The `temporal workflow stack` command queries a [Workflow Execution](/workflows#workflow-execution) with `--stack-trace` as the [Query](/workflows#stack-trace-query) type.
Returning the stack trace of all the threads owned by a Workflow Execution can be great for troubleshooting in production.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`temporal workflow stack [command options] [arguments]`

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

- [--workflow-id](/cmd-options/workflow-id)

