# Temporal CLI Commands

Commands for the Temporal CLI.

<!--

This document has a specific structure used by a parser. Here are the rules:

* Each command is a `### <command>: <short-description>` heading. Command is full command path including parent
  commands.
  * Heading must be a single line.
  * <command> will have all up until the leading "[" as the name and the bracket-and-post-bracket content as just
    additional "use" info. This would be used to document positional arguments.
  * Contents of each command section up to `#### Options` is the long description of the command.
    * End of long description can have XML comment section that has `*` bulleted attributes (one line per bullet):
      * `* has-init` - Will assume an `initCommand` method is on the command
      * `* exact-args=<number>` - Require this exact number of args
  * Can have `#### Options` or `#### Options set for <options-set-name>` which can have options.
    * Can have bullets
      * Each bullet is `* <option-names> (<data-type>) - <short-description>. <extra-attributes>`.
      * `<option-names>` is `` `--<option-name>` `` and can optionally be followed by ``, `-<short-name>` ``.
      * `<data-type>` must be one of `bool`, `duration`, `int`, `string`, `string[]`, `string-enum`, TODO: more
      * `<short-description>` can be just about anything so long as it doesn't match trailing attributes. Any wrap
        around to newlines + two-space indention is trimmed to a single space.
      * `<extra-attributes>` can be:
        * `Required.` - Marks the option as required.
        * `Default: <default-value>.` - Sets the default value of the option. No default means zero value of the type.
        * `Options: <option>, <option>.` - Sets the possible options for a string enum type.
        * `Env: <env-var>.` - Binds the environment variable to this flag.
      * Options should be in order of most commonly used.
    * Also can have single lines below options that say
      `Includes options set for [<options-set-name>](#options-set-for-<options-set-link-name>).` which is the equivalent
      of just having them appended to the bulleted list.
      * Just because a command may have a couple of similar options with another doesn't mean you _have_ to make a
        shareable options set. Copy/paste is acceptable.
* Keep commands in alphabetical order.
* Commands that have subcommands cannot be run on their own.
* Keep lines at 120 chars if possible.
* Add punctuation even at the end of phrases.

-->

### temporal: Temporal command-line interface and development server.

<!--
* has-init
-->

#### Options

* `--env` (string) - Environment to read environment-specific flags from. Default: default. Env: TEMPORAL_ENV.
* `--env-file` (string) - File to read all environments (defaults to `$HOME/.config/temporalio/temporal.yaml`).
* `--log-level` (string-enum) - Log level. Options: debug, info, warn, error, off. Default: info.
* `--log-format` (string-enum) - Log format. Options: text, json. Default: text.
* `--output`, `-o` (string-enum) - Data output format. Options: text, json. Default: text.
* `--time-format` (string-enum) - Time format. Options: relative, iso, raw. Default: relative.
* `--color` (string-enum) - Set coloring. Options: always, never, auto. Default: auto.
* `--no-json-shorthand-payloads` (bool) - Always all payloads as raw payloads even if they are JSON.

### temporal env: Manage environments.

Use the '--env <env name>' option with other commands to point the CLI at a different Temporal Server instance. If --env
is not passed, the 'default' environment is used.

### temporal env delete [environment or property]: Delete an environment or environment property.

`temporal env delete [environment or property]`

Delete an environment or just a single property:

`temporal env delete prod`
`temporal env delete prod.tls-cert-path`

<!--
* exact-args=1
-->

### temporal env get [environment or property]: Print environment properties.

`temporal env get [environment or property]`

Print all properties of the 'prod' environment:

`temporal env get prod`

tls-cert-path  /home/my-user/certs/client.cert
tls-key-path   /home/my-user/certs/client.key
address        temporal.example.com:7233
namespace      someNamespace

Print a single property:

`temporal env get prod.tls-key-path`

tls-key-path  /home/my-user/certs/cluster.key

<!--
* exact-args=1
-->

### temporal env list: Print all environments.

List all environments.

### temporal env set [environment.property name] [property value]: Set environment properties.

`temporal env set [environment.property name] [property value]`

Property names match CLI option names, for example '--address' and '--tls-cert-path':

`temporal env set prod.address 127.0.0.1:7233`
`temporal env set prod.tls-cert-path  /home/my-user/certs/cluster.cert`

<!--
* exact-args=2
-->

### temporal server: Run Temporal Server.

Start a development version of [Temporal Server](/concepts/what-is-the-temporal-server):

`temporal server start-dev`

### temporal server start-dev: Start Temporal development server.

Start [Temporal Server](/concepts/what-is-the-temporal-server) on `localhost:7233` with:

`temporal server start-dev`

View the UI at http://localhost:8233

To persist Workflows across runs, use:

