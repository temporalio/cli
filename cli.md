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

## workflow

Operations on Workflows

### start

Start a new Workflow Execution

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--cron**="": Optional cron schedule for the Workflow. Cron spec is as following: 
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--env**="": Env name to read the client environment variables from (default: default)

**--execution-timeout**="": Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--id-reuse-policy**="": Configure if the same Workflow Id is allowed for use in new Workflow Execution. Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning

**--input, -i**="": Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**="": Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--limit**="": number of items to print (default: 0)

**--max-field-length**="": Maximum length for each attribute field (default: 0)

**--memo**="": Pass a memo in a format key=value. Use valid JSON formats for value

**--memo-file**="": Pass a memo from a file, where each line follows the format key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--run-timeout**="": Single workflow run timeout (seconds) (default: 0)

**--search-attribute**="": Pass Search Attribute in a format key=value. Use valid JSON formats for value

**--task-queue, -t**="": Task queue

**--task-timeout**="": Workflow task start to close timeout (seconds) (default: 10)

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": Workflow type name

**--workflow-id, -w**="": Workflow Id

### execute

Start a new Workflow Execution and print progress

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--cron**="": Optional cron schedule for the Workflow. Cron spec is as following: 
	┌───────────── minute (0 - 59) 
	│ ┌───────────── hour (0 - 23) 
	│ │ ┌───────────── day of the month (1 - 31) 
	│ │ │ ┌───────────── month (1 - 12) 
	│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) 
	│ │ │ │ │ 
	* * * * *

**--env**="": Env name to read the client environment variables from (default: default)

**--execution-timeout**="": Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--id-reuse-policy**="": Configure if the same Workflow Id is allowed for use in new Workflow Execution. Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning

**--input, -i**="": Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**="": Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--limit**="": number of items to print (default: 0)

**--max-field-length**="": Maximum length for each attribute field (default: 0)

**--memo**="": Pass a memo in a format key=value. Use valid JSON formats for value

**--memo-file**="": Pass a memo from a file, where each line follows the format key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--run-timeout**="": Single workflow run timeout (seconds) (default: 0)

**--search-attribute**="": Pass Search Attribute in a format key=value. Use valid JSON formats for value

**--task-queue, -t**="": Task queue

**--task-timeout**="": Workflow task start to close timeout (seconds) (default: 10)

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": Workflow type name

**--workflow-id, -w**="": Workflow Id

### describe

Show information about a Workflow Execution

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--raw**: Print properties as they are stored

**--reset-points**: Only show auto-reset points

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

### list

List Workflow Executions based on a Query

**--address**="": host:port for Temporal frontend service

**--archived**: List archived Workflow Executions (EXPERIMENTAL)

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--limit**="": number of items to print (default: 0)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--query, -q**="": Filter results using SQL like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for details

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### show

Show Event History for a Workflow Execution

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--follow, -f**: Follow the progress of Workflow Execution

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--limit**="": number of items to print (default: 0)

**--max-field-length**="": Maximum length for each attribute field (default: 500)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--output-filename**="": Serialize history event to a file

**--pager**="": pager to use: less, more, favoritePager..

**--reset-points**: Only show events that are eligible for reset

**--run-id, -r**="": Run Id

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

### query

Query a Workflow Execution

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Optional input for the query, in JSON format. If there are multiple parameters, concatenate them and separate by space

**--input-file**="": Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. Input from file will be overwrite by input from command line

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--reject-condition**="": Optional flag to reject queries based on Workflow state. Valid values are "not_open" and "not_completed_cleanly"

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": The query type you want to run

**--workflow-id, -w**="": Workflow Id

### stack

Query a Workflow Execution with __stack_trace as the query type

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Optional input for the query, in JSON format. If there are multiple parameters, concatenate them and separate by space

**--input-file**="": Optional input for the query from JSON file. If there are multiple JSON, concatenate them and separate by space or newline. Input from file will be overwrite by input from command line

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--reject-condition**="": Optional flag to reject queries based on Workflow state. Valid values are "not_open" and "not_completed_cleanly"

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

### signal

Signal Workflow Execution by Id or List Filter

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Input for the signal (JSON)

