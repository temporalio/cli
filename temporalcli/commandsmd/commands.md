# Temporal CLI Commands

Commands for the Temporal CLI.

<!--

This document is automatically parsed.
Follow these rules.

IN-HOUSE STYLE

* Use alphabetical order for commands.
* Commands with subcommands cannot be run on their own.
* Line length: 120 characters max.
* Add punctuation consistently, even at the end of phrases.
* Short descriptions do not end with a period.
* Every long description demonstrates at least one example use of the command.
* Use imperative-form verbs:
  * Good: `Pause or unpause a Schedule.`
  * Bad: `This command pauses or unpauses a Schedule.`

For options and flags:

* For optional and required flags, use complete sentences with imperative-form verbs. 
* When flags can be passed multiple times, say so explicitly in the usage text.
* Do not rely on the flag type (e.g. `string`, `bool`, etc.) being shown to the user.
  It is hidden if a `META-VARIABLE` is used.
* Where possible, use a `META-VARIABLE` (all caps and wrapped in `\``s) to describe/reference content passed to an option.
* Limit `code spans` to meta-variables. 
  To reference other options or specify literal values, use double quotes.
* Avoid parentheticals unless absolutely necessary.

Examples: 

These show correct/incorrect usage text for the optional `--history-uri` flag:

* Preferred: "Archive history at \`URI\`. Cannot be changed after archival is enabled."
* Avoid incomplete sentences: "_History archival \`URI\`_. Cannot be changed after archival is enabled."
* Avoid wrong verb tenses: "_Archives history at \`URI\`_. Cannot be changed after archival is enabled."
* Avoid missing metavariables: "Archive history at _the specified URI_. Cannot be changed after archival is enabled."
* Avoid unnecessary parenthetical: `Archive history at \`URI\` _(note: cannot be changed after archival is enabled)_.`


COMMAND ENTRY OVERVIEW

