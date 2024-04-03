---
id: [temporal workflow query]
title: temporal temporal workflow query
sidebar_label: query
description: Query a Workflow Execution.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow query]
---

The `temporal workflow query` command is used to [Query](/concepts/what-is-a-query) a
[Workflow Execution](/concepts/what-is-a-workflow-execution)
by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow query \
		--workflow-id MyWorkflowId \
		--name MyQuery \
		--input '{"MyInputKey": "MyInputValue"}'
```

Use the options listed below to change the command's behavior.

`temporal temporal workflow query`

Use the following command options to change the information returned by this command.



- [type](/cli/cmd-options/type)

- [reject-condition](/cli/cmd-options/reject-condition)


