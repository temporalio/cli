# NAME

temporal - Temporal command-line interface and development server

# SYNOPSIS

temporal

**Usage**:

```
temporal [GLOBAL OPTIONS] command [COMMAND OPTIONS] [ARGUMENTS...]
```

# COMMANDS

## server

Commands for managing the Temporal Server.

    Server commands allow you to start and manage the [Temporal Server](/concepts/what-is-a-temporal-server) from the command line.
    
    Currently, `cli` server functionality extends to starting the Server. 

### start-dev

Start Temporal development server.

    The `temporal server start-dev` command starts the Temporal Server on `localhost:7233`.
    The results of any command run on the Server can be viewed at http://localhost:7233.

**--config, -c**="": Path to config directory.

**--db-filename, -f**="": File in which to persist Temporal state (by default, Workflows are lost when the process dies).

**--dynamic-config-value**="": Dynamic config value, as KEY=JSON_VALUE (string values need quotes).

**--headless**: Disable the Web UI.

**--ip**="": IPv4 address to bind the frontend service to. (default: 127.0.0.1)

**--log-format**="": Set the log formatting. Options: ["json", "pretty"]. (default: json)

**--log-level**="": Set the log level. Options: ["debug" "info" "warn" "error" "fatal"]. (default: info)

**--metrics-port**="": Port for /metrics (default: 0)

**--namespace, -n**="": Specify namespaces that should be pre-created (namespace "default" is always created).

**--port, -p**="": Port for the frontend gRPC service. (default: 7233)

**--sqlite-pragma**="": Specify sqlite pragma statements in pragma=value format. Pragma options: ["journal_mode" "synchronous"].

**--ui-asset-path**="": UI Custom Assets path.

**--ui-codec-endpoint**="": UI Remote data converter HTTP endpoint.

**--ui-ip**="": IPv4 address to bind the Web UI to.

**--ui-port**="": Port for the Web UI. (default: 0)

## workflow

Operations that can be performed on Workflows.

>Workflow commands allow operations to be performed on [Workflow Executions](/concepts/what-is-a-workflow-execution).

### start