A Command entry uses the following format:

    ### <command>: <short-description>
    
    <long description>

    (optional command implementation configuration)

    #### Options

    * `--<long-option>`( , `-<short option>`) (data-type) - <short-description>( <extra-attributes>)
    * `--<long-option>`( , `-<short option>`) (data-type) - <short-description>( <extra-attributes>)
    * `--<long-option>`( , `-<short option>`) (data-type) - <short-description>( <extra-attributes>)

    ...

    optional: Includes options set for [<options-set-name>](#options-set-for-<options-set-link-name>)

COMMAND LISTING

* Use H3 `### <command>: <short-description>` headings for each command.
  * One line only. Use the complete command path with parent commands.
  * Use square-bracked delimited arguments to document positional arguments.
    For example `temporal operator namespace delete [namespace]`.
  * Everything up to ':' or '[' is the command. 
    Square-bracketed positional arguments are not part of the command.
  * Everything following the ':' or '[' is used for the short-description, a concise command explanation.
* A command's long description continues until encountering the H4 (`#### Options`) header.
* At the end of the long description, an optional XML comment configures the command implementation.
  Use one asterisk-delimited bullet per line.
  * `* has-init` - invokes `initCommand` method.
  * `* exact-args=<number>` - Require this exact number of args.
  * `* maximum-args=<number>` - Require this maximum number of args.

  
LONG DESCRIPTION SECTION

* Your text format is preserved literally, so be careful with line lengths.
  Use manual wrapping as needed.

OPTIONS SECTION

* Start the optional Options section with an H4 `#### Options` header line.
* To configure option inclusion, add ` set for <options-set-name>` after `#### Options`.
  For example: `Options set for client:`.
  End every `set for` line with a trailing colon.
* Follow the header declaration with a list of options.
* In the current implementation, you must include at least one option. 
  Although `gen-commands` will complete, the CLI utility will not run.
* To incorporate an existing options set, add a single line below options saying:

  ```
  Includes options set for [<options-set-name>](#options-set-for-<options-set-link-name>).
  ```
    
  For example:
  
  ```
  Includes options set for [client](#options-set-for-client).
  ```
    
  An options set declaration is the equivalent of pasting those options into the bulleted options list.

  Note: Command overlap or similarity does not require option sets.
  Reserve option sets for parallel functionality.
  Copy/paste is fine, otherwise.

OPTION LISTING

* List one option per line, using asterisk-delimited bullets.
* Order the options with most commonly used options first.
* Use this format:

  ```
  `--<long-option>`( , `-<short option>`) (data-type) - <short-description>( <extra-attributes>)
  ```

* Each option listing includes a long option with a double dash and a meaningful name.
* [Optional] A short option uses a single dash and a short string.
  When used, separate the long and short option with a comma and a space.
* Backtick every option and short description. 
  Include the dash or dashes within the ticks.
  For example: `` `--workflow-id`, `-w` ``
* A data type follows option names indicating the required value type for the option.
  The type is `bool`, `duration`, `int`, `string`, `string[]`, `string-enum`, or `timestamp`. (_TODO: more_.)
  Always parenthesize data types.
  For example: `` `--raw` (bool) ``
* A dash follows the data type, with a space on either side. 
* The short description is free-form text and follows the dash.
  Take care not to match trailing attributes. 
  Newline wrapping and/or two-space indention condenses to a single space.
* [Optional] extra attributes include:
  * `Required.` - Marks the option as required.
  * `Default: <default-value>.` - Sets the default value of the option. No default means zero value of the type.
  * `Options: <option>, <option>.` - Sets the possible options for a string enum type.
  * `Env: <env-var>.` - Binds the environment variable to this flag. For example: `Env: TEMPORAL_ADDRESS`.

-->

### temporal: Temporal command-line interface and development server

The Temporal CLI (Command Line Interface) provide a powerful tool to manage, 
monitor, and debug your Temporal applications. It also lets you run a local
Temporal Service directly from your terminal. With this CLI, you can 
start Workflows, pass messages, cancel application steps, and more.

* Start a local development service: 
      `temporal server start-dev` 
* Help messages: pass --help for any command
      `temporal activity complete --help`

Read more: https://docs.temporal.io/cli

<!--
* has-init
-->

#### Options

* `--env` (string) - Active environment name. Default: default. Env: TEMPORAL_ENV.
* `--env-file` (string) - Path to environment settings file (defaults to `$HOME/.config/temporalio/temporal.yaml`).
* `--log-level` (string-enum) - Log level. Default is "info" for most commands and "warn" for `server start-dev`.
  Options: debug, info, warn, error, never. Default: info.
* `--log-format` (string) - Log format. Options are "text" and "json". Default is "text".
* `--output`, `-o` (string-enum) - Non-logging data output format. Options: text, json, jsonl,
  none. Default: text.
* `--time-format` (string-enum) - Time format. Options: relative, iso, raw. Default: relative.
* `--color` (string-enum) - Output coloring. Options: always, never, auto. Default: auto.
* `--no-json-shorthand-payloads` (bool) - Show all payloads as raw, even if they are JSON.

### temporal activity: Complete or fail an Activity

Update an Activity to report that it has completed or failed. This process
marks an activity as successfully finished or as having encountered an error
during execution.

Read more: https://docs.temporal.io/cli/activity

#### Options

Includes options set for [client](#options-set-for-client).

### temporal activity complete: Complete an Activity

Complete an Activity, marking it as successfully finished. Specify 
the ID and include a JSON result to use for the returned value.

```
temporal activity complete \
    --activity-id=YourActivityId \
    --workflow-id=YourWorkflowId \
    --result='{"YourResultKey": "YourResultVal"}'
```

Read more: https://docs.temporal.io/cli/activity#complete

#### Options

* `--activity-id` (string) - Activity `ID` to complete. Required.
* `--result` (string) - Result `JSON` for completing the Activity. Required.
* `--identity` (string) - Identity of the user submitting this request

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal activity fail: Fail an Activity

Fail an Activity, marking it as having encountered an error during execution.
Specify the Activity and Workflow Ids.

```
temporal activity fail \
    --activity-id=YourActivityId \
    --workflow-id=YourWorkflowId
```

Read more: https://docs.temporal.io/cli/activity#fail

#### Options

* `--activity-id` (string) - `ID` of the Activity to be failed. Required.
* `--detail` (string) - `JSON` data describing the reason for failing the Activity
* `--identity` (string) - Identity of the user submitting this request
* `--reason` (string) - Reason for failing the Activity

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal batch: Manage Batch Jobs

A batch job will execute a single command affecting multiple Workflow 
Executions in tandem. These commands include:

* Cancel: Cancel the Workflow Executions specified by the List Filter.
* Signal: Signal the Workflow Executions specified by the List Filter.
* Terminate: Terminates the Workflow Executions specified by the List Filter.

You specify which Workflow Executions to include and the kind of batch 
job to apply. For example, cancel all the running 'YourWorkflow' Workflows:

```
temporal workflow cancel \
  --query 'ExecutionStatus = "Running" AND WorkflowType="YourWorkflow"' \
  --reason "Testing"
 ```

Read more: https://docs.temporal.io/cli/batch

#### Options

Includes options set for [client](#options-set-for-client).

### temporal batch describe: Show Batch Job progress

Show the progress of an ongoing Batch Job. Pass a valid Job ID
to return the job's information:

```
temporal batch describe --job-id=YourJobId
```

Read more: https://docs.temporal.io/cli/batch

#### Options

* `--job-id` (string) - The Batch Job Id to describe. Required.

### temporal batch list: List all Batch Jobs

Return a list of Batch jobs, for the entire Service or a single Namespace.

```
temporal batch list --namespace=YourNamespace
```

Read more: https://docs.temporal.io/cli/batch#list

#### Options

* `--limit` (int) - Show only the first `N` Batch Jobs in the list.

### temporal batch terminate: Terminate a Batch Job

Terminate the Batch job with the provided Job Id. You must provide
a reason motivating the termination, which is stored with the Service
for later reference.

```
temporal batch terminate --job-id=YourJobId --reason=YourTerminationReason
```

Read more: https://docs.temporal.io/cli/batch#terminate

#### Options

* `--job-id` (string) - The Batch Job Id to terminate. Required.
* `--reason` (string) - Reason for terminating the Batch Job. Required.

### temporal env: Manage environments

Environments enable you to create and manage groups of option presets.
They help automate your command configuration. This provides easy set-up
for separate environments, like "dev" and "prod" work. 

You might set an endpoint preset for the `--address` option. 
This enables you to select distinct Temporal Services 
and Namespaces for your commands.

All environments are named so you can refer to them together.
Each environment stores a list of key-value pairs:

```
temporal env set prod.namespace production.f45a2  
temporal env set prod.address production.f45a2.tmprl.cloud:7233  
temporal env set prod.tls-cert-path /temporal/certs/prod.pem  
temporal env set prod.tls-key-path /temporal/certs/prod.key
```

Check your "prod" presets with `temporal env get prod`:

```  
address production.f45a2.tmprl.cloud:7233  
namespace production.f45a2  
tls-cert-path /temporal/certs/prod.pem  
tls-key-path /temporal/certs/prod.key
```

To use the environment with a command, pass `--env` followed by the
environment name. For example, to list workflows in the "prod" environment:

```
$ temporal workflow list --env prod
```

You specify an active environment using the `TEMPORAL_ENV` environment variable. If no environment is specified, the 'default' environment is used.

Read more: https://docs.temporal.io/cli/env

### temporal env delete: Delete an environment or environment property

Remove an environment or a key-value pair within that environment:

```
temporal env delete [environment or property]
```

For example:

```
temporal env delete --env prod
temporal env delete --env prod --key tls-cert-path
```

If you don't specify an environment, you delete the `default` environment:

```
temporal env delete --key tls-cert-path
```

Read more: https://docs.temporal.io/cli/env#delete

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) - The name of the property

### temporal env get: Print environment properties

Prints the environmental properties for a given environment.

```
temporal env get --env environment-name
```

Print all properties of the "prod" environment:

`temporal env get prod`

```
tls-cert-path  /home/my-user/certs/client.cert
tls-key-path   /home/my-user/certs/client.key
address        temporal.example.com:7233
namespace      someNamespace
```

Print a single property:

`temporal env get --env prod --key tls-key-path`

```
tls-key-path  /home/my-user/certs/cluster.key
```

If you do not specify an environment name, you list the `default`
environment properties.

Read more: https://docs.temporal.io/cli/env#get

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) - The name of the property

### temporal env list: Print all environments

STOPPED HERE

List the environments you have set up on your local computer with
`temporal env list`. For example:

```
default 
prod
dev
```

Read more: https://docs.temporal.io/cli/env#list

### temporal env set: Set environment properties

`temporal env set --env environment --key property --value value`

Property names match CLI option names, for example '--address' and '--tls-cert-path':

`temporal env set --env prod --key address --value 127.0.0.1:7233`
`temporal env set --env prod --key tls-cert-path --value /home/my-user/certs/cluster.cert`

If the environment is not specified, the `default` environment is used.

<!--
* maximum-args=2
-->

#### Options

* `--key`, `-k` (string) - The name of the property (required)
* `--value`, `-v` (string) - The value to set the property to (required)

### temporal operator: Manage a Temporal deployment

Operator commands perform actions on Namespaces, Search Attributes, and Temporal Clusters.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal operator cluster: Operations for running a Temporal Cluster

Cluster commands perform actions on Temporal Clusters.

### temporal operator cluster describe: Describe a cluster

#### Options

* `--detail` (bool) - Show extra details

### temporal operator cluster health: Check the health of a cluster

### temporal operator cluster list: List all remote clusters

#### Options

* `--limit` (int) - Show only the first `N` clusters in the list.

### temporal operator cluster remove: Remove a remote cluster

#### Options

* `--name` (string) - Name of cluster. Required.

### temporal operator cluster system: Provide system info

`temporal operator cluster system` command provides information about the system the Cluster is running on. This information can be used to diagnose problems occurring in the Temporal Server.

### temporal operator cluster upsert: Add a remote cluster

`temporal operator cluster upsert` command allows the user to add or update a remote Cluster.

#### Options

* `--frontend-address` (string) - `IP` address to bind the frontend service to. Required.
* `--enable-connection` (bool) - Enable cross-cluster connection.

### temporal operator namespace: Operations on Namespaces

Namespace commands perform operations on Namespaces contained in the Temporal Cluster.

Cluster commands follow this syntax: `temporal operator namespace [command] [command options]`

### temporal operator namespace create: Register a new Namespace

The temporal operator namespace create command creates a new Namespace on the Server.
Namespaces can be created on the active Cluster, or any named Cluster.
`temporal operator namespace create --cluster=YourCluster -n example-1`

Global Namespaces can also be created.
`temporal operator namespace create --global -n example-2`

Other settings, such as retention and Visibility Archival State, can be configured as needed.
For example, the Visibility Archive can be set on a separate URI.
`temporal operator namespace create --retention=5 --visibility-archival-state=enabled --visibility-uri=some-uri -n example-3`

<!--
* maximum-args=1
-->

#### Options

* `--active-cluster` (string) - Active cluster name
* `--cluster` (string[]) - Cluster names. "--cluster" may be passed multiple times to specify multiple clusters.
* `--data` (string) - Namespace data in `KEY=VALUE` format, separated by commas. `KEY` and `VALUE` may be arbitrary strings.
* `--description` (string) - Namespace description
* `--email` (string) - Owner email
* `--global` (bool) - Enable cross-region replication for this namespace.
* `--history-archival-state` (string-enum) - History archival state. Options: disabled, enabled. Default: disabled.
* `--history-uri` (string) - `URI` at which to archive history. Cannot be changed after archival is first enabled.
* `--retention` (duration) - Length of time a closed Workflow is preserved before deletion. Default: 72h.
* `--visibility-archival-state` (string-enum) - Visibility archival state. Options: disabled, enabled. Default: disabled.
* `--visibility-uri` (string) - `URI` at which to archive visibility data. Cannot be changed after archival is first enabled.

### temporal operator namespace delete [namespace]: Delete an existing Namespace

The temporal operator namespace delete command deletes a given Namespace from the system.

<!--
* maximum-args=1
-->

#### Options

* `--yes`, `-y` (bool) - Don't ask for confirmation.

### temporal operator namespace describe [namespace]: Describe a Namespace by its name or ID

The temporal operator namespace describe command provides Namespace information.
Namespaces are identified either by Namespace ID or by name.

`temporal operator namespace describe --namespace-id=some-namespace-id`
`temporal operator namespace describe -n example-namespace-name`

<!--
* maximum-args=1
-->

#### Options

* `--namespace-id` (string) - Namespace ID

### temporal operator namespace list: List all Namespaces

The temporal operator namespace list command lists all Namespaces on the Server.

### temporal operator namespace update: Update a Namespace

The temporal operator namespace update command updates a Namespace.

Namespaces can be assigned a different active Cluster.
`temporal operator namespace update -n namespace --active-cluster=NewActiveCluster`

Namespaces can also be promoted to global Namespaces.
`temporal operator namespace update -n namespace --promote-global`

Any Archives that were previously enabled or disabled can be changed through this command.
However, URI values for archival states cannot be changed after the states are enabled.
`temporal operator namespace update -n namespace --history-archival-state=enabled --visibility-archival-state=disabled`

<!--
* maximum-args=1
-->

#### Options
* `--active-cluster` (string) - Active cluster name
* `--cluster` (string[]) - Cluster names
* `--data` (string[]) - Set a `KEY=VALUE` pair in namespace data. `KEY` and `VALUE` may be arbitrary strings, but JSON is recommended for `VALUE`. May be used multiple times to set multiple pairs.
* `--description` (string) - Namespace description
* `--email` (string) - Owner email
* `--promote-global` (bool) - Enable cross-region replication on this namespace.
* `--history-archival-state` (string-enum) - History archival state. Options: disabled, enabled.
* `--history-uri` (string) - Archive history to this `URI`. Cannot be changed after archival is first enabled.
* `--retention` (duration) - Length of time a closed Workflow is preserved before deletion
* `--visibility-archival-state` (string-enum) - Visibility archival state. Options: disabled, enabled.
* `--visibility-uri` (string) - Archive visibility information to this `URI`. Cannot be changed after archival is first enabled.

### temporal operator search-attribute: Operations for Search Attributes

Search Attribute commands enable operations for the creation, listing, and removal of Search Attributes.

### temporal operator search-attribute create: Add custom Search Attributes

`temporal operator search-attribute create` command adds one or more custom Search Attributes.

#### Options

* `--name` (string[]) - Search Attribute name. Required.
* `--type` (string[]) - Search Attribute type. Options: Text, Keyword, Int, Double, Bool, Datetime, KeywordList. Required.

### temporal operator search-attribute list: List all Search Attributes

`temporal operator search-attribute list` displays a list of all Search Attributes that can be used in list Workflow Queries.

### temporal operator search-attribute remove: Remove custom search attribute metadata

`temporal operator search-attribute remove` command removes custom Search Attribute metadata.

#### Options

* `--name` (string[]) - Search Attribute name. Required.
* `--yes`, `-y` (bool) - Don't ask for confirmation.

### temporal schedule: Perform operations on Schedules

Schedule commands allow the user to create, use, and update Schedules.
Schedules allow starting Workflow Execution at regular times.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal schedule backfill: Backfill a past time range of actions

 The `temporal schedule backfill` command runs the Actions that would have been run in a given time
interval, all at once.

 You can use backfill to fill in Workflow Runs from a time period when the Schedule was paused, from
before the Schedule was created, from the future, or to re-process an interval that was processed.

Schedule backfills require a Schedule ID, along with the time in which to run the Schedule. You can
optionally override the overlap policy. It usually only makes sense to run backfills with either
`BufferAll` or `AllowAll` (other policies will only let one or two runs actually happen).

Example:

```
  temporal schedule backfill           \
    --schedule-id 'your-schedule-id'   \
    --overlap-policy BufferAll         \
    --start-time 2022-05-01T00:00:00Z  \
    --end-time   2022-05-31T23:59:59Z
```

#### Options set for overlap policy:

* `--overlap-policy` (string-enum) - Overlap policy. Options: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll. Default: Skip.

#### Options set for schedule id:

* `--schedule-id`, `-s` (string) - Schedule id. Required.

#### Options

* `--end-time` (timestamp) - Backfill end time. Required.
* `--start-time` (timestamp) - Backfill start time. Required.

### temporal schedule create: Create a new Schedule

The `temporal schedule create` command creates a new Schedule.

Example:

```
  temporal schedule create                                    \
    --schedule-id 'your-schedule-id'                          \
    --calendar '{"dayOfWeek":"Fri","hour":"3","minute":"11"}' \
    --workflow-id 'your-base-workflow-id'                     \
    --task-queue 'your-task-queue'                            \
    --workflow-type 'YourWorkflowType'
```

Any combination of `--calendar`, `--interval`, and `--cron` is supported.
Actions will be executed at any time specified in the Schedule.

#### Options set for schedule configuration:

* `--calendar` (string[]) - Calendar specification in JSON, e.g. `{"dayOfWeek":"Fri","hour":"17","minute":"5"}`
* `--catchup-window` (duration) - Maximum allowed catch-up time if server is down
* `--cron` (string[]) - Calendar spec in cron string format, e.g. `3 11 * * Fri`
* `--end-time` (timestamp) - Overall schedule end time
* `--interval` (string[]) - Interval duration, e.g. 90m, or 90m/13m to include phase offset.
* `--jitter` (duration) - Per-action jitter range
* `--notes` (string) - Initial value of notes field
* `--paused` (bool) - Initial value of paused state
* `--pause-on-failure` (bool) - Pause schedule after any workflow failure.
* `--remaining-actions` (int) - Total number of actions allowed. Zero (default) means unlimited.
* `--start-time` (timestamp) - Overall schedule start time
* `--time-zone` (string) - Interpret all calendar specs in the `TZ` time zone. For a list of time zones, see: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
* `--schedule-search-attribute` (string[]) - Set a Search Attribute for the schedule in `KEY=VALUE` format. `KEY` must be a string identifier (no quotes) and `VALUE` must be a JSON value. May be passed multiple times to set multiple Search Attributes.
* `--schedule-memo` (string[]) - Set a memo for the schedule in `KEY=VALUE` format. `KEY` must be a string identifier (no quotes) and `VALUE` must be a JSON value. May be passed multiple times to set multiple memo values.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

### temporal schedule delete: Delete a Schedule

The `temporal schedule delete` command deletes a Schedule.
Deleting a Schedule does not affect any Workflows started by the Schedule.

If you do also want to cancel or terminate Workflows started by a Schedule, consider using `temporal
workflow delete` with the `TemporalScheduledById` Search Attribute.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule describe: Get Schedule configuration and current state

The `temporal schedule describe` command shows the current configuration of one Schedule,
including information about past, current, and future Workflow Runs.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule list: List all Schedules

The `temporal schedule list` command lists all Schedules in a namespace.

#### Options

* `--long`, `-l` (bool) - Emit detailed information.
* `--really-long` (bool) - Emit even more detailed information that's not usable in table form.

### temporal schedule toggle: Pause or unpause a Schedule

The `temporal schedule toggle` command can pause and unpause a Schedule.

Toggling a Schedule takes a reason. The reason will be set as the `notes` field of the Schedule,
to help with operations communication.

Examples:

* `temporal schedule toggle --schedule-id 'your-schedule-id' --pause --reason "paused because the database is down"`
* `temporal schedule toggle --schedule-id 'your-schedule-id' --unpause --reason "the database is back up"`

#### Options

* `--pause` (bool) - Pauses the schedule.
* `--reason` (string) - Reason for pausing/unpausing. Default: "(no reason provided)".
* `--unpause` (bool) - Unpauses the schedule.

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule trigger: Trigger a schedule to run immediately

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).