**--input-file**="": Input for the signal from file (JSON)

**--name**="": Signal Name

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--query, -q**="": Signal Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/

**--reason**="": Reason for signaling with List Filter

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Signal Workflow Execution by Id

**--yes, -y**: Confirm all prompts

### count

Count Workflow Executions (requires ElasticSearch to be enabled)

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--query, -q**="": Filter results using SQL like query. See https://docs.temporal.io/docs/tctl/workflow/list#--query for details

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### cancel

Cancel a Workflow Execution

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--query, -q**="": Cancel Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/

**--reason**="": Reason for canceling with List Filter

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Cancel Workflow Execution by Id

**--yes, -y**: Confirm all prompts

### terminate

Terminate Workflow Execution by Id or List Filter

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--query, -q**="": Terminate Workflow Executions by List Filter. See https://docs.temporal.io/concepts/what-is-a-list-filter/

**--reason**="": Reason for termination

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Terminate Workflow Execution by Id

**--yes, -y**: Confirm all prompts

### delete

Delete a Workflow Execution

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

### reset

Reset a Workflow Execution by event Id or reset type

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--event-id**="": The eventId of any event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--reapply-type**="": Event types to reapply after the reset point: , Signal, None. (default: All)

**--reason**="": Reason to reset

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": Event type to which you want to reset: LastWorkflowTask, LastContinuedAsNew, FirstWorkflowTask

**--workflow-id, -w**="": Workflow Id

### reset-batch

Reset a batch of Workflow Executions by reset type: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--dry-run**: Simulate reset without resetting any Workflow Executions

**--env**="": Env name to read the client environment variables from (default: default)

**--exclude-file**="": Input file that specifies Workflow Executions to exclude from resetting

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input-file**="": Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id

**--input-parallelism**="": Number of goroutines to run in parallel. Each goroutine processes one line for every second (default: 1)

**--input-separator**="": Separator for the input file. The default is a tab (	) (default: 	)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--non-deterministic**: Reset Workflow Execution only if its last Event is WorkflowTaskFailed with a nondeterministic error

**--query, -q**="": Visibility query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query

**--reason**="": Reason for resetting the Workflow Executions

**--skip-base-is-not-current**: Skip a Workflow Execution if the base Run is not the current Run

**--skip-current-open**: Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": Event type to which you want to reset: LastContinuedAsNew, FirstWorkflowTask, LastWorkflowTask

### trace

Trace progress of a Workflow Execution and its children

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--concurrency**="": Request concurrency (default: 10)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--depth**="": Number of child workflows to expand, -1 to expand all child workflows (default: -1)

**--env**="": Env name to read the client environment variables from (default: default)

**--fold**="": Statuses for which child workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if --no-fold supplied (default: completed,canceled,terminated)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-fold**: Disable folding. All child workflows within the set depth will be fetched and displayed

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

## activity

Operations on Activities of Workflows

### complete

Complete an activity

**--activity-id**="": The Activity Id to complete

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--identity**="": Specify operator's identity

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--result**="": Set the result value of completion

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

### fail

Fail an activity

**--activity-id**="": The Activity Id to fail

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--detail**="": Detail to fail the Activity

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--identity**="": Specify operator's identity

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--reason**="": Reason to fail the Activity

**--run-id, -r**="": Run Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

## task-queue

Operations on Task Queues

### describe

Describe the Workers that have recently polled on this Task Queue

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--task-queue, -t**="": Task Queue name

**--task-queue-type**="": Task Queue type [workflow|activity] (default: workflow)

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### list-partition

List the Task Queue's partitions and which matching node they are assigned to

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--task-queue, -t**="": Task Queue name

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

## schedule

Operations on Schedules

### create

Create a new schedule

**--address**="": host:port for Temporal frontend service

**--calendar**="": Calendar specification in JSON, e.g. {"dayOfWeek":"Fri","hour":"17","minute":"5"}

**--catchup-window**="": Maximum allowed catch-up time if server is down

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--cron**="": Calendar specification as cron string, e.g. "30 2 * * 5" or "@daily"

**--end-time**="": Overall schedule end time

**--env**="": Env name to read the client environment variables from (default: default)

