---
id: start-dev
title: temporal server start-dev
sidebar_label: start-dev
description: Start Temporal development server.
tags:
	- cli
---


**--config**
Alias: **-c**
Path to config directory.

**--db-filename**
Alias: **-f**
File in which to persist Temporal state (by default, Workflows are lost when the process dies).

**--dynamic-config-value**
Dynamic config value, as KEY=JSON_VALUE (string values need quotes).

**--headless**
Disable the Web UI.

**--ip**
IPv4 address to bind the frontend service to. (default: 127.0.0.1)

**--log-format**
Set the log formatting. Options: ["json", "pretty"]. (default: json)

**--log-level**
Set the log level. Options: ["debug" "info" "warn" "error" "fatal"]. (default: info)

**--metrics-port**
Port for /metrics. (default: 0)

**--namespace**
Alias: **-n**
Specify namespaces that should be pre-created (namespace "default" is always created).

**--port**
Alias: **-p**
Port for the frontend gRPC service. (default: 7233)

**--sqlite-pragma**
Specify sqlite pragma statements in pragma=value format. Pragma options: ["journal_mode" "synchronous"].

**--ui-asset-path**
UI Custom Assets path.

**--ui-codec-endpoint**
UI Remote data converter HTTP endpoint.

**--ui-ip**
IPv4 address to bind the Web UI to.

**--ui-port**
Port for the Web UI. (default: 0)