`temporal server start-dev --db-filename temporal.db`

#### Options

* `--db-filename`, `-f` (string) - File in which to persist Temporal state (by default, Workflows are lost when the
  process dies).
* `--namespace`, `-n` (string[]) - Specify namespaces that should be pre-created (namespace "default" is always
  created).
* `--port`, `-p` (int) - Port for the frontend gRPC service. Default: 7233.
* `--http-port` (int) - Port for the frontend HTTP API service. Default is off.
* `--metrics-port` (int) - Port for /metrics. Default is off.
* `--ui-port` (int) - Port for the Web UI. Default is --port + 1000.
* `--headless` (bool) - Disable the Web UI.
* `--ip` (string) - IP address to bind the frontend service to. Default: 127.0.0.1.
* `--ui-ip` (string) - IP address to bind the Web UI to. Default is same as --ip.
* `--ui-asset-path` (string) - UI custom assets path.
* `--ui-codec-endpoint` (string) - UI remote codec HTTP endpoint.
* `--sqlite-pragma` (string[]) - Specify SQLite pragma statements in pragma=value format.
* `--dynamic-config-value` (string[]) - Dynamic config value, as KEY=JSON_VALUE (string values need quotes).
* `--log-config` (bool) - Log the server config being used in stderr.

### temporal task-queue: Manage Task Queues.

Task Queue commands allow operations to be performed on [Task Queues](/concepts/what-is-a-task-queue). To run a Task
Queue command, run `temporal task-queue [command] [command options]`.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal task-queue describe: Provides information for Workers that have recently polled on this Task Queue.

The `temporal task-queue describe` command provides [poller](/application-development/worker-performance#poller-count)
information for a given [Task Queue](/concepts/what-is-a-task-queue).

The [Server](/concepts/what-is-the-temporal-server) records the last time of each poll request. A `LastAccessTime` value
in excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the
Worker has shut down. [Workers](/concepts/what-is-a-worker) are removed if 5 minutes have passed since the last poll
request.

Information about the Task Queue can be returned to troubleshoot server issues.

`temporal task-queue describe --task-queue=MyTaskQueue --task-queue-type="activity"`

Use the options listed below to modify what this command returns.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.
* `--task-queue-type` (string-enum) - Task Queue type. Options: workflow, activity. Default: workflow.

### temporal workflow: Start, list, and operate on Workflows.

[Workflow](/concepts/what-is-a-workflow) commands perform operations on 
[Workflow Executions](/concepts/what-is-a-workflow-execution).

Workflow commands use this syntax:`temporal workflow COMMAND [ARGS]`.

#### Options set for client:

* `--address` (string) - Temporal server address. Default: 127.0.0.1:7233. Env: TEMPORAL_ADDRESS.
* `--namespace`, `-n` (string) - Temporal server namespace. Default: default. Env: TEMPORAL_NAMESPACE.
* `--grpc-meta` (string[]) - HTTP headers to send with requests (formatted as key=value).
* `--tls` (bool) - Enable TLS encryption without additional options such as mTLS or client certificates. Env:
  TEMPORAL_TLS.
* `--tls-cert-path` (string) - Path to x509 certificate. Env: TEMPORAL_TLS_CERT.
* `--tls-key-path` (string) - Path to private certificate key. Env: TEMPORAL_TLS_KEY.
* `--tls-ca-path` (string) - Path to server CA certificate. Env: TEMPORAL_TLS_CA.
* `--tls-disable-host-verification` (bool) - Disables TLS host-name verification. Env:
  TEMPORAL_TLS_DISABLE_HOST_VERIFICATION.
* `--tls-server-name` (string) - Overrides target TLS server name. Env: TEMPORAL_TLS_SERVER_NAME.
* `--codec-endpoint` (string) - Endpoint for a remote Codec Server. Env: TEMPORAL_CODEC_ENDPOINT.
* `--codec-auth` (string) - Sets the authorization header on requests to the Codec Server. Env: TEMPORAL_CODEC_AUTH.

### temporal workflow cancel: Cancel a Workflow Execution.

TODO

### temporal workflow count: Count Workflow Executions.

TODO

### temporal workflow delete: Deletes a Workflow Execution.

TODO

### temporal workflow describe: Show information about a Workflow Execution.

The `temporal workflow describe` command shows information about a given
[Workflow Execution](/concepts/what-is-a-workflow-execution).

This information can be used to locate Workflow Executions that weren't able to run successfully.

`temporal workflow describe --workflow-id=meaningful-business-id`

Output can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.

`temporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true`

Use the command options below to change the information returned by this command.

#### Options

* `--workflow-id`, `-w` (string) - Workflow Id. Required.
* `--run-id`, `-r` (string) - Run Id.
* `--reset-points` (bool) - Only show auto-reset points.
* `--raw` (bool) - Print properties without changing their format.

### temporal workflow execute: Start a new Workflow Execution and prints its progress.

The `temporal workflow execute` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution) and
prints its progress. The command completes when the Workflow Execution completes.

