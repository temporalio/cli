---
id: [temporal schedule toggle]
title: temporal temporal schedule toggle
sidebar_label: toggle
description: Pauses or unpauses a Schedule.
tags:
	- cli reference
	- temporal cli
	- [temporal schedule toggle]
---

The `temporal schedule toggle` command can pause and unpause a Schedule.

Toggling a Schedule takes a reason. The reason will be set as the `notes` field of the Schedule,
to help with operations communication.

Examples:

* `temporal schedule toggle --schedule-id 'your-schedule-id' --pause --reason "paused because the database is down"`
* `temporal schedule toggle --schedule-id 'your-schedule-id' --unpause --reason "the database is back up"`

`temporal temporal schedule toggle`

Use the following command options to change the information returned by this command.



- [pause](/cli/cmd-options/pause)

- [reason](/cli/cmd-options/reason)

- [unpause](/cli/cmd-options/unpause)


