---
id: [temporal operator namespace create]
title: temporal temporal operator namespace create
sidebar_label: create
description: Registers a new Namespace.
tags:
	- cli reference
	- temporal cli
	- [temporal operator namespace create]
---

The temporal operator namespace create command creates a new Namespace on the Server.
Namespaces can be created on the active Cluster, or any named Cluster.
`temporal operator namespace create --cluster=MyCluster -n example-1`

Global Namespaces can also be created.
`temporal operator namespace create --global -n example-2`

Other settings, such as retention and Visibility Archival State, can be configured as needed.
For example, the Visibility Archive can be set on a separate URI.
`temporal operator namespace create --retention=5 --visibility-archival-state=enabled --visibility-uri=some-uri -n example-3`

`temporal temporal operator namespace create`

Use the following command options to change the information returned by this command.



- [active-cluster](/cli/cmd-options/active-cluster)

- [cluster](/cli/cmd-options/cluster)

- [data](/cli/cmd-options/data)

- [description](/cli/cmd-options/description)

- [email](/cli/cmd-options/email)

- [global](/cli/cmd-options/global)

- [history-archival-state](/cli/cmd-options/history-archival-state)

- [history-uri](/cli/cmd-options/history-uri)

- [retention](/cli/cmd-options/retention)

- [visibility-archival-state](/cli/cmd-options/visibility-archival-state)

- [visibility-uri](/cli/cmd-options/visibility-uri)


