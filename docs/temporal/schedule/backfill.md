---
id: [temporal schedule backfill]
title: temporal temporal schedule backfill
sidebar_label: backfill
description: Backfills a past time range of actions.
tags:
	- cli reference
	- temporal cli
	- [temporal schedule backfill]
---

The `temporal schedule backfill` command runs the Actions that would have been run in a given time
interval, all at once.

 You can use backfill to fill in Workflow Runs from a time period when the Schedule was paused, from
before the Schedule was created, from the future, or to re-process an interval that was processed.

Schedule backfills require a Schedule ID, along with the time in which to run the Schedule. You can
optionally override the overlap policy. It usually only makes sense to run backfills with either
`BufferAll` or `AllowAll` (other policies will only let one or two runs actually happen).

Example:

```
  temporal schedule backfill           \
    --schedule-id 'your-schedule-id'   \
    --overlap-policy BufferAll         \
    --start-time 2022-05-01T00:00:00Z  \
    --end-time   2022-05-31T23:59:59Z
```

`temporal temporal schedule backfill`

Use the following command options to change the information returned by this command.



- [overlap-policy](/cli/cmd-options/overlap-policy)



- [schedule-id](/cli/cmd-options/schedule-id)



- [end-time](/cli/cmd-options/end-time)

- [start-time](/cli/cmd-options/start-time)