Single quotes('') are used to wrap input as JSON.

```
temporal workflow execute
		--workflow-id meaningful-business-id \
		--type MyWorkflow \
		--task-queue MyTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options

* `--event-details` (bool) - If set when using text output, this will print the event details instead of just the event
  during workflow progress. If set when using JSON output, this will include the entire "history" JSON key of the
  started run (does not follow runs).

Includes options set for [workflow start](#options-set-for-workflow-start).
Includes options set for [payload input](#options-set-for-payload-input).

### temporal workflow list: List Workflow Executions based on a Query.

The `temporal workflow list` command provides a list of [Workflow Executions](/concepts/what-is-a-workflow-execution)
that meet the criteria of a given [Query](/concepts/what-is-a-query).
By default, this command returns up to 10 closed Workflow Executions.

`temporal workflow list --query=MyQuery`

The command can also return a list of archived Workflow Executions.

`temporal workflow list --archived`

Use the command options below to change the information returned by this command.

#### Options

* `--query`, `-q` (string) - Filter results using a SQL-like query.
* `--archived` (bool) - If set, will only query and list archived workflows instead of regular workflows.
* `--limit` (int) - Limit the number of items to print.

### temporal workflow query: Query a Workflow Execution.

TODO

### temporal workflow reset: Resets a Workflow Execution by Event ID or reset type.

TODO

### temporal workflow reset-batch: Reset a batch of Workflow Executions by reset type.

TODO

### temporal workflow show: Show Event History for a Workflow Execution.

The `temporal workflow show` command provides the [Event History](/concepts/what-is-an-event-history) for a
[Workflow Execution](/concepts/what-is-a-workflow-execution).

Use the options listed below to change the command's behavior.

#### Options

* `--workflow-id`, `-w` (string) - Workflow Id. Required.
* `--run-id`, `-r` (string) - Run Id.
* `--reset-points` (bool) - Only show auto-reset points.
* `--follow` (bool) - Follow the progress of a Workflow Execution if it goes to a new run.

### temporal workflow signal: Signal Workflow Execution by Id or List Filter.

TODO

### temporal workflow stack: Query a Workflow Execution with __stack_trace as the query type.

TODO

### temporal workflow start: Starts a new Workflow Execution.

The `temporal workflow start` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution). The
Workflow and Run IDs are returned after starting the [Workflow](/concepts/what-is-a-workflow).

```
temporal workflow start \
		--workflow-id meaningful-business-id \
		--type MyWorkflow \
		--task-queue MyTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options set for workflow start:

* `--workflow-id`, `-w` (string) - Workflow Id.
* `--type` (string) - Workflow Type name. Required.
* `--task-queue`, `-t` (string) - Workflow Task queue. Required.
* `--run-timeout` (duration) - Timeout of a Workflow Run.
* `--execution-timeout` (duration) - Timeout for a WorkflowExecution, including retries and ContinueAsNew tasks.
* `--task-timeout` (duration) - Start-to-close timeout for a Workflow Task. Default: 10s.
* `--cron` (string) - Cron schedule for the workflow. Deprecated - use schedules instead.
* `--id-reuse-policy` (string) - Allows the same Workflow Id to be used in a new Workflow Execution. Options:
  AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.
* `--search-attribute` (string[]) - Passes Search Attribute in key=value format. Use valid JSON formats for value.
* `--memo` (string[]) - Passes Memo in key=value format. Use valid JSON formats for value.
* `--fail-existing` (bool) - Fail if the workflow already exists.

#### Options set for payload input:

* `--input`, `-i` (string[]) - Input value (default JSON unless --input-payload-meta is non-JSON encoding). Can
  be given multiple times for multiple arguments. Cannot be combined with --input-file.
* `--input-file` (string[]) - Reads a file as the input (JSON by default unless --input-payload-meta is non-JSON
  encoding). Can be given multiple times for multiple arguments. Cannot be combined with --input.
* `--input-meta` (string[]) - Metadata for the input payload. Expected as key=value. If key is encoding, overrides the
  default of json/plain.
* `--input-base64` (bool) - If set, assumes --input or --input-file are base64 encoded and attempts to decode.

### temporal workflow terminate: Terminate Workflow Execution by ID or List Filter.

TODO

### temporal workflow trace: Trace progress of a Workflow Execution and its children.

TODO

### temporal workflow update: Updates a running workflow synchronously.

TODO