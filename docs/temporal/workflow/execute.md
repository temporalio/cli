---
id: [temporal workflow execute]
title: temporal temporal workflow execute
sidebar_label: execute
description: Start a new Workflow Execution and prints its progress.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow execute]
---

The `temporal workflow execute` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution) and
prints its progress. The command completes when the Workflow Execution completes.

Single quotes('') are used to wrap input as JSON.

```
temporal workflow execute
		--workflow-id meaningful-business-id \
		--type MyWorkflow \
		--task-queue MyTaskQueue \
		--input '{"Input": "As-JSON"}'
```

`temporal temporal workflow execute`

Use the following command options to change the information returned by this command.



- [event-details](/cli/cmd-options/event-details)


