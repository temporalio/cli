---
id: [temporal workflow reset]
title: temporal temporal workflow reset
sidebar_label: reset
description: Resets a Workflow Execution by Event ID or reset type.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow reset]
---

The temporal workflow reset command resets a [Workflow Execution](/concepts/what-is-a-workflow-execution).
A reset allows the Workflow to resume from a certain point without losing its parameters or [Event History](/concepts/what-is-an-event-history).

The Workflow Execution can be set to a given [Event Type](/concepts/what-is-an-event):
```
temporal workflow reset --workflow-id=meaningful-business-id --type=LastContinuedAsNew
```

...or a specific any Event after `WorkflowTaskStarted`.
```
temporal workflow reset --workflow-id=meaningful-business-id --event-id=MyLastEvent
```
For batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow Id, run Id and event Id
should not be set.
Use the options listed below to change reset behavior.

`temporal temporal workflow reset`

Use the following command options to change the information returned by this command.



- [workflow-id](/cli/cmd-options/workflow-id)

- [run-id](/cli/cmd-options/run-id)

- [event-id](/cli/cmd-options/event-id)

- [reason](/cli/cmd-options/reason)

- [reapply-type](/cli/cmd-options/reapply-type)

- [type](/cli/cmd-options/type)

- [build-id](/cli/cmd-options/build-id)

- [query](/cli/cmd-options/query)

- [yes](/cli/cmd-options/yes)


