
Commands for managing Temporal server

### start-dev

Start Temporal development server

**--config, -c**="": config dir path

**--db-filename, -f**="": File in which to persist Temporal state

**--dynamic-config-value**="": dynamic config value, as KEY=JSON_VALUE (meaning strings need quotes)

**--headless**: disable the temporal web UI

**--ip**="": IPv4 address to bind the frontend service to instead of localhost (default: 127.0.0.1)

**--log-format**="": customize the log formatting (allowed: ["json" "pretty"]) (default: json)

**--log-level**="": customize the log level (allowed: ["debug" "info" "warn" "error" "fatal"]) (default: info)

**--metrics-port**="": Port for the metrics listener (default: 0)

**--namespace, -n**="": Specify namespaces that should be pre-created. Namespace 'default' is auto created

**--port, -p**="": Port for the temporal-frontend GRPC service (default: 7233)

**--sqlite-pragma**="": specify sqlite pragma statements in pragma=value format (allowed: ["journal_mode" "synchronous"])

**--ui-asset-path**="": UI Custom Assets path

**--ui-codec-endpoint**="": UI Remote data converter HTTP endpoint

**--ui-ip**="": IPv4 address to bind the web UI to instead of localhost

**--ui-port**="": port for the temporal web UI (default: 0)

