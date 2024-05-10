---
id: [temporal task-queue get-build-id-reachability]
title: temporal temporal task-queue get-build-id-reachability
sidebar_label: get-build-id-reachability
description: Retrieves information about the reachability of Build IDs on one or more Task Queues.
tags:
	- cli reference
	- temporal cli
	- [temporal task-queue get-build-id-reachability]
---

This command can tell you whether or not Build IDs may be used for new, existing, or closed workflows. Both the '--build-id' and '--task-queue' flags may be specified multiple times. If you do not provide a task queue, reachability for the provided Build IDs will be checked against all task queues.

`temporal temporal task-queue get-build-id-reachability`

Use the following command options to change the information returned by this command.



- [build-id](/cli/cmd-options/build-id)

- [reachability-type](/cli/cmd-options/reachability-type)

- [task-queue](/cli/cmd-options/task-queue)