### temporal schedule update: Update a Schedule with a new definition

The temporal schedule update command updates an existing Schedule. It replaces the entire
configuration of the schedule, including spec, action, and policies.

#### Options

Includes options set for [schedule-configuration](#options-set-for-schedule-configuration).
Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

### temporal server: Run Temporal Server

Start a development version of [Temporal Server](/concepts/what-is-the-temporal-server):

`temporal server start-dev`

### temporal server start-dev: Start Temporal development server

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
* `--ip` (string) - IP address to bind the frontend service to. Default: localhost.
* `--ui-ip` (string) - IP address to bind the Web UI to. Default is same as --ip.
* `--ui-asset-path` (string) - UI custom assets path.
* `--ui-codec-endpoint` (string) - UI remote codec HTTP endpoint.
* `--sqlite-pragma` (string[]) - Specify SQLite pragma statements in pragma=value format.
* `--dynamic-config-value` (string[]) - Dynamic config value, as KEY=JSON_VALUE (string values need quotes).
* `--log-config` (bool) - Log the server config being used to stderr.

### temporal task-queue: Manage Task Queues

Task Queue commands allow operations to be performed on [Task Queues](/concepts/what-is-a-task-queue). To run a Task
Queue command, run `temporal task-queue [command] [command options]`.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal task-queue describe: Show Workers that have recently polled on a Task Queue

The `temporal task-queue describe` command provides [poller](/application-development/worker-performance#poller-count)
information for a given [Task Queue](/concepts/what-is-a-task-queue).

The [Server](/concepts/what-is-the-temporal-server) records the last time of each poll request. A `LastAccessTime` value
in excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the
Worker has shut down. [Workers](/concepts/what-is-a-worker) are removed if 5 minutes have passed since the last poll
request.

Information about the Task Queue can be returned to troubleshoot server issues.

`temporal task-queue describe --task-queue=YourTaskQueue --task-queue-type="activity"`

Use the options listed below to modify what this command returns.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.
* `--task-queue-type` (string-enum) - Task Queue type. Options: workflow, activity. Default: workflow.
* `--partitions` (int) - Query for all partitions up to this number (experimental+temporary feature). Default: 1.

### temporal task-queue get-build-id-reachability: Show which Build IDs are available on a Task Queue

This command can tell you whether or not Build IDs may be used for new, existing, or closed workflows. Both the '--build-id' and '--task-queue' flags may be specified multiple times. If you do not provide a task queue, reachability for the provided Build IDs will be checked against all task queues.

#### Options

* `--build-id` (string[]) - Which Build ID to get reachability information for. May be specified multiple times.
* `--reachability-type` (string-enum) - Specify how you'd like to filter the reachability of Build IDs. Valid choices are `open` (reachable by one or more open workflows), `closed` (reachable by one or more closed workflows), or `existing` (reachable by either). If a Build ID is reachable by new workflows, that is always reported. Options: open, closed, existing. Default: existing.
* `--task-queue`, `-t` (string[]) - Which Task Queue(s) to constrain the reachability search to. May be specified multiple times.

### temporal task-queue get-build-ids: Show worker Build ID versions on a Task Queue

Fetch the sets of compatible build IDs associated with a Task Queue and associated information.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.
* `--max-sets` (int) - Limits how many compatible sets will be returned. Specify 1 to only return the current default major version set. 0 returns all sets. (default: 0). Default: 0.

### temporal task-queue list-partition: List a Task Queue's partitions

The temporal task-queue list-partition command displays the partitions of a Task Queue, along with the matching node they are assigned to.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.

### temporal task-queue update-build-ids: Operations to manage Build ID versions on a Task Queue

Provides various commands for adding or changing the sets of compatible build IDs associated with a Task Queue. See the help of each sub-command for more.

### temporal task-queue update-build-ids add-new-compatible: Add a new build ID compatible with an existing ID to a Task Queue's version sets

The new build ID will become the default for the set containing the existing ID. See per-flag help for more.

#### Options

* `--build-id` (string) - The new build id to be added. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.
* `--existing-compatible-build-id` (string) - A build id which must already exist in the version sets known by the task queue. The new id will be stored in the set containing this id, marking it as compatible with the versions within. Required.
* `--set-as-default` (bool) - When set, establishes the compatible set being targeted as the overall default for the queue. If a different set was the current default, the targeted set will replace it as the new default. Defaults to false.

### temporal task-queue update-build-ids add-new-default: Add a new default (incompatible) build ID to a Task Queue's version sets

Creates a new build id set which will become the new overall default for the queue with the provided build id as its only member. This new set is incompatible with all previous sets/versions.

#### Options

* `--build-id` (string) - The new build id to be added. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.

### temporal task-queue update-build-ids promote-id-in-set: Promote a build ID to become the default for its containing set

New tasks compatible with the set will be dispatched to the default id.

#### Options

* `--build-id` (string) - An existing build id which will be promoted to be the default inside its containing set. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.

### temporal task-queue update-build-ids promote-set: Promote a build ID set to become the default for a Task Queue

If the set is already the default, this command has no effect.

#### Options

* `--build-id` (string) - An existing build id whose containing set will be promoted. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.


### temporal workflow: Start, list, and operate on Workflows

[Workflow](/concepts/what-is-a-workflow) commands perform operations on [Workflow Executions](/concepts/what-is-a-workflow-execution).

Workflow commands use this syntax: `temporal workflow COMMAND [ARGS]`.

#### Options set for client:

* `--address` (string) - Temporal server address. Default: 127.0.0.1:7233. Env: TEMPORAL_ADDRESS.
* `--namespace`, `-n` (string) - Temporal server namespace. Default: default. Env: TEMPORAL_NAMESPACE.
* `--api-key` (string) - Sets the API key on requests. Env: TEMPORAL_API_KEY.
* `--grpc-meta` (string[]) - HTTP headers to send with requests (formatted as key=value).
* `--tls` (bool) - Enable TLS encryption without additional options such as mTLS or client certificates. Env:
  TEMPORAL_TLS.
* `--tls-cert-path` (string) - Path to x509 certificate. Env: TEMPORAL_TLS_CERT.
* `--tls-key-path` (string) - Path to private certificate key. Env: TEMPORAL_TLS_KEY.
* `--tls-ca-path` (string) - Path to server CA certificate. Env: TEMPORAL_TLS_CA.
* `--tls-cert-data` (string) - Data for x509 certificate. Exclusive with -path variant. Env: TEMPORAL_TLS_CERT_DATA.
* `--tls-key-data` (string) - Data for private certificate key. Exclusive with -path variant. Env: TEMPORAL_TLS_KEY_DATA.
* `--tls-ca-data` (string) - Data for server CA certificate. Exclusive with -path variant. Env: TEMPORAL_TLS_CA_DATA.
* `--tls-disable-host-verification` (bool) - Disables TLS host-name verification. Env:
  TEMPORAL_TLS_DISABLE_HOST_VERIFICATION.
* `--tls-server-name` (string) - Overrides target TLS server name. Env: TEMPORAL_TLS_SERVER_NAME.
* `--codec-endpoint` (string) - Endpoint for a remote Codec Server. Env: TEMPORAL_CODEC_ENDPOINT.
* `--codec-auth` (string) - Sets the authorization header on requests to the Codec Server. Env: TEMPORAL_CODEC_AUTH.

### temporal workflow cancel: Cancel a Workflow Execution

The `temporal workflow cancel` command is used to cancel a [Workflow Execution](/concepts/what-is-a-workflow-execution).
Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new
Command Task will be scheduled, and the Workflow Execution will perform cleanup work.

Executions may be cancelled by [ID](/concepts/what-is-a-workflow-id):
```
temporal workflow cancel --workflow-id YourWorkflowId
```

...or in bulk via a visibility query [list filter](/concepts/what-is-a-list-filter):
```
temporal workflow cancel --query=YourQuery
```

Use the options listed below to change the behavior of this command.

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch)

### temporal workflow count: Count Workflow Executions

The `temporal workflow count` command returns a count of [Workflow Executions](/concepts/what-is-a-workflow-execution).

Use the options listed below to change the command's behavior.

#### Options

* `--query`, `-q` (string) - Filter results using a SQL-like query.

### temporal workflow delete: Delete a Workflow Execution

The `temporal workflow delete` command is used to delete a specific [Workflow Execution](/concepts/what-is-a-workflow-execution).
This asynchronously deletes a workflow's [Event History](/concepts/what-is-an-event-history).
If the [Workflow Execution](/concepts/what-is-a-workflow-execution) is Running, it will be terminated before deletion.

```
temporal workflow delete \
		--workflow-id YourWorkflowId \
```

Use the options listed below to change the command's behavior.

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch)

