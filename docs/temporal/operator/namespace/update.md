---
id: [temporal operator namespace update]
title: temporal temporal operator namespace update
sidebar_label: update
description: Updates a Namespace.
tags:
	- cli reference
	- temporal cli
	- [temporal operator namespace update]
---

The temporal operator namespace update command updates a Namespace.

Namespaces can be assigned a different active Cluster.
`temporal operator namespace update -n namespace --active-cluster=NewActiveCluster`

Namespaces can also be promoted to global Namespaces.
`temporal operator namespace update -n namespace --promote-global`

Any Archives that were previously enabled or disabled can be changed through this command.
However, URI values for archival states cannot be changed after the states are enabled.
`temporal operator namespace update -n namespace --history-archival-state=enabled --visibility-archival-state=disabled`

`temporal temporal operator namespace update`

Use the following command options to change the information returned by this command.



- [active-cluster](/cli/cmd-options/active-cluster)

- [cluster](/cli/cmd-options/cluster)

- [data](/cli/cmd-options/data)

- [description](/cli/cmd-options/description)

- [email](/cli/cmd-options/email)

- [promote-global](/cli/cmd-options/promote-global)

- [history-archival-state](/cli/cmd-options/history-archival-state)

- [history-uri](/cli/cmd-options/history-uri)

- [retention](/cli/cmd-options/retention)

- [visibility-archival-state](/cli/cmd-options/visibility-archival-state)

- [visibility-uri](/cli/cmd-options/visibility-uri)