**--execution-timeout**="": Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**="": Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--interval**="": Interval duration, e.g. 90m, or 90m/13m to include phase offset

**--jitter**="": Jitter duration

**--max-field-length**="": Maximum length for each attribute field (default: 0)

**--memo**="": Set a memo on a schedule. Format: key=value. Use valid JSON formats for value

**--memo-file**="": Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--notes**="": Initial value of notes field

**--overlap-policy**="": Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll

**--pause**: Initial value of paused state

**--pause-on-failure**: Pause schedule after any workflow failure

**--remaining-actions**="": Total number of actions allowed (default: 0)

**--run-timeout**="": Single workflow run timeout (seconds) (default: 0)

**--schedule-id, -s**="": Schedule Id

**--search-attribute**="": Set Search Attribute on a schedule. Format: key=value. Use valid JSON formats for value

**--start-time**="": Overall schedule start time

**--task-queue, -t**="": Task queue

**--task-timeout**="": Workflow task start to close timeout (seconds) (default: 10)

**--time-zone**="": Time zone (IANA name)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

**--workflow-type**="": Workflow type name

### update

Updates a schedule with a new definition (full replacement, not patch)

**--address**="": host:port for Temporal frontend service

**--calendar**="": Calendar specification in JSON, e.g. {"dayOfWeek":"Fri","hour":"17","minute":"5"}

**--catchup-window**="": Maximum allowed catch-up time if server is down

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--cron**="": Calendar specification as cron string, e.g. "30 2 * * 5" or "@daily"

**--end-time**="": Overall schedule end time

**--env**="": Env name to read the client environment variables from (default: default)

**--execution-timeout**="": Workflow Execution timeout, including retries and continue-as-new (seconds) (default: 0)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--input, -i**="": Optional input for the Workflow in JSON format. Pass "null" for null values

**--input-file**="": Pass an optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line overwrites input from the file

**--interval**="": Interval duration, e.g. 90m, or 90m/13m to include phase offset

**--jitter**="": Jitter duration

**--max-field-length**="": Maximum length for each attribute field (default: 0)

**--memo**="": Set a memo on a schedule. Format: key=value. Use valid JSON formats for value

**--memo-file**="": Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--notes**="": Initial value of notes field

**--overlap-policy**="": Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll

**--pause**: Initial value of paused state

**--pause-on-failure**: Pause schedule after any workflow failure

**--remaining-actions**="": Total number of actions allowed (default: 0)

**--run-timeout**="": Single workflow run timeout (seconds) (default: 0)

**--schedule-id, -s**="": Schedule Id

**--search-attribute**="": Set Search Attribute on a schedule. Format: key=value. Use valid JSON formats for value

**--start-time**="": Overall schedule start time

**--task-queue, -t**="": Task queue

**--task-timeout**="": Workflow task start to close timeout (seconds) (default: 10)

**--time-zone**="": Time zone (IANA name)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--workflow-id, -w**="": Workflow Id

**--workflow-type**="": Workflow type name

### toggle

Pauses or unpauses a schedule

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--pause**: Pauses the schedule

**--reason**="": Free-form text to describe reason for pause/unpause (default: (no reason provided))

**--schedule-id, -s**="": Schedule Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--unpause**: Unpauses the schedule

### trigger

Triggers an immediate action

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--overlap-policy**="": Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll

**--schedule-id, -s**="": Schedule Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### backfill

Backfills a past time range of actions

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--end-time**="": Backfill end time

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--overlap-policy**="": Overlap policy: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll

**--schedule-id, -s**="": Schedule Id

**--start-time**="": Backfill start time

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### describe

Get schedule configuration and current state

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--raw**: Print raw data as json (prefer this over -o json for scripting)

**--schedule-id, -s**="": Schedule Id

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### delete

Deletes a schedule

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--schedule-id, -s**="": Schedule Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### list

Lists schedules

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--limit**="": number of items to print (default: 0)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

## batch

Operations on Batch jobs. Use workflow commands with --query flag to start batch jobs

### describe

Describe a batch operation job

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--job-id**="": Batch Job Id

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### list

List batch operation jobs

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--limit**="": number of items to print (default: 0)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### terminate