### temporal workflow describe: Show information about a Workflow Execution

The `temporal workflow describe` command shows information about a given
[Workflow Execution](/concepts/what-is-a-workflow-execution).

This information can be used to locate Workflow Executions that weren't able to run successfully.

`temporal workflow describe --workflow-id=meaningful-business-id`

Output can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.

`temporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true`

Use the command options below to change the information returned by this command.

#### Options set for workflow reference

* `--workflow-id`, `-w` (string) - Workflow Id. Required.
* `--run-id`, `-r` (string) - Run Id.

#### Options

* `--reset-points` (bool) - Only show auto-reset points.
* `--raw` (bool) - Print properties without changing their format.

### temporal workflow execute: Start a new Workflow Execution and prints its progress

The `temporal workflow execute` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution) and
prints its progress. The command completes when the Workflow Execution completes.

Single quotes('') are used to wrap input as JSON.

```
temporal workflow execute
		--workflow-id meaningful-business-id \
		--type YourWorkflow \
		--task-queue YourTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options

* `--event-details` (bool) - If set when using text output, this will print the event details instead of just the event
  during workflow progress. If set when using JSON output, this will include the entire "history" JSON key of the
  started run (does not follow runs).

Includes options set for [shared workflow start](#options-set-for-shared-workflow-start).
Includes options set for [workflow start](#options-set-for-workflow-start).
Includes options set for [payload input](#options-set-for-payload-input).

### temporal workflow fix-history-json: Updates an event history JSON file to the current format

```
temporal workflow fix-history-json \
	--source original.json \
	--target reserialized.json
