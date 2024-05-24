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
      * `* maximum-args=<number>` - Require this maximum number of args
  * Can have `#### Options` or `#### Options set for <options-set-name>` which can have options.
    * Can have bullets
      * Each bullet is `* <option-names> (<data-type>) - <short-description>. <extra-attributes>`.
      * `<option-names>` is `` `--<option-name>` `` and can optionally be followed by ``, `-<short-name>` ``.
      * `<data-type>` must be one of `bool`, `duration`, `int`, `string`, `string[]`, `string-enum`, `timestamp`, TODO: more
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
* `--log-level` (string-enum) - Log level. Default is "info" for most commands and "warn" for `server start-dev`.
  Options: debug, info, warn, error, never. Default: info.
* `--log-format` (string) - Log format. Options are "text" and "json". Default is "text".
* `--output`, `-o` (string-enum) - Data output format. Note, this does not affect logging. Options: text, json, jsonl,
  none. Default: text.
* `--time-format` (string-enum) - Time format. Options: relative, iso, raw. Default: relative.
* `--color` (string-enum) - Set coloring. Options: always, never, auto. Default: auto.
* `--no-json-shorthand-payloads` (bool) - Always show all payloads as raw payloads even if they are JSON.

### temporal activity: Complete or fail an Activity.

#### Options

Includes options set for [client](#options-set-for-client).


### temporal activity complete: Complete an Activity.

Complete an Activity.

`temporal activity complete --activity-id=MyActivityId --workflow-id=MyWorkflowId --result='{"MyResultKey": "MyResultVal"}'`

#### Options

* `--activity-id` (string) - The Activity to be completed. Required.
* `--identity` (string) - Identity of user submitting this request.
* `--result` (string) - The result with which to complete the Activity (JSON). Required.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal activity fail: Fail an Activity.

Fail an Activity.

`temporal activity fail --activity-id=MyActivityId --workflow-id=MyWorkflowId`

#### Options

* `--activity-id` (string) - The Activity to be failed. Required.
* `--detail` (string) - JSON data describing reason for failing the Activity.
* `--identity` (string) - Identity of user submitting this request.
* `--reason` (string) - Reason for failing the Activity.

Includes options set for [workflow reference](#options-set-for-workflow-reference).


### temporal batch: Manage Batch Jobs

Batch commands change multiple Workflow Executions.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal batch describe: Show Batch Job progress.

The temporal batch describe command shows the progress of an ongoing Batch Job.

`temporal batch describe --job-id=MyJobId`

#### Options

* `--job-id` (string) - The Batch Job Id to describe. Required.

### temporal batch list: List all Batch Jobs

The temporal batch list command returns all Batch Jobs.
Batch Jobs can be returned for an entire Cluster or a single Namespace.

`temporal batch list --namespace=MyNamespace`

#### Options

* `--limit` (int) - Limit the number of items to print.

### temporal batch terminate: Terminate a Batch Job

The temporal batch terminate command terminates a Batch Job with the provided Job Id.
For future reference, provide a reason for terminating the Batch Job.

`temporal batch terminate --job-id=MyJobId --reason=JobReason`

#### Options

* `--job-id` (string) - The Batch Job Id to terminate. Required.
* `--reason` (string) - Reason for terminating the Batch Job. Required.

### temporal env: Manage environments.

Use the '--env <env name>' option with other commands to point the CLI at a different Temporal Server instance. If --env
is not passed, the 'default' environment is used.

### temporal env delete: Delete an environment or environment property.

`temporal env delete --env environment [-k property]`

Delete an environment or just a single property:

`temporal env delete --env prod`
`temporal env delete --env prod -k tls-cert-path`

If the environment is not specified, the `default` environment is deleted:

`temporal env delete -k tls-cert-path`

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) - The name of the property.

### temporal env get: Print environment properties.

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

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) - The name of the property.

### temporal env list: Print all environments.

List all environments.

### temporal env set: Set environment properties.

`temporal env set --env environment -k property -v value`

Property names match CLI option names, for example '--address' and '--tls-cert-path':

`temporal env set --env prod -k address -v 127.0.0.1:7233`
`temporal env set --env prod -k tls-cert-path -v /home/my-user/certs/cluster.cert`

If the environment is not specified, the `default` environment is used.

<!--
* maximum-args=2
-->

#### Options

* `--key`, `-k` (string) - The name of the property.
* `--value`, `-v` (string) - The value to set the property to.

