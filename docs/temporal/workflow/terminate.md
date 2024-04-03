---
id: [temporal workflow terminate]
title: temporal temporal workflow terminate
sidebar_label: terminate
description: Terminate Workflow Execution by ID or List Filter.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow terminate]
---

The `temporal workflow terminate` command is used to terminate a
[Workflow Execution](/concepts/what-is-a-workflow-execution). Canceling a running Workflow Execution records a
`WorkflowExecutionTerminated` event as the closing Event in the workflow's Event History. Workflow code is oblivious to
termination. Use `temporal workflow cancel` if you need to perform cleanup in your workflow.

Executions may be terminated by [ID](/concepts/what-is-a-workflow-id) with an optional reason:
```
temporal workflow terminate [--reason my-reason] --workflow-id MyWorkflowId
```

...or in bulk via a visibility query [list filter](/concepts/what-is-a-list-filter):
```
temporal workflow terminate --query=MyQuery
```

Use the options listed below to change the behavior of this command.

`temporal temporal workflow terminate`

Use the following command options to change the information returned by this command.



- [workflow-id](/cli/cmd-options/workflow-id)

- [run-id](/cli/cmd-options/run-id)

- [query](/cli/cmd-options/query)

- [reason](/cli/cmd-options/reason)

- [yes](/cli/cmd-options/yes)