```

Use the options listed below to change the command's behavior.

#### Options

* `--source`, `-s` (string) - Path to the input file. Required.
* `--target`, `-t` (string) - Path to the output file, or standard output if not set.

### temporal workflow list: List Workflow Executions based on a Query

The `temporal workflow list` command provides a list of [Workflow Executions](/concepts/what-is-a-workflow-execution)
that meet the criteria of a given [Query](/concepts/what-is-a-query).
By default, this command returns up to 10 closed Workflow Executions.

`temporal workflow list --query=YourQuery`

The command can also return a list of archived Workflow Executions.

`temporal workflow list --archived`

Use the command options below to change the information returned by this command.

#### Options

* `--query`, `-q` (string) - Filter results using a SQL-like query.
* `--archived` (bool) - If set, will only query and list archived workflows instead of regular workflows.
* `--limit` (int) - Limit the number of items to print.

### temporal workflow query: Query a Workflow Execution

The `temporal workflow query` command is used to [Query](/concepts/what-is-a-query) a
[Workflow Execution](/concepts/what-is-a-workflow-execution)
by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow query \
		--workflow-id YourWorkflowId \
		--name YourQuery \
		--input '{"YourInputKey": "YourInputValue"}'
```

Use the options listed below to change the command's behavior.

#### Options