### temporal operator: Manage a Temporal deployment.

Operator commands enable actions on Namespaces, Search Attributes, and Temporal Clusters. These actions are performed through subcommands.

To run an Operator command, `run temporal operator [command] [subcommand] [command options]`

#### Options

Includes options set for [client](#options-set-for-client).

### temporal operator cluster: Operations for running a Temporal Cluster.

Cluster commands enable actions on Temporal Clusters.

Cluster commands follow this syntax: `temporal operator cluster [command] [command options]`

### temporal operator cluster describe: Describe a cluster

`temporal operator cluster describe` command shows information about the Cluster.

#### Options

* `--detail` (bool) - Prints extra details.

### temporal operator cluster health: Checks the health of a cluster

`temporal operator cluster health` command checks the health of the Frontend Service.

### temporal operator cluster list: List all clusters

`temporal operator cluster list` command prints a list of all remote Clusters on the system.

#### Options

* `--limit` (int) - Limit the number of items to print.

### temporal operator cluster remove: Remove a cluster

`temporal operator cluster remove` command removes a remote Cluster from the system.

#### Options

* `--name` (string) - Name of cluster. Required.

### temporal operator cluster system: Provide system info

`temporal operator cluster system` command provides information about the system the Cluster is running on. This information can be used to diagnose problems occurring in the Temporal Server.

### temporal operator cluster upsert: Add a remote

`temporal operator cluster upsert` command allows the user to add or update a remote Cluster.

#### Options

* `--frontend-address` (string) - IP address to bind the frontend service to. Required.
* `--enable-connection` (bool) - enable cross cluster connection.

### temporal operator namespace: Operations performed on Namespaces.

Namespace commands perform operations on Namespaces contained in the Temporal Cluster.

Cluster commands follow this syntax: `temporal operator namespace [command] [command options]`

### temporal operator namespace create: Registers a new Namespace.

The temporal operator namespace create command creates a new Namespace on the Server.
Namespaces can be created on the active Cluster, or any named Cluster.
`temporal operator namespace create --cluster=MyCluster -n example-1`

Global Namespaces can also be created.
`temporal operator namespace create --global -n example-2`

Other settings, such as retention and Visibility Archival State, can be configured as needed.
For example, the Visibility Archive can be set on a separate URI.
`temporal operator namespace create --retention=5 --visibility-archival-state=enabled --visibility-uri=some-uri -n example-3`

<!--
* maximum-args=1
-->

#### Options

* `--active-cluster` (string) - Active cluster name.
* `--cluster` (string[]) - Cluster names.
* `--data` (string) - Namespace data in key=value format. Use JSON for values.
* `--description` (string) - Namespace description.
* `--email` (string) - Owner email.
* `--global` (bool) - Whether the namespace is a global namespace.
* `--history-archival-state` (string-enum) - History archival state. Options: disabled, enabled. Default: disabled.
* `--history-uri` (string) - Optionally specify history archival URI (cannot be changed after first time archival is enabled).
* `--retention` (duration) - Length of time a closed Workflow is preserved before deletion. Default: 72h.
* `--visibility-archival-state` (string-enum) - Visibility archival state. Options: disabled, enabled. Default: disabled.
* `--visibility-uri` (string) - Optionally specify visibility archival URI (cannot be changed after first time archival is enabled).

### temporal operator namespace delete [namespace]: Deletes an existing Namespace.

The temporal operator namespace delete command deletes a given Namespace from the system.

<!--
* maximum-args=1
-->

#### Options

* `--yes`, `-y` (bool) - Confirm prompt to perform deletion.

### temporal operator namespace describe [namespace]: Describe a Namespace by its name or ID.

The temporal operator namespace describe command provides Namespace information.
Namespaces are identified either by Namespace ID or by name.

`temporal operator namespace describe --namespace-id=some-namespace-id`
`temporal operator namespace describe -n example-namespace-name`

<!--
* maximum-args=1
-->

#### Options

* `--namespace-id` (string) -  Namespace ID.

### temporal operator namespace list:  List all Namespaces.

The temporal operator namespace list command lists all Namespaces on the Server.

### temporal operator namespace update: Updates a Namespace.

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
* `--active-cluster` (string) - Active cluster name.
* `--cluster` (string[]) - Cluster names.
* `--data` (string[]) - Namespace data in key=value format. Use JSON for values.
* `--description` (string) - Namespace description.
* `--email` (string) - Owner email.
* `--promote-global` (bool) - Promote local namespace to global namespace.
* `--history-archival-state` (string-enum) - History archival state. Options: disabled, enabled.
* `--history-uri` (string) - Optionally specify history archival URI (cannot be changed after first time archival is enabled).
* `--retention` (duration) - Length of time a closed Workflow is preserved before deletion.
* `--visibility-archival-state` (string-enum) - Visibility archival state. Options: disabled, enabled.
* `--visibility-uri` (string) - Optionally specify visibility archival URI (cannot be changed after first time archival is enabled).

### temporal operator search-attribute: Operations applying to Search Attributes

Search Attribute commands enable operations for the creation, listing, and removal of Search Attributes.

### temporal operator search-attribute create: Adds one or more custom Search Attributes

`temporal operator search-attribute create` command adds one or more custom Search Attributes.

#### Options

* `--name` (string[]) - Search Attribute name. Required.
* `--type` (string[]) - Search Attribute type. Options: Text, Keyword, Int, Double, Bool, Datetime, KeywordList. Required.

### temporal operator search-attribute list: Lists all Search Attributes that can be used in list Workflow Queries

`temporal operator search-attribute list` displays a list of all Search Attributes.

### temporal operator search-attribute remove: Removes custom search attribute metadata only

`temporal operator search-attribute remove` command removes custom Search Attribute metadata.

#### Options

* `--name` (string[]) - Search Attribute name. Required.
* `--yes`, `-y` (bool) - Confirm prompt to perform deletion.

### temporal schedule: Perform operations on Schedules.

Schedule commands allow the user to create, use, and update Schedules.
Schedules allow starting Workflow Execution at regular times.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal schedule backfill: Backfills a past time range of actions.

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

### temporal schedule create: Create a new Schedule.

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

* `--calendar` (string[]) - Calendar specification in JSON, e.g. `{"dayOfWeek":"Fri","hour":"17","minute":"5"}`.
* `--catchup-window` (duration) - Maximum allowed catch-up time if server is down.
* `--cron` (string[]) - Calendar spec in cron string format, e.g. `3 11 * * Fri`.
* `--end-time` (timestamp) - Overall schedule end time.
* `--interval` (string[]) - Interval duration, e.g. 90m, or 90m/13m to include phase offset.
* `--jitter` (duration) - Per-action jitter range.
* `--notes` (string) - Initial value of notes field.
* `--paused` (bool) - Initial value of paused state.
* `--pause-on-failure` (bool) - Pause schedule after any workflow failure.
* `--remaining-actions` (int) - Total number of actions allowed. Zero (default) means unlimited.
* `--start-time` (timestamp) - Overall schedule start time.
* `--time-zone` (string) - Time zone to interpret all calendar specs in (IANA name).
* `--schedule-search-attribute` (string[]) - Search Attribute for the _schedule_ in key=value format. Use valid JSON formats for value.
* `--schedule-memo` (string[]) - Memo for the _schedule_ in key=value format. Use valid JSON formats for value.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

### temporal schedule delete: Deletes a Schedule.

The `temporal schedule delete` command deletes a Schedule.
Deleting a Schedule does not affect any Workflows started by the Schedule.

If you do also want to cancel or terminate Workflows started by a Schedule, consider using `temporal
workflow delete` with the `TemporalScheduledById` Search Attribute.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule describe: Get Schedule configuration and current state.

The `temporal schedule describe` command shows the current configuration of one Schedule,
including information about past, current, and future Workflow Runs.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule list: Lists Schedules.

The `temporal schedule list` command lists all Schedules in a namespace.

#### Options

* `--long`, `-l` (bool) - Include detailed information.
* `--really-long` (bool) - Include even more detailed information that's not really usable in table form.

### temporal schedule toggle: Pauses or unpauses a Schedule.

The `temporal schedule toggle` command can pause and unpause a Schedule.

Toggling a Schedule takes a reason. The reason will be set as the `notes` field of the Schedule,
to help with operations communication.

Examples:

* `temporal schedule toggle --schedule-id 'your-schedule-id' --pause --reason "paused because the database is down"`
* `temporal schedule toggle --schedule-id 'your-schedule-id' --unpause --reason "the database is back up"`

#### Options

* `--pause` (bool) - Pauses the schedule.
* `--reason` (string) - Reason for pausing/unpausing. Default: "(no reason provided)".
* `--unpause` (bool) - Pauses the schedule.

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule trigger: Triggers a schedule to take an action immediately.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).

