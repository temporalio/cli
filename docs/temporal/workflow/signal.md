---
id: [temporal workflow signal]
title: temporal temporal workflow signal
sidebar_label: signal
description: Signal Workflow Execution by Id.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow signal]
---

The `temporal workflow signal` command is used to [Signal](/concepts/what-is-a-signal) a
[Workflow Execution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow signal \
		--workflow-id MyWorkflowId \
		--name MySignal \
		--input '{"MyInputKey": "MyInputValue"}'
```

Use the options listed below to change the command's behavior.

`temporal temporal workflow signal`

Use the following command options to change the information returned by this command.



- [name](/cli/cmd-options/name)



- [workflow-id](/cli/cmd-options/workflow-id)

- [run-id](/cli/cmd-options/run-id)

- [query](/cli/cmd-options/query)

- [reason](/cli/cmd-options/reason)

- [yes](/cli/cmd-options/yes)