Starts a new Workflow Execution.

    The `temporal workflow start` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution).
    When invoked successfully, the Workflow and Run ID are returned immediately after starting the [Workflow](/concepts/what-is-a-workflow).
    
    Use the command options listed below to change how the Workflow Execution behaves upon starting.
    Make sure to write the command in this format:
    `temporal workflow start [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--cron**="": Optional Cron Schedule for the Workflow. Cron spec is formatted as: 
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--execution-timeout**="": Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks. (default: 0)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--id-reuse-policy**="": Allows the same Workflow Id to be used in a new Workflow Execution (AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning).

**--input, -i**="": Optional JSON input to provide to the Workflow. Pass "null" for null values.

**--input-file**="": Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input.

**--limit**="": Number of items to print. (default: 0)

**--max-field-length**="": Maximum length for each attribute field. (default: 0)

**--memo**="": Passes a memo in key=value format. Use valid JSON formats for value.

**--memo-file**="": Passes a memo as file input, with each line following key=value format. Use valid JSON formats for value.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--run-timeout**="": Timeout (in seconds) of a single Workflow run. (default: 0)

**--search-attribute**="": Passes Search Attribute in key=value format. Use valid JSON formats for value.

**--task-queue, -t**="": Task Queue

**--task-timeout**="": Start-to-close timeout for a Workflow Task (in seconds). (default: 10)

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--type**="": Workflow type name.

**--workflow-id, -w**="": Workflow Id

### execute

Start a new Workflow Execution and prints its progress.

    The `temporal workflow execute` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution) and prints its progress.
    The command doesn't finish until the [Workflow](/concepts/what-is-a-workflow) completes.
    
    Single quotes('') are used to wrap input as JSON.
    
    Use the command options listed below to change how the Workflow Execution behaves during its run.
    Make sure to write the command in this format:
    `temporal workflow execute [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--cron**="": Optional Cron Schedule for the Workflow. Cron spec is formatted as: 
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--execution-timeout**="": Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks. (default: 0)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--id-reuse-policy**="": Allows the same Workflow Id to be used in a new Workflow Execution (AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning).

**--input, -i**="": Optional JSON input to provide to the Workflow. Pass "null" for null values.

**--input-file**="": Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input.

**--limit**="": Number of items to print. (default: 0)

**--max-field-length**="": Maximum length for each attribute field. (default: 0)

**--memo**="": Passes a memo in key=value format. Use valid JSON formats for value.

**--memo-file**="": Passes a memo as file input, with each line following key=value format. Use valid JSON formats for value.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--run-timeout**="": Timeout (in seconds) of a single Workflow run. (default: 0)

**--search-attribute**="": Passes Search Attribute in key=value format. Use valid JSON formats for value.

**--task-queue, -t**="": Task Queue

**--task-timeout**="": Start-to-close timeout for a Workflow Task (in seconds). (default: 10)

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--type**="": Workflow type name.

**--workflow-id, -w**="": Workflow Id

### describe

Show information about a Workflow Execution.

    The `temporal workflow describe` command shows information about a given [Workflow Execution](/concepts/what-is-a-workflow-execution).
    This information can be used to locate Workflow Executions that weren't able to run successfully.
    
    Use the command options listed below to change the information returned by this command.
    Make sure to write the command in this format:
    `temporal workflow describe [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--raw**: Print properties as they are stored.

**--reset-points**: Only show auto-reset points.

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

### list

List Workflow Executions based on a Query.

    The `temporal workflow list` command provides a list of [Workflow Executions](/concepts/what-is-a-workflow-execution) that meet the criteria of a given [Query](/concepts/what-is-a-query).
    By default, this command returns a list of up to 10 closed Workflow Executions.
    
    Use the command options listed below to change the information returned by this command.
    Make sure to write the command as follows:
    `temporal workflow list [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--archived**: List archived Workflow Executions. Currently an experimental feature.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--limit**="": Number of items to print. (default: 0)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--query, -q**="": Filter results using an SQL-like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for more information.

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### show

Show Event History for a Workflow Execution.

    The `temporal workflow show` command provides the [Event History](/concepts/what-is-an-event-history) for a specified [Workflow Execution](/concepts/what-is-a-workflow-execution).
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal workflow show [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--follow, -f**: Follow the progress of a Workflow Execution.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--limit**="": Number of items to print. (default: 0)

**--max-field-length**="": Maximum length for each attribute field. (default: 500)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--output-filename**="": Serializes Event History to a file.

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--reset-points**: Only show Workflow Events that are eligible for reset.

**--run-id, -r**="": Run Id

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

### query

Query a Workflow Execution.

    The `temporal workflow query` command sends a [Query](/concepts/what-is-a-query) to a [Workflow Execution](/concepts/what-is-a-workflow-execution).
    
    Queries can retrieve all or part of the Workflow state within given parameters.
    Queries can also be used on completed [Workflows](/concepts/what-is-a-workflow-execution).
    
    Use the command options listed below to change the information returned by this command.
    Make sure to write the command as follows:
    `temporal workflow query [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--input, -i**="": Optional JSON input to provide to the Workflow. Pass "null" for null values.

**--input-file**="": Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--reject-condition**="": Optional flag for rejecting Queries based on Workflow state. Valid values are "not_open" and "not_completed_cleanly".

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--type**="": The Query type you want to run.

**--workflow-id, -w**="": Workflow Id

### stack

Query a Workflow Execution with __stack_trace as the query type.

    The `temporal workflow stack` command queries a [Workflow Execution](/concepts/what-is-a-workflow-execution) with `--stack-trace` as the [Query](/concepts/what-is-a-query#stack-trace-query) type.
    Returning the stack trace of all the threads owned by a Workflow Execution can be great for troubleshooting in production.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal workflow stack [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--input, -i**="": Optional JSON input to provide to the Workflow. Pass "null" for null values.

**--input-file**="": Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--reject-condition**="": Optional flag for rejecting Queries based on Workflow state. Valid values are "not_open" and "not_completed_cleanly".

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

### signal

Signal Workflow Execution by Id or List Filter.

    The `temporal workflow signal` command is used to [Signal](/concepts/what-is-a-signal) a [Workflow Execution](/concepts/what-is-a-workflow-execution) by ID or [List Filter](/concepts/what-is-a-list-filter).
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal workflow signal [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--input, -i**="": Input for the Signal (JSON).

**--input-file**="": Input for the Signal from file (JSON).

**--name**="": Signal Name

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--query, -q**="": Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.

**--reason**="": Reason to perform a given operation on the Cluster.

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Signal Workflow Execution by Id.

**--yes, -y**: Confirm all prompts.

### count

Count Workflow Executions (requires ElasticSearch to be enabled).

    The `temporal workflow count` command returns a count of [Workflow Executions](/concepts/what-is-a-workflow-execution).
    This command requires Elasticsearch to be enabled.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal workflow count [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--query, -q**="": Filter results using an SQL-like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for more information.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### cancel

Cancel a Workflow Execution.

    The `temporal workflow cancel` command cancels a [Workflow Execution](/concepts/what-is-a-workflow-execution).
    
    Canceling a running Workflow Execution records a [`WorkflowExecutionCancelRequested` event](/references/events#workflow-execution-cancel-requested) in the [Event History](/concepts/what-is-an-event-history).
    A new [Command](/concepts/what-is-a-command) Task will be scheduled, and the Workflow Execution performs cleanup work.
    
    Use the options listed below to change the behavior of this command.
    Make sure to write the command as follows:
    `temporal workflow cancel [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--query, -q**="": Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.

**--reason**="": Reason to perform a given operation on the Cluster.

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Cancel Workflow Execution by Id.

**--yes, -y**: Confirm all prompts.

### terminate

Terminate Workflow Execution by Id or List Filter.

    The `temporal workflow terminate` command terminates a [Workflow Execution](/concepts/what-is-a-workflow-execution)
    
    Terminating a running Workflow Execution records a [`WorkflowExecutionTerminated` event](/references/events#workflowexecutionterminated) as the closing Event in the [Event History](/concepts/what-is-an-event-history).
    Any further [Command](/concepts/what-is-a-command) Tasks cannot be scheduled after running this command.
    
    Use the options listed below to change termination behavior.
    Make sure to write the command as follows:
    `temporal workflow terminate [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--query, -q**="": Terminate Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/.

**--reason**="": Reason to perform a given operation on the Cluster.

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Terminate Workflow Execution by Id.

**--yes, -y**: Confirm all prompts.

### delete

Deletes a Workflow Execution.

    The `temporal workflow delete` command deletes the specified [Workflow Execution](/concepts/what-is-a-workflow-execution).
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal workflow delete [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

### reset

Resets a Workflow Execution by Event Id or reset type.

    The `temporal workflow reset` command resets a [Workflow Execution](/concepts/what-is-a-workflow-execution).
    A reset allows the Workflow to be resumed from a certain point without losing your parameters or [Event History](/concepts/what-is-an-event-history).
    
    Use the options listed below to change reset behavior.
    Make sure to write the command as follows:
    `temporal workflow reset [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--event-id**="": The Event Id for any Event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--reapply-type**="": Event types to reapply after the reset point: , Signal, None. (default: All)

**--reason**="": Reason to perform a given operation on the Cluster.

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--type**="": Event type to which you want to reset: LastContinuedAsNew, FirstWorkflowTask, LastWorkflowTask

**--workflow-id, -w**="": Workflow Id

### reset-batch

Reset a batch of Workflow Executions by reset type (LastContinuedAsNew), FirstWorkflowTask), LastWorkflowTask

    The `temporal workflow reset-batch` command resets a batch of [Workflow Executions](/concepts/what-is-a-workflow-execution) by `resetType`.
    Resetting a [Workflow](/concepts/what-is-a-workflow) allows the process to resume from a certain point without losing your parameters or [Event History](/concepts/what-is-an-event-history).
    
    Use the options listed below to change reset behavior.
    Make sure to write the command as follows:
    `temporal workflow reset-batch [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--dry-run**: Simulate reset without resetting any Workflow Executions.

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--exclude-file**="": Input file that specifies Workflow Executions to exclude from resetting.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--input-file**="": Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id.

**--input-parallelism**="": Number of goroutines to run in parallel. Each goroutine processes one line for every second. (default: 1)

**--input-separator**="": Separator for the input file. The default is a tab (	). (default: 	)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--non-deterministic**: Reset Workflow Execution only if its last Event is WorkflowTaskFailed with a nondeterministic error.

**--query, -q**="": Visibility Query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query.

**--reason**="": Reason to perform a given operation on the Cluster.

**--skip-base-is-not-current**: Skip a Workflow Execution if the base Run is not the current Run.

**--skip-current-open**: Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--type**="": Event type to which you want to reset: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

### trace

Trace progress of a Workflow Execution and its children.

    The `temporal workflow trace` command tracks the progress of a [Workflow Execution](/concepts/what-is-a-workflow-execution) and any  [Child Workflows](/concepts/what-is-a-child-workflow-execution) it generates.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal workflow trace [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--concurrency**="": Request concurrency. (default: 10)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--depth**="": Number of Child Workflows to expand, -1 to expand all Child Workflows. (default: -1)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fold**="": Statuses for which Child Workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --no-fold supplied. (default: completed,canceled,terminated)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-fold**: Disable folding. All Child Workflows within the set depth will be fetched and displayed.

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

## activity

Operations that can be performed on Workflow Activities.

>Activity commands enable operations on [Activity Executions](/concepts/what-is-an-activity-execution).

### complete

Completes an Activity.

    The `temporal activity complete` command completes an [Activity Execution](/concepts/what-is-an-activity-execution).
    
    Use the options listed below to change the behavior of this command.
    Make sure to write the command as follows:
    `temporal activity complete [command options] [arguments]`

**--activity-id**="": Identifies the Activity Execution.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--identity**="": Specify operator's identity.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--result**="": Set the result value of Activity completion.

**--run-id, -r**="": Identifies the current Workflow Run.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Identifies the Workflow that the Activity is running on.

### fail

Fails an Activity.

    The `temporal activity fail` command fails an [Activity Execution](/concepts/what-is-an-activity-execution).
    
    Use the options listed below to change the behavior of this command.
    Make sure to write the command as follows:
    `temporal activity fail [command options] [arguments]`

**--activity-id**="": Identifies the Activity Execution.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--detail**="": Detail to fail the Activity.

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--identity**="": Specify operator's identity.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--reason**="": Reason to perform a given operation on the Cluster.

**--run-id, -r**="": Identifies the current Workflow Run.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Identifies the Workflow that the Activity is running on.

## task-queue

Operations performed on Task Queues.

>Task Queue commands allow operations to be performed on [Task Queues](/concepts/what-is-a-task-queue).

### describe

Describes the Workers that have recently polled on this Task Queue.

    The `temporal task-queue describe` command provides [poller](/application-development/worker-performance#poller-count) information for a given [Task Queue](/concepts/what-is-a-task-queue).
    
    The [Server](/concepts/what-is-the-temporal-server) records the last time of each poll request.
    Should `LastAccessTime` exceeds one minute, it's likely that the Worker is at capacity (all Workflow and Activity slots are full) or that the Worker has shut down.
    [Workers](/concepts/what-is-a-worker) are removed if 5 minutes have passed since the last poll request.
    
    Use the options listed below to modify what this command returns.
    Make sure to write the command as follows:
    `temporal task-queue describe [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--task-queue, -t**="": Task Queue name.

**--task-queue-type**="": Task Queue type [workflow|activity] (default: workflow)

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### list-partition

Lists the Task Queue's partitions and which matching node they are assigned to.

    The `temporal task-queue list-partition` command displays the partitions of a [Task Queue](/concepts/what-is-a-task-queue), along with the matching node they are assigned to.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal task-queue list-partition [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--task-queue, -t**="": Task Queue name.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

## schedule

Operations performed on Schedules.

    Schedule commands allow the user to create, use, and update [Schedules](/concepts/what-is-a-schedule).
    Schedules control when certain Actions for a Workflow Execution are performed, making it a useful tool for automation.
    
    To run a Schedule command, run `temporal schedule [command] [command options] [arguments]`.

### create

Create a new Schedule.

    The `temporal schedule create` command creates a new [Schedule](/concepts/what-is-a-schedule).
    Newly created Schedules return a Schedule ID to be used in other Schedule commands.
    
    Schedules need to follow a format like the example shown here:
    ```
    temporal schedule create \
    		--sid 'your-schedule-id' \
    		--cron '3 11 * * Fri' \
    		--wid 'your-workflow-id' \
    		--tq 'your-task-queue' \
    		--type 'YourWorkflowType' 
    ```
    
    Any combination of `--cal`, `--interval`, and `--cron` is supported.
    Actions will be executed at any time specified in the Schedule.
    
    Use the options provided below to change the command's behavior.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--calendar**="": Calendar specification in JSON ({"dayOfWeek":"Fri","hour":"17","minute":"5"}) or as a Cron string ("30 2 * * 5" or "@daily").

**--catchup-window**="": Maximum allowed catch-up time if server is down.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--cron**="": Optional Cron Schedule for the Workflow. Cron spec is formatted as: 
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--end-time**="": Overall schedule end time.

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--execution-timeout**="": Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks. (default: 0)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--input, -i**="": Optional JSON input to provide to the Workflow. Pass "null" for null values.

**--input-file**="": Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input.

**--interval**="": Interval duration, e.g. 90m, or 90m/13m to include phase offset.

**--jitter**="": Jitter duration.

**--max-field-length**="": Maximum length for each attribute field. (default: 0)

**--memo**="": Set a memo on a schedule (format: key=value). Use valid JSON formats for value.

**--memo-file**="": Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--notes**="": Initial value of notes field.

**--overlap-policy**="": Overlap policy (options: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll).

**--pause**: Initial value of paused state.

**--pause-on-failure**: Pause schedule after any workflow failure.

**--remaining-actions**="": Total number of actions allowed. (default: 0)

**--run-timeout**="": Timeout (in seconds) of a single Workflow run. (default: 0)

**--schedule-id, -s**="": Schedule Id

**--search-attribute**="": Set Search Attribute on a schedule (format: key=value). Use valid JSON formats for value.

**--start-time**="": Overall schedule start time.

**--task-queue, -t**="": Task Queue

**--task-timeout**="": Start-to-close timeout for a Workflow Task (in seconds). (default: 10)

**--time-zone**="": Time zone (IANA name).

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

**--workflow-type**="": Workflow type name.

### update

Updates a Schedule with a new definition (full replacement, not patch).

    The `temporal schedule update` command updates an existing [Schedule](/concepts/what-is-a-schedule).
    
    Like `temporal schedule create`, updated Schedules need to follow a certain format:
    ```
    temporal schedule update 			\
    		--sid 'your-schedule-id' 	\
    		--cron '3 11 * * Fri' 		\
    		--wid 'your-workflow-id' 	\
    		--tq 'your-task-queue' 		\
    		--type 'YourWorkflowType' 
    ```
    
    Updating a Schedule takes the given options and replaces the entire configuration of the Schedule with what's provided. 
    If you only change one value of the Schedule, be sure to provide the other unchanged fields to prevent them from being overwritten.
    
    Use the options provided below to change the command's behavior.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--calendar**="": Calendar specification in JSON ({"dayOfWeek":"Fri","hour":"17","minute":"5"}) or as a Cron string ("30 2 * * 5" or "@daily").

**--catchup-window**="": Maximum allowed catch-up time if server is down.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--cron**="": Optional Cron Schedule for the Workflow. Cron spec is formatted as: 
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--end-time**="": Overall schedule end time.

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--execution-timeout**="": Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks. (default: 0)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--input, -i**="": Optional JSON input to provide to the Workflow. Pass "null" for null values.

**--input-file**="": Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input.

**--interval**="": Interval duration, e.g. 90m, or 90m/13m to include phase offset.

**--jitter**="": Jitter duration.

**--max-field-length**="": Maximum length for each attribute field. (default: 0)

**--memo**="": Set a memo on a schedule (format: key=value). Use valid JSON formats for value.

**--memo-file**="": Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--notes**="": Initial value of notes field.

**--overlap-policy**="": Overlap policy (options: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll).

**--pause**: Initial value of paused state.

**--pause-on-failure**: Pause schedule after any workflow failure.

**--remaining-actions**="": Total number of actions allowed. (default: 0)

**--run-timeout**="": Timeout (in seconds) of a single Workflow run. (default: 0)

**--schedule-id, -s**="": Schedule Id

**--search-attribute**="": Set Search Attribute on a schedule (format: key=value). Use valid JSON formats for value.

**--start-time**="": Overall schedule start time.

**--task-queue, -t**="": Task Queue

**--task-timeout**="": Start-to-close timeout for a Workflow Task (in seconds). (default: 10)

**--time-zone**="": Time zone (IANA name).

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--workflow-id, -w**="": Workflow Id

**--workflow-type**="": Workflow type name.

### toggle

Pauses or unpauses a Schedule.

    The `temporal schedule toggle` command can pause and unpause a [Schedule](/concepts/what-is-a-schedule).
    
    Toggling a Schedule requires a reason to be entered on the command line. 
    Use `--reason` to note the issue leading to the pause or unpause.
    
    Schedule toggles are passed in this format:
    ` temporal schedule toggle --sid 'your-schedule-id' --pause --reason "paused because the database is down"`
    `temporal schedule toggle --sid 'your-schedule-id' --unpause --reason "the database is back up"`
    
    Use the options provided below to change this command's behavior.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--pause**: Pauses the Schedule.

**--reason**="": Reason to perform a given operation on the Cluster. (default: (no reason provided))

**--schedule-id, -s**="": Schedule Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--unpause**: Unpauses the Schedule.

### trigger

Triggers an immediate action.

    The `temporal schedule trigger` command triggers an immediate action with a given [Schedule](/concepts/what-is-a-schedule).
    By default, this action is subject to the Overlap Policy of the Schedule.
    
    `temporal schedule trigger` can be used to start a Workflow Run immediately.
    `temporal schedule trigger --sid 'your-schedule-id'` 
    
    The Overlap Policy of the Schedule can be overridden as well.
    `temporal schedule trigger --sid 'your-schedule-id' --overlap-policy 'AllowAll'`
    
    Use the options provided below to change this command's behavior.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--overlap-policy**="": Overlap policy (options: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll).

**--schedule-id, -s**="": Schedule Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### backfill

Backfills a past time range of actions.

    The `temporal schedule backfill` command executes Actions ahead of their specified time range. 
    Backfilling can be used to fill in [Workflow Runs](/concepts/what-is-a-run-id) from a time period when the Schedule was paused, or from before the Schedule was created. 
    
    ```
    temporal schedule backfill --sid 'your-schedule-id' \
    		--overlap-policy 'BufferAll' 				\
    		--start-time '2022-05-0101T00:00:00Z'		\
    		--end-time '2022-05-31T23:59:59Z'
    ```
    
    Use the options provided below to change this command's behavior.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--end-time**="": Backfill end time.

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--overlap-policy**="": Overlap policy (options: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll).

**--schedule-id, -s**="": Schedule Id

**--start-time**="": Backfill start time.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### describe

Get Schedule configuration and current state.

    The `temporal schedule describe` command shows the current [Schedule](/concepts/what-is-a-schedule) configuration.
    This command also provides information about past, current, and future [Workflow Runs](/concepts/what-is-a-run-id).
    
    `temporal schedule describe --sid 'your-schedule-id' [command options] [arguments]`
    
    Use the options below to change this command's output.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--raw**: Print raw data as json (prefer this over -o json for scripting).

**--schedule-id, -s**="": Schedule Id

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### delete

Deletes a Schedule.

    The `temporal schedule delete` command deletes a [Schedule](/concepts/what-is-a-schedule).
    Deleting a Schedule does not affect any [Workflows](/concepts/what-is-a-workflow) started by the Schedule.
    
    [Workflow Executions](/concepts/what-is-a-workflow-execution) started by Schedules can be cancelled or terminated like other Workflow Executions.
    However, Workflow Executions started by a Schedule can be identified by their [Search Attributes](/concepts/what-is-a-search-attribute), making them targetable by batch command for termination.
    
    `temporal schedule delete --sid 'your-schedule-id' [command options] [arguments]`
    
    Use the options below to change the behavior of this command.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--schedule-id, -s**="": Schedule Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### list

Lists Schedules.

    The `temporal schedule list` command lists all [Schedule](/concepts/what-is-a-schedule) configurations.
    Listing Schedules in [Standard Visibility](/concepts/what-is-standard-visibility) will only provide Schedule IDs.
    
    `temporal schedule list [command options] [arguments]`
    
    Use the options below to change the behavior of this command.

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--limit**="": Number of items to print. (default: 0)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

## batch

Operations performed on Batch jobs.

    Batch commands allow you to change multiple [Workflow Executions](/concepts/what-is-a-workflow-execution) without having to repeat yourself on the command line. 
    In order to do this, you provide the command with a [List Filter](/concepts/what-is-visibility) and the type of Batch job to execute.
    
    The List Filter identifies the Workflow Executions that will be affected by the Batch job.
    The Batch type determines the other parameters that need to be provided, along with what is being affected on the Workflow Executions.
    
    To start the Batch job, run `temporal workflow query`.
    Running Signal, Terminate, or Cancel with the `--query` modifier will start a Batch job automatically.
    
    A successfully started Batch job will return a Job ID.
    Use this Job ID to execute other actions on the Batch job.

### describe

Describe a Batch operation job.

    The `temporal batch describe` command shows the progress of an ongoing Batch job.
    
    Use the command options listed below to change the information returned by this command.
    Make sure to write the command in this format:
    `temporal batch describe [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--job-id**="": Batch Job Id

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### list

List Batch operation jobs.

    When used, `temporal batch list` returns all Batch jobs. 
    
    Use the command options listed below to change the information returned by this command.
    Make sure to write the command in this format:
    `temporal batch list [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--limit**="": Number of items to print. (default: 0)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### terminate

Stop a Batch operation job.

    The `temporal batch terminate` command terminates a Batch job with the provided Job ID. 
    
    Use the command options listed below to change the behavior of this command.
    Make sure to write the command as follows:
    `temporal batch terminate [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--job-id**="": Batch Job Id

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--reason**="": Reason to perform a given operation on the Cluster.

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

## operator

Operations performed on the Temporal Server.

    Operator commands enable actions on [Namespaces](/concepts/what-is-a-namespace), [Search Attributes](/concepts/what-is-a-search-attribute), and [Temporal Clusters](/concepts/what-is-a-temporal-cluster).
    These actions are performed through subcommands for each Operator area.
    
    To run an Operator command, run `temporal operator [command] [subcommand] [command options] [arguments]`.

### namespace

Operations applying to Namespaces.

>Namespace commands allow [Namespace](/concepts/what-is-a-namespace) operations to be performed on the [Temporal Cluster](/concepts/what-is-a-temporal-cluster).

#### describe

Describe a Namespace by its name or Id.

    The `temporal operator namespace describe` command provides a description of a [Namespace](/concepts/what-is-a-namespace).
    Namespaces can be identified by name or Namespace ID.
    
    Use the options listed below to change the command's output.
    Make sure to write the command as follows:
    `temporal operator namespace describe [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--namespace-id**="": Namespace Id

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### list

List all Namespaces.

    The `temporal operator namespace list` command lists all [Namespaces](/namespaces) on the [Server](/concepts/what-is-a-frontend-service).
    
    Use the options listed below to change the command's output.
    Make sure to write the command as follows:
    `temporal operator namespace list [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### create

Registers a new Namespace.

    The `temporal operator namespace create` command creates a new [Namespace](/concepts/what-is-a-namespace).
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator namespace create [command options] [arguments]`

**--active-cluster**="": Active cluster name

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--cluster**="": Cluster name

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--data**="": Namespace data in a format key=value

**--description**="": Namespace description

**--email**="": Owner email

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--global**="": Flag to indicate whether namespace is a global namespace

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--history-archival-state**="": Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**="": Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--retention**="": Workflow Execution retention

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--visibility-archival-state**="": Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**="": Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### update

Updates a Namespace.

    The `temporal operator namespace update` command updates a given [Namespace](/concepts/what-is-a-namespace).
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator namespace update [command options] [arguments]`

**--active-cluster**="": Active cluster name

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--cluster**="": Cluster name

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--data**="": Namespace data in a format key=value

**--description**="": Namespace description

**--email**="": Owner email

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--history-archival-state**="": Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**="": Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--promote-global**: Promote local namespace to global namespace

**--reason**="": Reason for the operation

**--retention**="": Workflow Execution retention

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--visibility-archival-state**="": Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**="": Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### delete

Deletes an existing Namespace.

    The `temporal operator namespace delete` command deletes a given [Namespace](/concepts/what-is-a-namespace) from the system.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator namespace delete [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--yes, -y**: Confirm all prompts.

### search-attribute

Operations applying to Search Attributes.

>Search Attribute commands enable operations for the creation, listing, and removal of [Search Attributes](/concepts/what-is-a-search-attribute).

#### create

Adds one or more custom Search Attributes.

    The `temporal operator search-attribute create` command adds one or more custom [Search Attributes](/concepts/what-is-a-search-attribute).
    These Search Attributes can be used to [filter a list](/concepts/what-is-a-list-filter) of [Workflow Executions](/concepts/what-is-a-workflow-execution) that contain the given Search Attributes in their metadata.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator search-attribute create [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--name**="": Search Attribute name.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--type**="": Search attribute type: [Text Keyword Int Double Bool Datetime KeywordList]

**--yes, -y**: Confirm all prompts.

#### list

Lists all Search Attributes that can be used in list Workflow Queries.

    The `temporal operator search-attrbute list` command displays a list of all [Search Attributes](/concepts/what-is-a-search-attribute) that can be used in ` temporal workflow list --query`.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator search-attribute list [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### remove

Removes custom search attribute metadata only (Elasticsearch index schema is not modified).

    The `temporal operator search-attribute remove` command removes custom [Search Attribute](/concepts/what-is-a-search-attribute) metadata.
    This command does not remove custom Search Attributes from Elasticsearch.
    The index schema is not modified.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator search-attribute remove [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--name**="": Search Attribute name.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

**--yes, -y**: Confirm all prompts.

### cluster

Operations for running a Temporal Cluster.

>Cluster commands enabled operations on [Temporal Clusters](/concepts/what-is-a-temporal-cluster).

#### health

Checks the health of the Frontend Service.

    The `temporal operator cluster health` command checks the health of the [Frontend Service](/concepts/what-is-a-frontend-service).
    
    Use the options listed below to change the behavior and output of this command.
    Make sure to write the command as follows:
    `temporal operator cluster health [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### describe

Show information about the Cluster.

    The `temporal operator cluster describe` command shows information about the [Cluster](/concepts/what-is-a-temporal-cluster).
    
    Use the options listed below to change the output of this command.
    Make sure to write the command as follows:
    `temporal operator cluster describe [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### system

Shows information about the system and its capabilities.

    The `temporal operator cluster system` command provides information about the system the Cluster is running on.
    
    Use the options listed below to change this command's output.
    Make sure to write the command as follows:
    `temporal operator cluster system [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### upsert

Add or update a remote Cluster.

    The `temporal operator cluster upsert` command allows the user to add or update a remote [Cluster](/concepts/what-is-a-temporal-cluster).
    
    Use the options listed below to change the behavior of this command.
    Make sure to write the command as follows:
    `temporal operator cluster upsert [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--enable-connection**: Enable cross-cluster connection.

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--frontend-address**="": Frontend address of the remote Cluster.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### list

List all remote Clusters.

    The `temporal operator cluster list` command prints a list of all remote [Clusters](/concepts/what-is-a-temporal-cluster) on the system.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator cluster list [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--fields**="": Customize fields to print. Set to 'long' to automatically print more of main fields.

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--limit**="": Number of items to print. (default: 0)

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--no-pager, -P**: Disables the interactive pager.

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": Sets the pager for Temporal CLI to use (options: less, more, favoritePager).

**--time-format**="": Format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

#### remove

Remove a remote Cluster.

    The `temporal operator cluster remove` command removes a remote [Cluster](/concepts/what-is-a-temporal-cluster) from the system.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal operator cluster remove [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--name**="": Frontend address of the remote Cluster.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

## env

Manage environmental configurations on Temporal Client.

>Environment (or 'env') commands allow the user to configure the properties for the environment in use.

### get

Prints environmental properties.

    The `temporal env get` command prints the environmental properties for the environment in use.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal env get [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### set

Set environmental properties.

    The `temporal env set` command sets the value for an environmental property.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal env set [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

### delete

Delete an environment or environmental property.

    The `temporal env delete` command deletes a given environment or environmental property.
    
    Use the options listed below to change the command's behavior.
    Make sure to write the command as follows:
    `temporal env delete [command options] [arguments]`

**--address**="": The host and port (formatted as host:port) for the Temporal Frontend Service.

**--codec-auth**="": Sets the authorization header on requests to the Codec Server.

**--codec-endpoint**="": Endpoint for a remote Codec Server.

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": An optional timeout for the context of an RPC call (in seconds). (default: 5)

**--env**="": Name of the environment to read environmental variables from. (default: default)

**--grpc-meta**="": Contains gRPC metadata to send with requests (format: key=value). Values must be in a valid JSON format.

**--namespace, -n**="": Identifies a Namespace in the Temporal Workflow. (default: default)

**--tls-ca-path**="": Path to server CA certificate.

**--tls-cert-path**="": Path to x509 certificate.

**--tls-disable-host-verification**: Disables TLS host name verification if already enabled.

**--tls-key-path**="": Path to private certificate key.

**--tls-server-name**="": Provides an override for the target TLS server name.

## completion

Output shell completion code for the specified shell (zsh, bash).

### bash

bash completion output

>source <(temporal completion bash)

### zsh

zsh completion output

>source <(temporal completion zsh)
