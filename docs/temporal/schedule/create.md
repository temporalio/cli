---
id: [temporal schedule create]
title: temporal temporal schedule create
sidebar_label: create
description: Create a new Schedule.
tags:
	- cli reference
	- temporal cli
	- [temporal schedule create]
---

The `temporal schedule create` command creates a new Schedule.

Example:

```
  temporal schedule create                                    \
    --schedule-id 'your-schedule-id'                          \
    --calendar '{"dayOfWeek":"Fri","hour":"3","minute":"11"}' \
    --workflow-id 'your-base-workflow-id'                     \
    --task-queue 'your-task-queue'                            \
    --workflow-type 'YourWorkflowType'
```

Any combination of `--calendar`, `--interval`, and `--cron` is supported.
Actions will be executed at any time specified in the Schedule.

`temporal temporal schedule create`

Use the following command options to change the information returned by this command.



- [calendar](/cli/cmd-options/calendar)

- [catchup-window](/cli/cmd-options/catchup-window)

- [cron](/cli/cmd-options/cron)

- [end-time](/cli/cmd-options/end-time)

- [interval](/cli/cmd-options/interval)

- [jitter](/cli/cmd-options/jitter)

- [notes](/cli/cmd-options/notes)

- [paused](/cli/cmd-options/paused)

- [pause-on-failure](/cli/cmd-options/pause-on-failure)

- [remaining-actions](/cli/cmd-options/remaining-actions)

- [start-time](/cli/cmd-options/start-time)

- [time-zone](/cli/cmd-options/time-zone)

- [schedule-search-attribute](/cli/cmd-options/schedule-search-attribute)

- [schedule-memo](/cli/cmd-options/schedule-memo)




