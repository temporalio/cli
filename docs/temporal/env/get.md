---
id: [temporal env get]
title: temporal temporal env get
sidebar_label: get
description: Print environment properties.
tags:
	- cli reference
	- temporal cli
	- [temporal env get]
---

`temporal env get --env environment`

Print all properties of the 'prod' environment:

`temporal env get prod`

```
tls-cert-path  /home/my-user/certs/client.cert
tls-key-path   /home/my-user/certs/client.key
address        temporal.example.com:7233
namespace      someNamespace
```

Print a single property:

`temporal env get --env prod -k tls-key-path`

```
tls-key-path  /home/my-user/certs/cluster.key
```

If the environment is not specified, the `default` environment is used.

`temporal temporal env get`

Use the following command options to change the information returned by this command.



- [key](/cli/cmd-options/key)


