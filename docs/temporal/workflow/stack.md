---
id: [temporal workflow stack]
title: temporal temporal workflow stack
sidebar_label: stack
description: Query a Workflow Execution for its stack trace.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow stack]
---

The `temporal workflow stack` command [Queries](/concepts/what-is-a-query) a
[Workflow Execution](/concepts/what-is-a-workflow-execution) with `__stack_trace` as the query type.
This returns a stack trace of all the threads or routines currently used by the workflow, and is
useful for troubleshooting.

```
temporal workflow stack --workflow-id MyWorkflowId
```

Use the options listed below to change the command's behavior.

`temporal temporal workflow stack`

Use the following command options to change the information returned by this command.



- [reject-condition](/cli/cmd-options/reject-condition)


