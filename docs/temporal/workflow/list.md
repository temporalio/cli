---
id: [temporal workflow list]
title: temporal temporal workflow list
sidebar_label: list
description: List Workflow Executions based on a Query.
tags:
	- cli reference
	- temporal cli
	- [temporal workflow list]
---

The `temporal workflow list` command provides a list of [Workflow Executions](/concepts/what-is-a-workflow-execution)
that meet the criteria of a given [Query](/concepts/what-is-a-query).
By default, this command returns up to 10 closed Workflow Executions.

`temporal workflow list --query=MyQuery`

The command can also return a list of archived Workflow Executions.

`temporal workflow list --archived`

Use the command options below to change the information returned by this command.

`temporal temporal workflow list`

Use the following command options to change the information returned by this command.



- [query](/cli/cmd-options/query)

- [archived](/cli/cmd-options/archived)

- [limit](/cli/cmd-options/limit)


