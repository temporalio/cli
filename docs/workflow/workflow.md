
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

**--type**="": Event type to which you want to reset: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew

**--workflow-id, -w**="": Workflow Id

### reset-batch

Reset a batch of Workflow Executions by reset type: LastContinuedAsNew, FirstWorkflowTask, LastWorkflowTask

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

