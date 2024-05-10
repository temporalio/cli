---
id: [temporal workflow update]
title: temporal temporal workflow update
sidebar_label: update
description: Updates a running workflow synchronously.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow update]
---

The `temporal workflow update` command is used to synchronously [Update](/concepts/what-is-an-update) a
[WorkflowExecution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow update \
		--workflow-id MyWorkflowId \
		--name MyUpdate \
		--input '{"Input": "As-JSON"}'
```

Use the options listed below to change the command's behavior.

`temporal temporal workflow update`

Use the following command options to change the information returned by this command.



- [name](/cli/cmd-options/name)

- [workflow-id](/cli/cmd-options/workflow-id)

- [update-id](/cli/cmd-options/update-id)

- [run-id](/cli/cmd-options/run-id)

- [first-execution-run-id](/cli/cmd-options/first-execution-run-id)