### temporal schedule update: Updates a Schedule with a new definition.

The temporal schedule update command updates an existing Schedule. It replaces the entire
configuration of the schedule, including spec, action, and policies.

#### Options

Includes options set for [schedule-configuration](#options-set-for-schedule-configuration).
Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

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
* `--ip` (string) - IP address to bind the frontend service to. Default: localhost.
* `--ui-ip` (string) - IP address to bind the Web UI to. Default is same as --ip.
* `--ui-asset-path` (string) - UI custom assets path.
* `--ui-codec-endpoint` (string) - UI remote codec HTTP endpoint.
* `--sqlite-pragma` (string[]) - Specify SQLite pragma statements in pragma=value format.
* `--dynamic-config-value` (string[]) - Dynamic config value, as KEY=JSON_VALUE (string values need quotes).
* `--log-config` (bool) - Log the server config being used to stderr.

### temporal task-queue: Manage Task Queues.

Task Queue commands allow operations to be performed on [Task Queues](/concepts/what-is-a-task-queue). To run a Task
Queue command, run `temporal task-queue [command] [command options]`.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal task-queue describe: Provides task reachability and pollers information for Workers on this Task Queue.

The `temporal task-queue describe` command provides task reachability information for the requested versions and all task types,
which can be used to safely retire Workers with old code versions, provided that they were assigned a Build ID.

The reachability states of a Build ID are:
    - `Reachable`: the Build ID may be used by new workflows or activities
(based on versioning rules), or there are open workflows or backlogged activities assigned to it.
    - `ClosedWorkflowsOnly`: the Build ID does not have open workflows, and is not reachable by new workflows, but MAY have closed workflows within the namespace retention period.
    - `Unreachable`: indicates that this Build ID is not used for new executions, nor has been used by any existing execution within the retention period.

Task reachability is eventually consistent; there may be a delay until it converges to the most
accurate value but it is designed in a way to take the more conservative side until it converges.
For example, `Reachable` is more conservative than `ClosedWorkflowsOnly`.

There is a non-trivial cost of computing task reachability, use the flag `--report-reachability` to enable it.

This command also provides [poller](/application-development/worker-performance#poller-count)
information for a given [Task Queue](/concepts/what-is-a-task-queue).

The [Server](/concepts/what-is-the-temporal-server) records the last time of each poll request. A `LastAccessTime` value
in excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the
Worker has shut down. [Workers](/concepts/what-is-a-worker) are removed if 5 minutes have passed since the last poll
request.

Information about the Task Queue can be returned to troubleshoot server issues.

Use the options listed below to modify what this command returns.

Note that without a `--select-*` option the result for the default Build ID will be returned.
The default Build ID is the one mentioned in the first unconditional Assignment Rule.
If there is no default Build ID, the result for the unversioned queue will be returned.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.
* `--task-queue-type` (string[]) - Task queue types considered. If not specified, all types are reported. The current valid queue types are workflow, activity, or nexus.
* `--select-build-id` (string[]) - Task queue filter based on Build ID.
* `--select-unversioned` (bool) - Include the unversioned queue.
* `--select-all-active` (bool) - Include all active versions. A version is active if it had new tasks or polls recently.
* `--report-reachability` (bool) - Display task reachability information.
* `--legacy-mode` (bool) - Enable a legacy mode for servers that do not support rules-based worker versioning. This mode only provides pollers info.
* `--task-queue-type-legacy` (string-enum) - Task Queue type (legacy mode only). Options: workflow, activity. Default: workflow.
* `--partitions-legacy` (int) - Query for all partitions up to this number (experimental+temporary feature) (legacy mode only). Default: 1.

### temporal task-queue get-build-id-reachability: Retrieves information about the reachability of Build IDs on one or more Task Queues (Deprecated).

This command can tell you whether or not Build IDs may be used for new, existing, or closed workflows. Both the '--build-id' and '--task-queue' flags may be specified multiple times. If you do not provide a task queue, reachability for the provided Build IDs will be checked against all task queues.

#### Options

* `--build-id` (string[]) - Which Build ID to get reachability information for. May be specified multiple times.
* `--reachability-type` (string-enum) - Specify how you'd like to filter the reachability of Build IDs. Valid choices are `open` (reachable by one or more open workflows), `closed` (reachable by one or more closed workflows), or `existing` (reachable by either). If a Build ID is reachable by new workflows, that is always reported. Options: open, closed, existing. Default: existing.
* `--task-queue`, `-t` (string[]) - Which Task Queue(s) to constrain the reachability search to. May be specified multiple times.

### temporal task-queue get-build-ids: Fetch the sets of worker Build ID versions on the Task Queue (Deprecated).

Fetch the sets of compatible build IDs associated with a Task Queue and associated information.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.
* `--max-sets` (int) - Limits how many compatible sets will be returned. Specify 1 to only return the current default major version set. 0 returns all sets. (default: 0). Default: 0.

### temporal task-queue list-partition: Lists the Task Queue's partitions and the matching nodes they are assigned to.

The temporal task-queue list-partition command displays the partitions of a Task Queue, along with the matching node they are assigned to.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.

### temporal task-queue update-build-ids: Operations to update the sets of worker Build ID versions on the Task Queue (Deprecated).

Provides various commands for adding or changing the sets of compatible build IDs associated with a Task Queue. See the help of each sub-command for more.

### temporal task-queue update-build-ids add-new-compatible: Add a new build ID compatible with an existing ID to the Task Queue version sets.

The new build ID will become the default for the set containing the existing ID. See per-flag help for more.

#### Options

* `--build-id` (string) - The new build id to be added. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.
* `--existing-compatible-build-id` (string) - A build id which must already exist in the version sets known by the task queue. The new id will be stored in the set containing this id, marking it as compatible with the versions within. Required.
* `--set-as-default` (bool) - When set, establishes the compatible set being targeted as the overall default for the queue. If a different set was the current default, the targeted set will replace it as the new default. Defaults to false.

### temporal task-queue update-build-ids add-new-default: Add a new default (incompatible) build ID to the Task Queue version sets.

Creates a new build id set which will become the new overall default for the queue with the provided build id as its only member. This new set is incompatible with all previous sets/versions.

#### Options

* `--build-id` (string) - The new build id to be added. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.

### temporal task-queue update-build-ids promote-id-in-set: Promote an existing build ID to become the default for its containing set.

New tasks compatible with the set will be dispatched to the default id.

#### Options

* `--build-id` (string) - An existing build id which will be promoted to be the default inside its containing set. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.

### temporal task-queue update-build-ids promote-set: Promote an existing build ID set to become the default for the Task Queue.

If the set is already the default, this command has no effect.

#### Options

* `--build-id` (string) - An existing build id whose containing set will be promoted. Required.
* `--task-queue`, `-t` (string) - Name of the Task Queue. Required.

### temporal task-queue versioning: Updates or retrieves the worker Build ID assignment and redirect rules on the Task Queue.

Provides various commands for adding, listing, removing, or replacing worker Build ID assignment and redirect rules associated with a Task Queue. See the help of each sub-command for more.

#### Options

* `--task-queue`, `-t` (string) - Task queue name. Required.

### temporal task-queue versioning add-redirect-rule: Adds the rule to the list of redirect rules for this Task Queue.

Adds a new redirect rule for this Task Queue. There can be at most one redirect rule for each distinct source build ID.

#### Options

* `--source-build-id` (string) - The source build ID for this redirect rule. Required.
* `--target-build-id` (string) - The target build ID for this redirect rule. Required.
* `--yes`, `-y` (bool) - Skip confirmation.

### temporal task-queue versioning commit-build-id: Completes the rollout of a Build ID for this Task Queue.

Completes  the rollout of a BuildID and cleanup unnecessary rules possibly
created during a gradual rollout. Specifically, this command will make the
following changes atomically:
	1. Adds an unconditional assignment rule for the target Build ID at the end of the list.
	2. Removes all previously added assignment rules to the given target Build ID.
	3. Removes any unconditional assignment rules for other Build IDs.

To prevent committing invalid Build IDs, we reject the request if no pollers
have been seen recently for this Build ID. Use the `force` option to disable this validation.


#### Options

* `--build-id` (string) - The target build ID to be committed. Required.
* `--force` (bool) - Bypass the validation that pollers have been recently seen for this build ID.
* `--yes`, `-y` (bool) - Skip confirmation.

### temporal task-queue versioning delete-assignment-rule: Deletes the rule at a given index in the list of assignment rules for this Task Queue.

Deletes an assignment rule for this Task Queue. By default presence of one
unconditional rule, i.e., no hint filter or percentage, is enforced, otherwise
the delete operation will be rejected. Set `force` to true to bypass this
validation.

#### Options

* `--rule-index`, `-i` (int) - Position of the assignment rule to be replaced. Required.
* `--yes`, `-y` (bool) - Skip confirmation.
* `--force` (bool) - Bypass the validation that one unconditional rule remains.

### temporal task-queue versioning delete-redirect-rule: Deletes the rule with the given build ID for this Task Queue.

Deletes the routing rule with the given source Build ID.

#### Options

* `--source-build-id` (string) - The source build ID for this redirect rule. Required.
* `--yes`, `-y` (bool) - Skip confirmation.

### temporal task-queue versioning get-rules: Retrieves the worker Build ID assignment and redirect rules on the Task Queue.

Fetch the worker build ID assignment and redirect rules associated with a Task Queue.

### temporal task-queue versioning insert-assignment-rule: Inserts the rule to the list of assignment rules for this Task Queue.

Inserts a new assignment rule for this Task Queue. The rules are evaluated in order, starting from index 0. The first applicable rule will be applied and the rest will be ignored.

#### Options

* `--build-id` (string) - The target build ID for this assignment rule. Required.
* `--rule-index`, `-i` (int) - Insertion position in the assignment rule list. An index 0 means insert at the beginning of the list. If the given index is larger than the list size, the rule will be appended at the end of the list. Default: 0.
* `--percentage` (int) - Percentage of traffic sent to the target build ID. Default: 100.
* `--yes`, `-y` (bool) - Skip confirmation.

### temporal task-queue versioning replace-assignment-rule: Replaces the rule at a given index in the list of assignment rules for this Task Queue.

Replaces an assignment rule for this Task Queue. By default presence of one
unconditional rule, i.e., no hint filter or percentage, is enforced, otherwise
the delete operation will be rejected. Set `force` to true to bypass this
validation.

#### Options

* `--build-id` (string) - The target build ID for this assignment rule. Required.
* `--rule-index`, `-i` (int) - Position of the assignment rule to be replaced. Required.
* `--percentage` (int) - Percentage of traffic sent to the target build ID. Default: 100.
* `--yes`, `-y` (bool) - Skip confirmation.
* `--force` (bool) - Bypass the validation that one unconditional rule remains.

### temporal task-queue versioning replace-redirect-rule: Replaces the redirect rule with the given source build ID for this Task Queue.

Replaces the redirect rule with the given source build ID for this Task Queue.

#### Options

* `--source-build-id` (string) - The source build ID for this redirect rule. Required.
* `--target-build-id` (string) - The target build ID for this redirect rule. Required.
* `--yes`, `-y` (bool) - Skip confirmation.

### temporal workflow: Start, list, and operate on Workflows.

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

### temporal workflow cancel: Cancel a Workflow Execution.

The `temporal workflow cancel` command is used to cancel a [Workflow Execution](/concepts/what-is-a-workflow-execution).
Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new
Command Task will be scheduled, and the Workflow Execution will perform cleanup work.

Executions may be cancelled by [ID](/concepts/what-is-a-workflow-id):
```
temporal workflow cancel --workflow-id MyWorkflowId
```

...or in bulk via a visibility query [list filter](/concepts/what-is-a-list-filter):
```
temporal workflow cancel --query=MyQuery
```

Use the options listed below to change the behavior of this command.

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch)

### temporal workflow count: Count Workflow Executions.

The `temporal workflow count` command returns a count of [Workflow Executions](/concepts/what-is-a-workflow-execution).

Use the options listed below to change the command's behavior.

#### Options

* `--query`, `-q` (string) - Filter results using a SQL-like query.

### temporal workflow delete: Deletes a Workflow Execution.

The `temporal workflow delete` command is used to delete a specific [Workflow Execution](/concepts/what-is-a-workflow-execution).
This asynchronously deletes a workflow's [Event History](/concepts/what-is-an-event-history).
If the [Workflow Execution](/concepts/what-is-a-workflow-execution) is Running, it will be terminated before deletion.

```
temporal workflow delete \
		--workflow-id MyWorkflowId \
```

Use the options listed below to change the command's behavior.

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch)

