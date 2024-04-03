---
id: [temporal workflow cancel]
title: temporal temporal workflow cancel
sidebar_label: cancel
description: Cancel a Workflow Execution.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow cancel]
---

The `temporal workflow cancel` command is used to cancel a [Workflow Execution](/concepts/what-is-a-workflow-execution).
Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new
Command Task will be scheduled, and the Workflow Execution will perform cleanup work.

Executions may be cancelled by [ID](/concepts/what-is-a-workflow-id):
```
temporal workflow cancel --workflow-id MyWorkflowId
```

...or in bulk via a visibility query [list filter](/concepts/what-is-a-list-filter):
```
temporal workflow cancel --query=MyQuery
```

Use the options listed below to change the behavior of this command.

`temporal temporal workflow cancel`

Use the following command options to change the information returned by this command.




