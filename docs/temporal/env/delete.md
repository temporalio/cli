---
id: [temporal env delete]
title: temporal temporal env delete
sidebar_label: delete
description: Delete an environment or environment property.
tags:
	- cli reference
	- temporal cli
	- [temporal env delete]
---

`temporal env delete --env environment [-k property]`

Delete an environment or just a single property:

`temporal env delete --env prod`
`temporal env delete --env prod -k tls-cert-path`

If the environment is not specified, the `default` environment is deleted:

`temporal env delete -k tls-cert-path`

`temporal temporal env delete`

Use the following command options to change the information returned by this command.



- [key](/cli/cmd-options/key)


