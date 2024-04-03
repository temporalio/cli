---
id: [temporal workflow describe]
title: temporal temporal workflow describe
sidebar_label: describe
description: Show information about a Workflow Execution.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow describe]
---

The `temporal workflow describe` command shows information about a given
[Workflow Execution](/concepts/what-is-a-workflow-execution).

This information can be used to locate Workflow Executions that weren't able to run successfully.

`temporal workflow describe --workflow-id=meaningful-business-id`

Output can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.

`temporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true`

Use the command options below to change the information returned by this command.

`temporal temporal workflow describe`

Use the following command options to change the information returned by this command.



- [workflow-id](/cli/cmd-options/workflow-id)

- [run-id](/cli/cmd-options/run-id)



- [reset-points](/cli/cmd-options/reset-points)

- [raw](/cli/cmd-options/raw)