### temporal workflow describe: Show information about a Workflow Execution.

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

Includes options set for [shared workflow start](#options-set-for-shared-workflow-start).
Includes options set for [workflow start](#options-set-for-workflow-start).
Includes options set for [payload input](#options-set-for-payload-input).

### temporal workflow fix-history-json: Updates an event history JSON file to the current format.

```
temporal workflow fix-history-json \
	--source original.json \
	--target reserialized.json
```

Use the options listed below to change the command's behavior.

#### Options

* `--source`, `-s` (string) - Path to the input file. Required.
* `--target`, `-t` (string) - Path to the output file, or standard output if not set.

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

The `temporal workflow query` command is used to [Query](/concepts/what-is-a-query) a
[Workflow Execution](/concepts/what-is-a-workflow-execution)
by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow query \
		--workflow-id MyWorkflowId \
		--name MyQuery \
		--input '{"MyInputKey": "MyInputValue"}'
```

Use the options listed below to change the command's behavior.

#### Options

* `--type` (string) - Query Type/Name. Required.
* `--reject-condition` (string-enum) - Optional flag for rejecting Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [payload input](#options-set-for-payload-input).
Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow reset: Resets a Workflow Execution by Event ID or reset type.

The temporal workflow reset command resets a [Workflow Execution](/concepts/what-is-a-workflow-execution).
A reset allows the Workflow to resume from a certain point without losing its parameters or [Event History](/concepts/what-is-an-event-history).

The Workflow Execution can be set to a given [Event Type](/concepts/what-is-an-event):
```
temporal workflow reset --workflow-id=meaningful-business-id --type=LastContinuedAsNew
```

...or a specific any Event after `WorkflowTaskStarted`.
```
temporal workflow reset --workflow-id=meaningful-business-id --event-id=MyLastEvent
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
* `--yes`, `-y` (bool) - Confirm prompt to perform batch. Only allowed if query is present.



### temporal workflow show: Show Event History for a Workflow Execution.

The `temporal workflow show` command provides the [Event History](/concepts/what-is-an-event-history) for a
[Workflow Execution](/concepts/what-is-a-workflow-execution). With JSON output specified, this output can be given to
an SDK to perform a replay.

Use the options listed below to change the command's behavior.

#### Options

* `--follow`, `-f` (bool) - Follow the progress of a Workflow Execution in real time (does not apply
  to JSON output).

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow signal: Signal Workflow Execution by Id.

The `temporal workflow signal` command is used to [Signal](/concepts/what-is-a-signal) a
[Workflow Execution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow signal \
		--workflow-id MyWorkflowId \
		--name MySignal \
		--input '{"MyInputKey": "MyInputValue"}'
```

Use the options listed below to change the command's behavior.

#### Options

* `--name` (string) - Signal Name. Required.

Includes options set for [payload input](#options-set-for-payload-input).

#### Options set for single workflow or batch:

* `--workflow-id`, `-w` (string) - Workflow Id. Either this or query must be set.
* `--run-id`, `-r` (string) - Run Id. Cannot be set when query is set.
* `--query`, `-q` (string) - Start a batch to operate on Workflow Executions with given List Filter. Either this or
  Workflow Id must be set.
* `--reason` (string) - Reason to perform batch. Only allowed if query is present unless the command specifies
  otherwise. Defaults to message with the current user's name.
* `--yes`, `-y` (bool) - Confirm prompt to perform batch. Only allowed if query is present.

### temporal workflow stack: Query a Workflow Execution for its stack trace.

The `temporal workflow stack` command [Queries](/concepts/what-is-a-query) a
[Workflow Execution](/concepts/what-is-a-workflow-execution) with `__stack_trace` as the query type.
This returns a stack trace of all the threads or routines currently used by the workflow, and is
useful for troubleshooting.

```
temporal workflow stack --workflow-id MyWorkflowId
```

Use the options listed below to change the command's behavior.

#### Options

* `--reject-condition` (string-enum) - Optional flag for rejecting Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

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

#### Options set for shared workflow start:

* `--workflow-id`, `-w` (string) - Workflow Id.
* `--type` (string) - Workflow Type name. Required.
* `--task-queue`, `-t` (string) - Workflow Task queue. Required.
* `--run-timeout` (duration) - Timeout of a Workflow Run.
* `--execution-timeout` (duration) - Timeout for a WorkflowExecution, including retries and ContinueAsNew tasks.
* `--task-timeout` (duration) - Start-to-close timeout for a Workflow Task. Default: 10s.
* `--search-attribute` (string[]) - Passes Search Attribute in key=value format. Use valid JSON formats for value.
* `--memo` (string[]) - Passes Memo in key=value format. Use valid JSON formats for value.

#### Options set for workflow start:

* `--cron` (string) - Cron schedule for the workflow. Deprecated - use schedules instead.
* `--fail-existing` (bool) - Fail if the workflow already exists.
* `--start-delay` (duration) - Specify a delay before the workflow starts. Cannot be used with a cron schedule. If the
  workflow receives a signal or update before the delay has elapsed, it will begin immediately.
* `--id-reuse-policy` (string) - Allows the same Workflow Id to be used in a new Workflow Execution. Options:
  AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.

#### Options set for payload input:

* `--input`, `-i` (string[]) - Input value (default JSON unless --input-payload-meta is non-JSON encoding). Can
  be given multiple times for multiple arguments. Cannot be combined with --input-file.
* `--input-file` (string[]) - Reads a file as the input (JSON by default unless --input-payload-meta is non-JSON
  encoding). Can be given multiple times for multiple arguments. Cannot be combined with --input.
* `--input-meta` (string[]) - Metadata for the input payload. Expected as key=value. If key is encoding, overrides the
  default of json/plain.
* `--input-base64` (bool) - If set, assumes --input or --input-file are base64 encoded and attempts to decode.

### temporal workflow terminate: Terminate Workflow Execution by ID or List Filter.

The `temporal workflow terminate` command is used to terminate a
[Workflow Execution](/concepts/what-is-a-workflow-execution). Canceling a running Workflow Execution records a
`WorkflowExecutionTerminated` event as the closing Event in the workflow's Event History. Workflow code is oblivious to
termination. Use `temporal workflow cancel` if you need to perform cleanup in your workflow.

Executions may be terminated by [ID](/concepts/what-is-a-workflow-id) with an optional reason:
```
temporal workflow terminate [--reason my-reason] --workflow-id MyWorkflowId
```

...or in bulk via a visibility query [list filter](/concepts/what-is-a-list-filter):
```
temporal workflow terminate --query=MyQuery
```

Use the options listed below to change the behavior of this command.

#### Options

* `--workflow-id`, `-w` (string) - Workflow Id. Either this or query must be set.
* `--run-id`, `-r` (string) - Run Id. Cannot be set when query is set.
* `--query`, `-q` (string) - Start a batch to terminate Workflow Executions with given List Filter. Either this or
  Workflow Id must be set.
* `--reason` (string) - Reason for termination. Defaults to message with the current user's name.
* `--yes`, `-y` (bool) - Confirm prompt to perform batch. Only allowed if query is present.

### temporal workflow trace: Terminate Workflow Execution by ID or List Filter.

The `temporal workflow trace` command display the progress of a [Workflow Execution](/concepts/what-is-a-workflow-execution) and its child workflows with a trace.
This view provides a great way to understand the flow of a workflow.

Use the options listed below to change the behavior of this command.

#### Options

* `--fold` (string[]) - Statuses for which Child Workflows will be folded in (this will reduce the number of information fetched and displayed). Case-insensitive and ignored if no-fold supplied. Available values: running, completed, failed, canceled, terminated, timedout, continueasnew.
* `--no-fold` (bool) - Disable folding. All Child Workflows within the set depth will be fetched and displayed. 
* `--depth` (int) - Depth of child workflows to fetch. Use -1 to fetch child workflows at any depth. Default: -1.
* `--concurrency` (int) - Number of concurrent workflow histories that will be requested at any given time. Default: 10.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow update: Updates a running workflow synchronously.

The `temporal workflow update` command is used to synchronously [Update](/concepts/what-is-an-update) a
[WorkflowExecution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id).

```
temporal workflow update \
		--workflow-id MyWorkflowId \
		--name MyUpdate \
		--input '{"Input": "As-JSON"}'
```

Use the options listed below to change the command's behavior.

#### Options

* `--name` (string) - Update Name. Required.
* `--workflow-id`, `-w` (string) - Workflow Id. Required.
* `--update-id` (string) - Update ID. If unset, default to a UUID.
* `--run-id`, `-r` (string) - Run Id. If unset, the currently running Workflow Execution receives the Update.
* `--first-execution-run-id` (string) - Send the Update to the last Workflow Execution in the chain that started
  with this Run Id.

Includes options set for [payload input](#options-set-for-payload-input).