* `--type` (string) - Query Type/Name. Required.
* `--reject-condition` (string-enum) - Optional flag for rejecting Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [payload input](#options-set-for-payload-input).
Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow reset: Reset a Workflow Execution to an older point in history

The temporal workflow reset command resets a [Workflow Execution](/concepts/what-is-a-workflow-execution).
A reset allows the Workflow to resume from a certain point without losing its parameters or [Event History](/concepts/what-is-an-event-history).

The Workflow Execution can be set to a given [Event Type](/concepts/what-is-an-event):
```
temporal workflow reset --workflow-id=meaningful-business-id --type=LastContinuedAsNew
```

...or a specific any Event after `WorkflowTaskStarted`.
```
temporal workflow reset --workflow-id=meaningful-business-id --event-id=YourLastEvent
```
For batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow Id, run Id and event Id
should not be set.
Use the options listed below to change reset behavior.

#### Options

* `--workflow-id`, `-w` (string) - Workflow Id. Required for non-batch reset operations.
* `--run-id`, `-r` (string) - Run Id.
* `--event-id`, `-e` (int) - The Event Id for any Event after `WorkflowTaskStarted` you want to reset to (exclusive). It can be `WorkflowTaskCompleted`, `WorkflowTaskFailed` or others.
* `--reason` (string) - The reason why this workflow is being reset. Required.
* `--reapply-type` (string-enum) - Event types to reapply after the reset point. Options: All, Signal, None. Default: All.
* `--type`, `-t` (string-enum) - Event type to which you want to reset. Options: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.
* `--build-id` (string) - Only used if type is BuildId. Reset the first workflow task processed by this build id. Note that by default, this reset is allowed to be to a prior run in a chain of continue-as-new.
* `--query`, `-q` (string) - Start a batch reset to operate on Workflow Executions with given List Filter.
* `--yes`, `-y` (bool) - Don't ask for confirmation. (Note: Only allowed if --query is present)



### temporal workflow show: Show Event History for a Workflow Execution

The `temporal workflow show` command provides the [Event History](/concepts/what-is-an-event-history) for a
[Workflow Execution](/concepts/what-is-a-workflow-execution). With JSON output specified, this output can be given to
an SDK to perform a replay.

Use the options listed below to change the command's behavior.

#### Options

* `--follow`, `-f` (bool) - Follow the progress of a Workflow Execution in real time (does not apply
  to JSON output).

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow signal: Signal a Workflow Execution

The `temporal workflow signal` command is used to [Signal](/concepts/what-is-a-signal) a
[Workflow Execution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow signal \
		--workflow-id YourWorkflowId \
		--name YourSignal \
		--input '{"YourInputKey": "YourInputValue"}'
```

Use the options listed below to change the command's behavior.

#### Options

* `--name` (string) - Signal Name. Required.

Includes options set for [payload input](#options-set-for-payload-input).

#### Options set for single workflow or batch:

* `--workflow-id`, `-w` (string) - Workflow Id. Either this or --query must be set.
* `--run-id`, `-r` (string) - Run Id. Cannot be set when --query is set.
* `--query`, `-q` (string) - Start a batch to operate on Workflow Executions with given List Filter. Either --query or --workflow-id must be set.
* `--reason` (string) - Reason to perform batch. Only allowed if query is present unless the command specifies
  otherwise. Defaults to message with the current user's name.
* `--yes`, `-y` (bool) - Don't ask for confirmation. (Note: Only allowed if --query is present)

### temporal workflow stack: Show the stack trace of a Workflow Execution

The `temporal workflow stack` command [Queries](/concepts/what-is-a-query) a
[Workflow Execution](/concepts/what-is-a-workflow-execution) with `__stack_trace` as the query type.
This returns a stack trace of all the threads or routines currently used by the workflow, and is
useful for troubleshooting.

```
temporal workflow stack --workflow-id YourWorkflowId
```

Use the options listed below to change the command's behavior.

#### Options

* `--reject-condition` (string-enum) - Optional flag for rejecting Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow start: Start a new Workflow Execution

The `temporal workflow start` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution). The
Workflow and Run IDs are returned after starting the [Workflow](/concepts/what-is-a-workflow).

```
temporal workflow start \
		--workflow-id meaningful-business-id \
		--type YourWorkflow \
		--task-queue YourTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options set for shared workflow start:

* `--workflow-id`, `-w` (string) - Workflow Id
* `--type` (string) - Workflow Type name. Required.
* `--task-queue`, `-t` (string) - Workflow Task queue. Required.
* `--run-timeout` (duration) - Fail a Workflow Run if it takes longer than `DURATION`.
* `--execution-timeout` (duration) - Fail a WorkflowExecution if it takes longer than `DURATION`, including retries and ContinueAsNew tasks.
* `--task-timeout` (duration) - Fail a Workflow Task if it takes longer than `DURATION`. (Start-to-close timeout for a Workflow Task.) Default: 10s.
* `--search-attribute` (string[]) - Passes Search Attribute in key=value format. Use valid JSON formats for value.
* `--memo` (string[]) - Passes Memo in key=value format. Use valid JSON formats for value.

#### Options set for workflow start:

* `--cron` (string) - Cron schedule for the Workflow. Deprecated - use schedules instead.
* `--fail-existing` (bool) - Fail if the Workflow already exists.
* `--start-delay` (duration) - Wait before starting the Workflow. Cannot be used with a cron schedule. If the
  Workflow receives a signal or update before the delay has elapsed, it will start immediately.
* `--id-reuse-policy` (string) - Allow the same Workflow Id to be used in a new Workflow Execution. Options:
  AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.

#### Options set for payload input:

* `--input`, `-i` (string[]) - Input value (default JSON unless --input-meta is non-JSON encoding). Can
  be passed multiple times for multiple arguments. Cannot be combined with --input-file.
* `--input-file` (string[]) - Read `PATH` as the input value (JSON by default unless --input-meta is non-JSON
  encoding). Can be passed multiple times for multiple arguments. Cannot be combined with --input.
* `--input-meta` (string[]) - Metadata for the input payload, specified as a `KEY=VALUE` pair. If KEY is "encoding", overrides the
  default of "json/plain". Pass multiple --input-meta options to set multiple pairs.
* `--input-base64` (bool) - Assume inputs are base64-encoded and attempt to decode them.

### temporal workflow terminate: Terminate a Workflow Execution

The `temporal workflow terminate` command is used to terminate a
[Workflow Execution](/concepts/what-is-a-workflow-execution). Canceling a running Workflow Execution records a
`WorkflowExecutionTerminated` event as the closing Event in the workflow's Event History. Workflow code is oblivious to
termination. Use `temporal workflow cancel` if you need to perform cleanup in your workflow.

Executions may be terminated by [ID](/concepts/what-is-a-workflow-id) with an optional reason:
```
temporal workflow terminate [--reason my-reason] --workflow-id YourWorkflowId
```

...or in bulk via a visibility query [list filter](/concepts/what-is-a-list-filter):
```
temporal workflow terminate --query=YourQuery
```

Use the options listed below to change the behavior of this command.

#### Options

* `--workflow-id`, `-w` (string) - Workflow Id. Either this or query must be set.
* `--run-id`, `-r` (string) - Run Id. Cannot be set when query is set.
* `--query`, `-q` (string) - Start a batch to terminate Workflow Executions with the `QUERY` List Filter. Either this or
  Workflow Id must be set.
* `--reason` (string) - Reason for termination. Defaults to message with the current user's name.
* `--yes`, `-y` (bool) - Confirm prompt to perform batch. Only allowed if query is present.

### temporal workflow trace: Interactively show the progress of a Workflow Execution

The `temporal workflow trace` command displays the progress of a [Workflow Execution](/concepts/what-is-a-workflow-execution) and its child workflows with a real-time trace.
This view provides a great way to understand the flow of a workflow.

Use the options listed below to change the behavior of this command.

#### Options

* `--fold` (string[]) - Fold Child Workflows with the specified `STATUS`. To specify multiple statuses, pass --fold multiple times. This will reduce the amount of information fetched and displayed. Case-insensitive. Ignored if --no-fold supplied. Available values: running, completed, failed, canceled, terminated, timedout, continueasnew.
* `--no-fold` (bool) - Disable folding. All Child Workflows within the set depth will be fetched and displayed.
* `--depth` (int) - Fetch up to N Child Workflows deep. Use -1 to fetch child workflows at any depth. Default: -1.
* `--concurrency` (int) - Fetch up to N Workflow Histories at a time. Default: 10.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow update: Update a running workflow synchronously

The `temporal workflow update` command is used to synchronously [Update](/concepts/what-is-an-update) a
[WorkflowExecution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow update \
		--workflow-id YourWorkflowId \
		--name YourUpdate \
		--input '{"Input": "As-JSON"}'
```

Use the options listed below to change the command's behavior.

#### Options

* `--name` (string) - Update Name. Required.
* `--workflow-id`, `-w` (string) - Workflow `ID`. Required.
* `--update-id` (string) - Update `ID`. If unset, default to a UUID.
* `--run-id`, `-r` (string) - Run `ID`. If unset, the currently running Workflow Execution receives the Update.
* `--first-execution-run-id` (string) - Send the Update to the last Workflow Execution in the chain that started
  with `ID`.

Includes options set for [payload input](#options-set-for-payload-input).