Stop a batch operation job

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--job-id**="": Batch Job Id

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--reason**="": Reason to stop the batch job

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

## operator

Operation on Temporal server

### namespace

Operations on namespaces

#### describe

Describe a Namespace by name or Id

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--namespace-id**="": Namespace Id

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### list

List all Namespaces

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### register

Register a new Namespace

**--active-cluster**="": Active cluster name

**--address**="": host:port for Temporal frontend service

**--cluster**="": Cluster name

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--data**="": Namespace data in a format key=value

**--description**="": Namespace description

**--email**="": Owner email

**--env**="": Env name to read the client environment variables from (default: default)

**--global**="": Flag to indicate whether namespace is a global namespace

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--history-archival-state**="": Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**="": Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--retention**="": Workflow Execution retention

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--visibility-archival-state**="": Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**="": Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### update

Update a Namespace

**--active-cluster**="": Active cluster name

**--address**="": host:port for Temporal frontend service

**--cluster**="": Cluster name

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--data**="": Namespace data in a format key=value

**--description**="": Namespace description

**--email**="": Owner email

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--history-archival-state**="": Flag to set history archival state, valid values are "disabled" and "enabled"

**--history-uri**="": Optionally specify history archival URI (cannot be changed after first time archival is enabled)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--promote-global**: Promote local namespace to global namespace

**--reason**="": Reason for the operation

**--retention**="": Workflow Execution retention

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--visibility-archival-state**="": Flag to set visibility archival state, valid values are "disabled" and "enabled"

**--visibility-uri**="": Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)

#### delete

Delete existing Namespace

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--yes, -y**: Confirm all prompts

### search-attribute

Operations on search attributes

#### create

Add custom search attributes

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--name**="": Search attribute name

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--type**="": Search attribute type: [Text Keyword Int Double Bool Datetime KeywordList]

**--yes, -y**: Confirm all prompts

#### list

List search attributes that can be used in list workflow query

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### remove

Remove custom search attributes metadata only (Elasticsearch index schema is not modified)

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--name**="": Search attribute name

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

**--yes, -y**: Confirm all prompts

### cluster

Operations on a Temporal cluster

#### health

Check health of frontend service

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### describe

Show information about the cluster

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### system

Show information about the system and capabilities

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--output, -o**="": format output as: table, json, card. (default: table)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### upsert

Add or update a remote cluster

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--enable-connection**: Enable cross cluster connection

**--env**="": Env name to read the client environment variables from (default: default)

**--frontend-address**="": Frontend address of the remote cluster

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### list

List all remote clusters

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--fields**="": customize fields to print. Set to 'long' to automatically print more of main fields

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--limit**="": number of items to print (default: 0)

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--no-pager, -P**: disable interactive pager

**--output, -o**="": format output as: table, json, card. (default: table)

**--pager**="": pager to use: less, more, favoritePager..

**--time-format**="": format time as: relative, iso, raw. (default: relative)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

#### remove

Remove a remote cluster

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--name**="": Frontend address of the remote cluster

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

## env

Manage client environment configurations

### get

Print environment properties

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### set

Set environment property

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

### delete

Delete environment or environment property

**--address**="": host:port for Temporal frontend service

**--codec-auth**="": Authorization header to set for requests to Codec Server

**--codec-endpoint**="": Remote Codec Server Endpoint

**--color**="": when to use color: auto, always, never. (default: auto)

**--context-timeout**="": Optional timeout for context of RPC call in seconds (default: 5)

**--env**="": Env name to read the client environment variables from (default: default)

**--grpc-meta**="": gRPC metadata to send with requests. Format: key=value. Use valid JSON formats for value

**--namespace, -n**="": Temporal workflow namespace (default: default)

**--tls-ca-path**="": Path to server CA certificate

**--tls-cert-path**="": Path to x509 certificate

**--tls-disable-host-verification**: Disable tls host name verification (tls must be enabled)

**--tls-key-path**="": Path to private key

**--tls-server-name**="": Override for target server name

## completion

Output shell completion code for the specified shell (zsh, bash)

### bash

bash completion output

### zsh

zsh completion output
