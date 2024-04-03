---
id: [temporal workflow delete]
title: temporal temporal workflow delete
sidebar_label: delete
description: Deletes a Workflow Execution.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow delete]
---

The `temporal workflow delete` command is used to delete a specific [Workflow Execution](/concepts/what-is-a-workflow-execution).
This asynchronously deletes a workflow's [Event History](/concepts/what-is-an-event-history).
If the [Workflow Execution](/concepts/what-is-a-workflow-execution) is Running, it will be terminated before deletion.

```
temporal workflow delete \
		--workflow-id MyWorkflowId \
```

Use the options listed below to change the command's behavior.

`temporal temporal workflow delete`

Use the following command options to change the information returned by this command.




