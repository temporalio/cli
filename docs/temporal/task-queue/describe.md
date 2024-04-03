---
id: [temporal task-queue describe]
title: temporal temporal task-queue describe
sidebar_label: describe
description: Provides information for Workers that have recently polled on this Task Queue.
tags:
	- cli reference
	- temporal cli
	- [temporal task-queue describe]
---

The `temporal task-queue describe` command provides [poller](/application-development/worker-performance#poller-count)
information for a given [Task Queue](/concepts/what-is-a-task-queue).

The [Server](/concepts/what-is-the-temporal-server) records the last time of each poll request. A `LastAccessTime` value
in excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the
Worker has shut down. [Workers](/concepts/what-is-a-worker) are removed if 5 minutes have passed since the last poll
request.

Information about the Task Queue can be returned to troubleshoot server issues.

`temporal task-queue describe --task-queue=MyTaskQueue --task-queue-type="activity"`

Use the options listed below to modify what this command returns.

`temporal temporal task-queue describe`

Use the following command options to change the information returned by this command.



- [task-queue](/cli/cmd-options/task-queue)

- [task-queue-type](/cli/cmd-options/task-queue-type)

- [partitions](/cli/cmd-options/partitions)


