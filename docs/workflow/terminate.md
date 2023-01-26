---
id: terminate
title: temporal workflow terminate
sidebar_label: terminate
description: Terminate Workflow Execution by Id or List Filter.
tags:
	- cli
---


The `temporal workflow terminate` command terminates a [Workflow Execution](/workflows#workflow-execution)

Terminating a running Workflow Execution records a [`WorkflowExecutionTerminated` event](/events#workflowexecutionterminated) as the closing Event in the [Event History](/workflows#event-history).
Any further [Command](/workflows#command) Tasks cannot be scheduled after running this command.

Use the options listed below to change termination behavior.
Make sure to write the command as follows:
`temporal workflow terminate [command options] [arguments]`

## OPTIONS

- [--address](/cmd-options/address)

- [--codec-auth](/cmd-options/codec-auth)

- [--codec-endpoint](/cmd-options/codec-endpoint)

- [--color](/cmd-options/color)

- [--context-timeout](/cmd-options/context-timeout)

- [--env](/cmd-options/env)

- [--grpc-meta](/cmd-options/grpc-meta)

- [--namespace](/cmd-options/namespace)

- [--query](/cmd-options/query)

- [--reason](/cmd-options/reason)

- [--run-id](/cmd-options/run-id)

- [--tls-ca-path](/cmd-options/tls-ca-path)

- [--tls-cert-path](/cmd-options/tls-cert-path)

- [--tls-disable-host-verification](/cmd-options/tls-disable-host-verification)

- [--tls-key-path](/cmd-options/tls-key-path)

- [--tls-server-name](/cmd-options/tls-server-name)

- [--workflow-id](/cmd-options/workflow-id)

- [--yes](/cmd-options/yes)

