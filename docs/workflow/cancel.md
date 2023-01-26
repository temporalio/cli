---
id: cancel
title: temporal workflow cancel
sidebar_label: cancel
description: Cancel a Workflow Execution.
tags:
	- cli
---


The `temporal workflow cancel` command cancels a [Workflow Execution](/workflows#workflow-execution).

Canceling a running Workflow Execution records a [`WorkflowExecutionCancelRequested` event](/events#workflowexecutioncancelrequested) in the [Event History](/workflows#event-history).
A new [Command](/workflows#command) Task will be scheduled, and the Workflow Execution performs cleanup work.

Use the options listed below to change the behavior of this command.
Make sure to write the command as follows:
`temporal workflow cancel [command options] [arguments]`

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

