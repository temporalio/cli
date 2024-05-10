---
id: [temporal env set]
title: temporal temporal env set
sidebar_label: set
description: Set environment properties.
tags:
	- cli reference
	- temporal cli
	- [temporal env set]
---

`temporal env set --env environment -k property -v value`

Property names match CLI option names, for example '--address' and '--tls-cert-path':

`temporal env set --env prod -k address -v 127.0.0.1:7233`
`temporal env set --env prod -k tls-cert-path -v /home/my-user/certs/cluster.cert`

If the environment is not specified, the `default` environment is used.

`temporal temporal env set`

Use the following command options to change the information returned by this command.



- [key](/cli/cmd-options/key)

- [value](/cli/cmd-options/value)


