---
id: [temporal workflow start]
title: temporal temporal workflow start
sidebar_label: start
description: Starts a new Workflow Execution.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow start]
---

The `temporal workflow start` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution). The
Workflow and Run IDs are returned after starting the [Workflow](/concepts/what-is-a-workflow).

```
temporal workflow start \
		--workflow-id meaningful-business-id \
		--type MyWorkflow \
		--task-queue MyTaskQueue \
		--input '{"Input": "As-JSON"}'
```

`temporal temporal workflow start`

Use the following command options to change the information returned by this command.



- [workflow-id](/cli/cmd-options/workflow-id)

- [type](/cli/cmd-options/type)

- [task-queue](/cli/cmd-options/task-queue)

- [run-timeout](/cli/cmd-options/run-timeout)

- [execution-timeout](/cli/cmd-options/execution-timeout)

- [task-timeout](/cli/cmd-options/task-timeout)

- [search-attribute](/cli/cmd-options/search-attribute)

- [memo](/cli/cmd-options/memo)



- [cron](/cli/cmd-options/cron)

- [fail-existing](/cli/cmd-options/fail-existing)

- [start-delay](/cli/cmd-options/start-delay)

- [id-reuse-policy](/cli/cmd-options/id-reuse-policy)



- [input](/cli/cmd-options/input)

- [input-file](/cli/cmd-options/input-file)

- [input-meta](/cli/cmd-options/input-meta)

- [input-base64](/cli/cmd-options/input-base64)


