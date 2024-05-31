# Temporal CLI Commands

Commands for the Temporal CLI

<!-- 

NOTES FOR ERICA

* All URLs are borked due to Info Arch re-org
* Word wrapping to 80 chars
    * What about options? they tend to run long
    * Could the text go to a second line after the spaced-dash?
* --help, --long-help, --full-help to provide progressive support?
* --long, --extra-long/--full and not --really-long
  * `--long`, `-l` (bool) -
  * `--really-long` (bool) - 
* Auditing confirmation elements, especially bools
  * `--yes`, `-y` (bool) - has multiple meanings based on use
    (default true, false varies)
* Payloads
  * `--no-json-shorthand-payloads` (bool) - Shorter? --literal-payload? --string-payload?
  * `--input-base64` (bool) - --data-base64?

* These all seem to do with presentation
  * `--detail` (bool) -
  * `--event-details` (bool) -
  * `--log-config` (bool) -
  * `--no-fold` (bool) -
  * `--raw` (bool) -
  * `--follow`, `-f` (bool) -
  * `--reset-points` (bool) - ? --show-reset-points?

* Consistency?
  * `--pause` (bool) -
  * `--unpause` (bool) -
  * `--pause-on-failure` (bool) -
  * `--paused` (bool) - Perhaps: set-paused? set-paused-state?

* `--global` (bool) -
* `--promote-global` (bool) -

* `--enable-connection` (bool) - ? --connected?
* `--fail-existing` (bool) - ? --fail-current?

* `--archived` (bool) - ? --set-archived?

* `--headless` (bool) -

* `--set-as-default` (bool) -
 
* `--tls-disable-host-verification` (bool) -
* `--tls` (bool) -

cat `make path` | egrep '\]\(' | grep -v options-set | grep "/concepts" | open -f | sort

[Temporal Server](/clusters)
[Task Queue](/workers#task-queue)
[poller](/dev-guide/worker-performance#poller-count)
Temporal [Workers](/workers)
[Workflow](/workflows)
[Workflow Executions](/workflows#workflow-execution)
[Workflow Execution](/workflows#workflow-execution)
[Workflow ID](/workflows#workflow-id)
[list filter](/visibility#list-filter)
[Event History](/workflows#event-history)
[Query](workflows#query)
[Queries](/workflows#query)
[Event Type](/workflows#event)
[Signal](/workflows#signal)
[Update](/workflows#update)

* -h, -o, -v are wonky:
Flags:
      --color string                                      Output coloring. Accepted values: always, never, auto. (default "auto")
      --env string                                        Active environment name. (default "default")
      --env-file $HOME/.config/temporalio/temporal.yaml   Path to environment settings file (defaults to $HOME/.config/temporalio/temporal.yaml)
  -h, --help                                              help for temporal // WONKY
      --log-format string                                 Log format. Options are "text" and "json". Default is "text"
      --log-level server start-dev                        Log level. Default is "info" for most commands and "warn" for server start-dev. Accepted values: debug, info, warn, error, never. (default "info")
      --no-json-shorthand-payloads                        Raw payload output, even if they are JSON
  -o, --output string                                     Non-logging data output format. Accepted values: text, json, jsonl, none. (default "text") // WONKY
      --time-format string                                Time format. Accepted values: relative, iso, raw. (default "relative")
  -v, --version                                           version for temporal // WONKY

-----

This document is automatically parsed.
Follow these rules.

IN-HOUSE STYLE

* Ordering:
  * Use alphabetical order for commands.
  * Put the most commonly used options first in your list.
* Periods:
  * Don't end short command descriptions with periods.
  * Use periods consistently in long command descriptions at the end of sentences.
  * Use periods consistently in Option elements.
    Period use is mandatory to ensure proper parsing.
    Include periods at the end of Option descriptions and extra attributes.
  * Periods are required at the end of "Include Option set" lines.
  * Avoid deprecated period-delineated versions of environment-specific keys.
    * Yes:
          temporal env set \
              --env prod \
              --key tls-cert-path \
              --value /home/my-user/certs/cluster.cert`
    * No: `temporal env set prod.tls-cert-path /home/my-user/certs/cluster.cert`.
* Wrapping:
  * User-visible line length: 80 characters max.
  * Wrap your Options declarations to match existing examples.
    This improves the maintainability of this file.
  * Hand-wrap your long descriptions as they appear for users.
  * When wrapping code, use a single space and a backslash.
* Examples:
  * Demonstrate at least one example of the command in every long description.
  * For commands with a single command-level option, include it the mandatory
    example.
    Use square brackets to highlight its optional nature if the long
    description would suffer from two very similar command invocations.
  * Use YourEnvironment, YourNamespace, etc as metasyntactic stand-ins.
  * Commands with subcommands cannot be run on their own.
    Because of this, always use full command examples.
  * Prefer double quotes to single quotes for sample input
  * Use long options and flags for examples (`--namespace`), not short options
    and flags (`-n`).
* Re-use existing wording and phrases wherever possible.
* When presenting options use a space rather than equal to set them.
  This is more universally supported and consistent with POSIX guidelines.
  * Yes: `temporal command --namespace YourNamespace`.
  * No: `temporal command --namespace=YourNamespace`.
* Grammar and casing:
  * Word command short descriptions as if they began "This command will..."
    Use sentence casing for the short description.
  * Use imperative-form verbs for short descriptions:
    * Yes: `Pause or unpause a Schedule`.
    * No: `This command pauses or unpauses a Schedule`.
  * Use concise noun definitions for options.
  * ID is fully capitalized.

For options and flags:

* When options and flags can be passed multiple times, say so explicitly
  in the usage text: "Can be passed multiple times."
* Never rely on the flag type (e.g. `string`, `bool`, etc.) being
  shown to the user.
  It is replaced/hidden when a `META-VARIABLE` is used.
* Where possible, use a `META-VARIABLE` (all caps and wrapped in `\``s)
  to describe/reference content passed to an option.
* Limit `code spans` to meta-variables.
  To reference other options or specify literal values, use double quotes.
* Avoid parentheticals unless absolutely necessary.

Examples:

These show correct/incorrect usage text for the optional `--history-uri` flag:

* Preferred:
    "Archive history at \`URI\`. <explanation>".
* Avoid incomplete sentences:
    "_History archival \`URI\`_. <explanation>".
* Avoid wrong verb tenses:
    "_Archives history at \`URI\`_. <explanation>".
* Avoid missing metavariables:
    "Archive history at _the specified URI_. <explanation>".
* Avoid unnecessary parenthetical:
    "`Archive history at \`URI\` _(note: <explanation>)_.`".


COMMAND ENTRY OVERVIEW

A Command entry uses the following format:

    ### <command>: <short-description>

    <long description>

    (optional command implementation configuration)

    #### Options 
    (or)
    #### Options set for options set name

    * `--<long-option>`( , `-<short-option>`) <data-type> -
      <short-description>.
      ( <extra-attributes>. )
    * `--<long-option>`( , `-<short-option>`) <data-type> -
      <short-description>.
      ( <extra-attributes>. )
    * `--<long-option>`( , `-<short-option>`) <data-type> -
      <short-description>.
      ( <extra-attributes>. )
    * ...

    optional: Includes options set for [<options-set-name>](#options-set-for-<options-set-link-name>).

Note:
* option-set-name is the text after "for " in "#### Options set for ".
* option-set-link-name is the same text with spaces replaced with hyphens.

COMMAND LISTING

* Use H3 `### <command>: <short-description>` headings for each command.
  * One line only. Use the complete command path with parent commands.
  * Use square-bracked delimited arguments to document positional arguments.
    For example `temporal operator namespace delete [namespace]`.
  * Everything up to ':' or '[' is the command.
    Square-bracketed positional arguments are not part of the command.
  * Everything following the ':' or '[' is used for the short-description,
    a concise command explanation.
* A command's long description continues until encountering the H4
  (`#### Options`) header.
* At the end of the long description, an optional XML comment configures
  the command implementation.
  Use one asterisk-delimited bullet per line.
  * `* has-init` - invokes `initCommand` method.
  * `* exact-args=<number>` - Require this exact number of args.
  * `* maximum-args=<number>` - Require this maximum number of args.

LONG DESCRIPTION SECTION

* Your text format is preserved literally, so be careful with line lengths.
  Use manual wrapping as needed.

OPTIONS SECTION

* Start the optional Options section with an H4 `#### Options` header line.
* Follow the header declaration with a list of options.
  The individual option definition syntax follows below the header declaration.
* You must include at least one option.
  Otherwise, `gen-commands` will complete but the CLI utility will not run.
* To incorporate an existing options set, add a single line below options
  like this, remembering to end every `Includes options set for` line with a
  period:

  ```
  Includes options set for [<options-set-name>](#options-set-for-<options-set-link-name>).
  ```

  For example:

  ```
  Includes options set for [client](#options-set-for-client).
  ```

  An options set declaration is the equivalent of pasting those options into
  the bulleted options list.

  * Options that are similar but slightly different don't need to be in
    option sets.
  * Every "Option Set for" declaration links to the H4 entry that supplies
    the inherited options.
  * Reserve option sets for when the behavior of the option is the same
    across commands.
    Otherwise, just use copy/paste.

DEFINING AN OPTION

* List each option separately.
* Start each option definition with an asterisk-delimiting bullet.
* Order the most commonly used options first.
* Use this format:

  ```
  * `--<long-option>`( , `-<short option>`) <data-type> -
    <short-description>.
    ( <extra-attributes>. )
  ```

  This contrived example uses all these features.
  In reality, the short option `-a` does not actually exist alongside
  `--address`:

  ```
  * `--address`, `-a` (string) -
    Connect to the Temporal Service at `HOST:PORT`.
    Default: 127.0.0.1:7233. Env: TEMPORAL_ADDRESS.
  ```

* Each option listing includes a long option with a double dash and a
  meaningful name.
* [Optional] A short option uses a single dash and a short string.
  When used, separate the long and short option with a comma and a space.
* Backtick every option and short description.
  Include the dash or dashes within the ticks.
  For example: `` `--workflow-id`, `-w` ``.
* A data type follows option names indicating the required value type for the
  option.
  The type is `bool`, `duration`, `int`, `string`, `string[]`, `string-enum`,
  or `timestamp`. (_TODO: more_.).
  Always parenthesize data types.
  For example: `` `--raw` (bool) ``.
* A dash follows the data type, with a space on either side.
* The short description is free-form text and follows the dash.
  Take care not to match trailing attributes.
  Newline wrapping and/or two-space indentation condenses to a single space.
* [Optional] extra attributes include:
  * `Required.` - Marks the option as required.
  * `Default: <default-value>.` - Sets the default value of the option.
     No default means zero value of the type.
     Do not include defaults for Boolean values.
  * `Options: <option>, <option>.` - Sets the possible options for a string
     enum type.
  * `Env: <env-var>.` - Binds the environment variable to this flag.
    For example: `Env: TEMPORAL_ADDRESS`.
    Use the variable name without dollar signs, etc.

-->

### temporal: Temporal command-line interface and development server

The Temporal CLI manages, monitors, and debugs Temporal apps. It lets you
run a local Temporal Service, start Workflow Executions, pass messages to
running Workflows, inspects state, and more.

* Start a local development service:
      `temporal server start-dev`
* View help: pass --help to any command:
      `temporal activity complete --help`

<!--
* has-init
-->

#### Options

* `--env` (string) -
  Active environment name (`ENV`).
  Default: default.
  Env: TEMPORAL_ENV.
* `--env-file` (string) -
  Path to environment settings file.
  (defaults to `$HOME/.config/temporalio/temporal.yaml`).
* `--log-level` (string-enum) -
  Log level.
  Default is "info" for most commands and "warn" for `server start-dev`.
  Options: debug, info, warn, error, never. Default: info.
* `--log-format` (string) -
  Log format.
  Options: text, json.
  Default: text.
* `--output`, `-o` (string-enum) -
  Non-logging data output format.
  Options: text, json, jsonl, none.
  Default: text.
* `--time-format` (string-enum) -
  Time format.
  Options: relative, iso, raw.
  Default: relative.
* `--color` (string-enum) -
  Output coloring.
  Options: always, never, auto.
  Default: auto.
* `--no-json-shorthand-payloads` (bool) -
  Raw payload output, even if they are JSON.

### temporal activity: Complete or fail an Activity

Update Activity state to report completion or failure. This command marks
an Activity as successfully finished or as having encountered an error.

#### Options

Includes options set for [client](#options-set-for-client).

### temporal activity complete: Complete an Activity

Complete an Activity, marking it as successfully finished. Specify the ID
and include a JSON result for the returned value:

```
temporal activity complete \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId \
    --result '{"YourResultKey": "YourResultVal"}'
```

#### Options

* `--activity-id` (string) -
  Activity `ID` to complete.
  Required.
* `--result` (string) -
  Result `JSON` for completing the Activity.
  Required.
* `--identity` (string) -
  Identity of the user submitting this request.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal activity fail: Fail an Activity

Fail an Activity, marking it as having encountered an error. Specify the
Activity and Workflow IDs:

```
temporal activity fail \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId
```

#### Options

* `--activity-id` (string) -
  `ID` of the Activity to be failed.
  Required.
* `--detail` (string) -
  JSON data with the reason for failing the Activity.
* `--identity` (string) -
  Identity of the user submitting this request.
* `--reason` (string) -
  Reason for failing the Activity.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal batch: Manage batch jobs


A batch job executes a command affecting multiple Workflow Executions at once.
Specify the Workflow Executions to include and the type of batch job to apply.
For example, to cancel all running 'YourWorkflow' Workflows:

```
temporal batch workflow cancel \
  --query 'ExecutionStatus = "Running" AND WorkflowType="YourWorkflow"' \
  --reason "Testing"
```

The `batch` command keyword is optional when you supply a `--query`. This
invocation is identical to the previous one:

```
temporal workflow cancel \
  --query 'ExecutionStatus = "Running" AND WorkflowType="YourWorkflow"' \
  --reason "Testing"
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal batch describe: Show batch job progress

Show the progress of an ongoing batch job. Pass a valid job ID to display
its information:

```
temporal batch describe --job-id YourJobId
```

#### Options

* `--job-id` (string) -
  Batch job ID.
  Required.

### temporal batch list: List all batch jobs

Return a list of batch jobs on the Service or within a single Namespace.
For example, list the batch jobs for "YourNamespace":
```
temporal batch list --namespace YourNamespace
```

#### Options

* `--limit` (int) -
  Max output count.

### temporal batch terminate: Terminate a batch job

Terminate a batch job with the provided job ID. Provide a reason for the
termination. This explanation is stored with the Service for later reference.
For example:

```
temporal batch terminate \
    --job-id YourJobId \
    --reason YourTerminationReason
```

#### Options

* `--job-id` (string) -
  Job Id to terminate.
  Required.
* `--reason` (string) -
  Reason for terminating the batch job.
  Required.

### temporal env: Manage environments

Environments manage key-value presets, auto-configuring Temporal CLI options
for you. Set up distinct environments like "dev" and "prod" for convenience.
For example:

```
temporal env set \
    --env prod \
    --key address \
    --value production.f45a2.tmprl.cloud:7233
```

Each environment is isolated. Changes to "prod" presets won't affect "dev".

For easiest use, set a `TEMPORAL_ENV` environmental variable in your shell.
The Temporal CLI checks for an `--env` option first, then checks the shell.
If neither is set, most commands use `default` environmental presets if
they are available.

### temporal env delete: Delete an environment or environment property


Remove an environment entirely or remove a key-value pair within an
environment. If you don't specify an environment (with --env or by setting
the TEMPORAL_ENV variable), this updates the default environment. 
For example:

```
temporal env delete --env YourEnvironment
```

or

```
temporal env delete --env prod  --key tls-key-path
```

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) -
  Property name.

### temporal env get: Show environment properties

List the properties for a given environment:

```
temporal env get --env YourEnvironment
```

Print a single property:

```
temporal env get --env YourEnvironment --key YourPropertyKey
```

If you don't specify an environment name, with `--env` or by setting
`TEMPORAL_ENV` as an environmental variable in your shell, this command
lists properties in the `default` environment.

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) -
  Property name.

### temporal env list: List environments

List the environments you have set up on your local computer. The output 
enumerates the environments stored in the Temporal environment file on
your computer ("$HOME/.config/temporalio/temporal.yaml").

### temporal env set: Set environment properties

Assign a value to a property key and store it to an environment:

```
temporal env set \
    --env environment \
    --key property \
    --value value
```

If you don't specify an environment name, with `--env` or by setting
`TEMPORAL_ENV` as an environmental variable in your shell, this command
sets properties in the `default` environment.

Setting property names lets the CLI automatically set options on your behalf.
Storing these property values in advance reduces the effort required to 
issue CLI commands and helps avoid typos.

<!--
* maximum-args=2
-->

#### Options

* `--key`, `-k` (string) -
  Property name.
  Required.
* `--value`, `-v` (string) -
  Property value.
  Required.

### temporal operator: Manage Temporal deployments

Operator commands manage and fetch information about Namespaces, 
Search Attributes, and Temporal Clusters:

```
temporal operator [command] [subcommand] [options]
```

For example, to view a cluster-by-cluster overview:

```
temporal operator cluster describe
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal operator cluster: Manage a Temporal Cluster

Perform operator actions on Temporal Clusters.

```
temporal operator cluster [subcommand] [options]
```

For example to check Cluster health:

```
temporal operator cluster health
```

### temporal operator cluster describe: Describe a Temporal Cluster

View information about a Temporal Cluster, including Cluster Name,
persistence store, and visibility store. Add `--detail` for additional
information:

```
temporal operator cluster describe [--detail]
```

#### Options

* `--detail` (bool) -
  High-detail output.

### temporal operator cluster health: Check Frontend Service health

View information about the health of a Frontend Service:

```
temporal operator cluster health
```

### temporal operator cluster list: Show Temporal Clusters

Print a list of Temporal Clusters registered to this system. Report details
include the Cluster's name, ID, address, History Shard count,
Failover version, and availability.

```
temporal operator cluster list [--limit max-count]
```

#### Options

* `--limit` (int) -
  Max count to list.

### temporal operator cluster remove: Remove a Temporal Cluster

Remove a registered Temporal Cluster from this system.

```
temporal operator cluster remove --name YourClusterName
```

#### Options

* `--name` (string) -
  Cluster name.
  Required.

### temporal operator cluster system: Show Temporal Cluster info


Show Temporal Server information for Temporal Clusters: Server version, 
scheduling support, and more. This information helps diagnose problems with
the Temporal Server. This command defaults to the local Service. Otherwise, 
use the `--frontend-address` option to specify a Cluster endpoint.

```
temporal operator cluster system \ 
    --frontend-address "your_remote_endpoint:your_remote_port"
``` 

### temporal operator cluster upsert: Add/update a Temporal Cluster

Add, remove, or update a registered ("remote") Temporal Cluster.

```
temporal operator cluster upsert [options]
```

For example:

```
temporal operator cluster upsert \
    --frontend_address "your_remote_endpoint:your_remote_port"
    --enable_connection false
```

#### Options

* `--frontend-address` (string) -
  Remote endpoint.
  Required.
* `--enable-connection` (bool) -
  Enabled/disabled connection.

### temporal operator namespace: Namespace operations

Perform operations on Temporal Cluster Namespaces:

```
temporal operator namespace [command] [command options]
```

For example:

```
temporal operator namespace create \
    --namespace YourNewNamespaceName
```

### temporal operator namespace create: Register a new Namespace

Create a new Namespace on the Temporal Server:

```
temporal operator namespace create \
    --namespace YourNewNamespaceName \
    [options]
````

To create a global Namespace, for cross-region data replication:

```
temporal operator namespace create \
    --global \
    --namespace YourNewNamespaceName
```

Configure other settings like retention and Visibility Archival State
as needed. For example, the Visibility Archive can be set on a separate URI:

```
temporal operator namespace create \
    --retention 5d \
    --visibility-archival-state enabled \
    --visibility-uri YourURI \
    --namespace YourNewNamespaceName
```

Note: URI values for archival states can't be changed after being enabled.

<!--
* maximum-args=1
-->

#### Options

* `--active-cluster` (string) -
  Active cluster name.
* `--cluster` (string[]) -
  Cluster names for Namespace creation.
  Can be passed multiple times.
* `--data` (string) -
  Namespace data.
* `--description` (string) -
  Namespace description.
  Comma-separated "KEY=VALUE" string pairs.
* `--email` (string) -
  Owner email.
* `--global` (bool) -
  Enable/disable cross-region data replication.
* `--history-archival-state` (string-enum) -
  History archival state.
  Options: disabled, enabled.
  Default: disabled.
* `--history-uri` (string) -
  Archive history to this `URI`.
  Once enabled, cannot be changed.
* `--retention` (duration) -
  Time to preserve closed Workflows before deletion.
  Default: 72h.
* `--visibility-archival-state` (string-enum) -
  Visibility archival state.
  Options: disabled, enabled.
  Default: disabled
* `--visibility-uri` (string) -
  Archive visibility data to this `URI`.
  Once enabled, cannot be changed.

### temporal operator namespace delete [namespace]: Delete Namespace

Removes a Namespace from the Service.

```
temporal operator namespace delete [options]
```

For example:

```
temporal operator namespace delete \
    --namespace YourNamespaceName
```

<!--
* maximum-args=1
-->

#### Options

* `--yes`, `-y` (bool) -
  Request confirmation before deletion.

### temporal operator namespace describe [namespace]: Describe Namespace

Provide long-form information about a Namespace identified by its ID or name:

```
temporal operator namespace describe --namespace-id YourNamespaceId
```

or

```
temporal operator namespace describe --namespace YourNamespaceName
```

<!--
* maximum-args=1
-->

#### Options

* `--namespace-id` (string) -
  Namespace ID.

### temporal operator namespace list: List Namespaces

A long-form listing for all Namespaces on the Service:

```
temporal operator namespace list
``` 

### temporal operator namespace update: Update Namespace

Update a Namespace using the properties you specify.

```
temporal operator namespace update [options]
``` 

Assign a Namespace's active Cluster:

```
temporal operator namespace update \
    --namespace YourNamespaceName \
    --active-cluster NewActiveCluster
```

Promote a Namespace for data replication:

```
temporal operator namespace update \
    --namespace YourNamespaceName \
    --promote-global
```

You may update archives that were previously enabled or disabled. 
Note: URI values for archival states can't be changed after being enabled:

```
temporal operator namespace update \
    --namespace YourNamespaceName \
    --history-archival-state enabled \
    --visibility-archival-state disabled
```

<!--
* maximum-args=1
-->

#### Options
* `--active-cluster` (string) -
  Active cluster name.
* `--cluster` (string[]) -
  Cluster names.
* `--data` (string[]) -
  Set a "KEY=VALUE" string pair in Namespace data.
  KEY is an unquoted string identifier.
  Unquoted JSON is recommended for the VALUE.
  Can be passed multiple times.
* `--description` (string) -
  Namespace description.
* `--email` (string) -
  Owner email.
* `--promote-global` (bool) -
  Enable/disable cross-region data replication.
* `--history-archival-state` (string-enum) -
  History archival state.
  Options: disabled, enabled.
* `--history-uri` (string) -
  Archive history to this `URI`.
  Once enabled, cannot be changed.
* `--retention` (duration) -
  Length of time a closed Workflow is preserved before deletion
* `--visibility-archival-state` (string-enum) -
  Visibility archival state. Options: disabled, enabled
* `--visibility-uri` (string) -
  Archive visibility data to this `URI`.
  Once enabled, cannot be changed.

### temporal operator search-attribute: Search Attribute operations

Create, list, or remove Search Attributes, fields stored in a Workflow
Execution's metadata:

```
temporal operator search-attribute create \
    --name YourAttributeName \
    --type Keyword
```

Supported types include: Text, Keyword, Int, Double, Bool, Datetime, and
KeywordList.

### temporal operator search-attribute create: Add custom Search Attributes

Add one or more custom Search Attributes:

```
temporal operator search-attribute create \
    --name YourAttributeName \
    --type Keyword
```

#### Options

* `--name` (string[]) -
  Search Attribute name. Required
* `--type` (string[]) -
  Search Attribute type.
  Options: Text, Keyword, Int, Double, Bool, Datetime, KeywordList.
  Required.

### temporal operator search-attribute list: List Search Attributes

Display a list of active Search Attributes that can be assigned or used with 
Workflow Queries.

```
temporal operator search-attribute list
```

### temporal operator search-attribute remove: Remove custom Search Attributes

Remove custom Search Attributes from the options that can be assigned or
used with Workflow Queries.

```
temporal operator search-attribute remove \
    --name YourAttributeName
```

Remove without confirmation:

```
temporal operator search-attribute remove \
    --name YourAttributeName \
    --yes false
```

#### Options

* `--name` (string[]) -
  Search Attribute name.
  Required.
* `--yes`, `-y` (bool) -
  Override confirmation request before removal.

### temporal schedule: Perform operations on Schedules

Create, use, and update Schedules that allow Workflow Executions to
be created at specified times:

```
temporal schedule [commands] [options]
```

For example:

```
temporal schedule describe \
    --schedule-id "YourScheduleId"
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal schedule backfill: Backfill past actions


Batch-execute actions that would have run during a specified time interval.
Use this command to fill in Workflow Runs from when a Schedule was paused,
from before a Schedule was created, from the future, or to re-process an
interval that was previously executed.

Backfills require a Schedule ID and the time period covered by the request. 
It's best to use the `BufferAll` or `AllowAll` policies.

For example:

```
  temporal schedule backfill \
    --schedule-id "YourScheduleId" \
    --start-time "2022-05-01T00:00:00Z" \
    --end-time "2022-05-31T23:59:59Z" \
    --overlap-policy BufferAll
```

#### Options set for overlap policy:

* `--overlap-policy` (string-enum) -
  Overlap policy.
  Options: Skip, BufferOne, BufferAll, CancelOther, TerminateOther, AllowAll.
  Default: Skip.

#### Options set for schedule id:

* `--schedule-id`, `-s` (string) -
  Schedule ID.
  Required.

#### Options

* `--end-time` (timestamp) -
  Backfill end time.
  Required.
* `--start-time` (timestamp) -
  Backfill start time.
  Required.

### temporal schedule create: Create a new Schedule

Create a new Schedule on the Frontend Service. A Schedule will
automatically start new Workflow Executions at the times you specify.

For example:

```
  temporal schedule create \
    --schedule-id "YourScheduleId" \
    --calendar '{"dayOfWeek":"Fri","hour":"3","minute":"30"}' \
    --workflow-id "YourBaseWorkflowId" \
    --task-queue "YourTaskQueue" \
    --workflow-type "YourWorkflowType"
```

Schedules support any combination of `--calendar`, `--interval`, 
and `--cron`.

#### Options set for schedule configuration:

* `--calendar` (string[]) -
  Calendar specification in JSON.
  For example: `{"dayOfWeek":"Fri","hour":"17","minute":"5"}`.
* `--catchup-window` (duration) -
  Maximum catch-up time for when the Service is unavailable.
* `--cron` (string[]) -
  Calendar specification in cron string format.
  For example: `"3 11 * * Fri"`.
* `--end-time` (timestamp) -
  Schedule end time
* `--interval` (string[]) -
  Interval duration.
  For example, 90m, or 90m/13m to include phase offset.
* `--jitter` (duration) -
  Per-action jitter range.
* `--notes` (string) -
  Initial notes field value.
* `--paused` (bool) -
  Initial value for paused state.
* `--pause-on-failure` (bool) -
  Pause schedule after Workflow failures.
* `--remaining-actions` (int) -
  Total allowed actions. 
  Default is zero (unlimited).
* `--start-time` (timestamp) -
  Schedule start time.
* `--time-zone` (string) -
  Interpret calendar specs with the `TZ` time zone.
  For a list of time zones, see: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones.
* `--schedule-search-attribute` (string[]) -
  Set schedule Search Attributes using "KEY=VALUE" format.
  KEY is an unquoted string identifier.
  VALUE is a JSON string.
  Can be passed multiple times.
* `--schedule-memo` (string[]) -
  Set a schedule memo using "KEY=VALUE" format.
  KEY is an unquoted string identifier.
  VALUE is a JSON string.
  Can be passed multiple times.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

### temporal schedule delete: Delete a Schedule

Deletes a Schedule on the front end Service:

```
temporal schedule delete --schedule-id YourScheduleId
``` 

Removing a schedule won't affect the Workflow Executions it started that
are still running. To cancel or terminate these Workflow Executions, use 
`temporal workflow delete` with the `TemporalScheduledById` Search Attribute.
 
#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule describe: Get Schedule configuration and current state

Show a Schedule configuration including information about past, current, 
and future Workflow runs:

```
temporal schedule describe --schedule-id YourScheduleId
```

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule list: List Schedules

Lists the Schedules hosted by a Namespace:

```
temporal schedule list --namespace YourNamespace
```

#### Options

* `--long`, `-l` (bool) -
  Show detailed information
* `--really-long` (bool) -
  Show extensive information in non-table form.

### temporal schedule toggle: Pause or unpause a Schedule

Pause or unpause a Schedule:

```
temporal schedule toggle \
    --schedule-id "YourScheduleId" \
    --pause \
    --reason "YourReason"
```

and

```
temporal schedule toggle
    --schedule-id "YourScheduleId" \
    --unpause \
    --reason "YourReason"
```

The reason updates the Schedule's `notes` field to support operations
communication. It defaults to "(no reason provided)" when omitted.

#### Options

* `--pause` (bool) -
  Pause the schedule.
* `--reason` (string) -
  Reason for pausing/unpausing.
  Default: "(no reason provided)"
* `--unpause` (bool) -
  Unpause the schedule.

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule trigger: Immediately run a Schedule

Trigger a Schedule to run immediately:

```
temporal schedule trigger --schedule-id "YourScheduleId"
```

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).

### temporal schedule update: Update Schedule details

Update an existing Schedule with new configuration details, including
spec, action, and policies:

```
temporal schedule update 
temporal schedule update \  
 --schedule-id "YourScheduleId"   
 --workflow-type "NewWorkflowType"
```

#### Options

Includes options set for [schedule-configuration](#options-set-for-schedule-configuration).
Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

### temporal server: Run Temporal Server

Run a development [Temporal Server](/clusters) on your local system.
View the Web UI for the default configuration at http://localhost:8233:

```
temporal server start-dev
```

Add persistence for Workflow Executions across runs:

```
temporal server start-dev \
    --db-filename path-to-your-local-persistent-store
```

Set the port from the front-end gRPC Service (7233 default):

```
temporal server start-dev \
    --port 7000
```

Use a custom port for the Web UI. The default is the gRPC port (7233 default)
plus 1000 (8233):

```
temporal server start-dev \
    --ui-port 3000
```

### temporal server start-dev: Start Temporal development server

Run a development [Temporal Server](/clusters) on your local system.
View the Web UI for the default configuration at http://localhost:8233:

```
temporal server start-dev
```

Add persistence for Workflow Executions across runs:

```
temporal server start-dev \
    --db-filename path-to-your-local-persistent-store
```

Set the port from the front-end gRPC Service (7233 default):

```
temporal server start-dev \
    --port 7000
```

Use a custom port for the Web UI. The default is the gRPC port (7233 default)
plus 1000 (8233):

```
temporal server start-dev \
    --ui-port 3000
```

#### Options

* `--db-filename`, `-f` (string) -
  Path to file for persistent Temporal state store.
  By default, Workflow Executions are lost when the server process dies.
* `--namespace`, `-n` (string[]) -
  Namespaces to be created at launch.
  The "default" Namespace is always created automatically.
* `--port`, `-p` (int) -
  Port for the frontend gRPC Service.
  Default: 7233.
* `--http-port` (int) -
  Port for the frontend HTTP API service.
  Default is off.
* `--metrics-port` (int) -
  Port for '/metrics'.
  Default is off.
* `--ui-port` (int) -
  Port for the Web UI.
  Default is '--port' value + 1000.
* `--headless` (bool) -
  Disable the Web UI.
* `--ip` (string) -
  IP address bound to the frontend Service.
  Default: localhost.
* `--ui-ip` (string) -
  IP address bound to the WebUI.
  Default is same as '--ip' value.
* `--ui-asset-path` (string) -
  UI custom assets path.
* `--ui-codec-endpoint` (string) -
  UI remote codec HTTP endpoint.
* `--sqlite-pragma` (string[]) -
  SQLite pragma statements in "PRAGMA=VALUE" format.
* `--dynamic-config-value` (string[]) -
  Dynamic configuration value in "KEY=VALUE" format.
  KEY is an unquoted string identifier.
  VALUE is a JSON string.
  For example, "YourKey=\"YourStringValue\"".
* `--log-config` (bool) -
  Log the server config to stderr.

### temporal task-queue: Manage Task Queues

Inspect and update [Task Queues](/workers#task-queue), the queues that Worker entities poll for
Workflow and Activity tasks:

```
temporal task-queue [command] [command options]
```

For example:

```
temporal task-queue describe --task-queue "YourTaskQueueName"
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal task-queue describe: Show Workers that have recently polled on a Task Queue

The `temporal task-queue describe` command provides [poller](/dev-guide/worker-performance#poller-count)
information for a given [Task Queue](/workers#task-queue).

The [Temporal Server](/clusters) records the last time of each poll request. A `LastAccessTime` value
in excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the
Worker has shut down. Temporal [Workers](/workers) are removed if 5 minutes have passed since the last poll
request

Information about the Task Queue can be returned to troubleshoot server issues

`temporal task-queue describe --task-queue YourTaskQueue --task-queue-type "activity"`

Use the options listed below to modify what this command returns

#### Options

* `--task-queue`, `-t` (string) -
  Task queue name. Required
* `--task-queue-type` (string-enum) -
  Task Queue type. Options: workflow, activity.
  Default: workflow
* `--partitions` (int) -
  Query for all partitions up to this number (experimental+temporary feature).
  Default: 1

### temporal task-queue get-build-id-reachability: Show which Build IDs are available on a Task Queue

This command can tell you whether or not Build IDs may be used for new, existing, or closed workflows. Both the '--build-id' and '--task-queue' flags may be specified multiple times. If you do not provide a task queue, reachability for the provided Build IDs will be checked against all task queues
  Can be passed multiple times.


#### Options

* `--build-id` (string[]) -
  Which Build ID to get reachability information for. May be specified multiple times
  Can be passed multiple times.

* `--reachability-type` (string-enum) -
  Specify how you'd like to filter the reachability of Build IDs. Valid choices are `open` (reachable by one or more open workflows), `closed` (reachable by one or more closed workflows), or `existing` (reachable by either). If a Build ID is reachable by new workflows, that is always reported. Options: open, closed, existing.
  Default: existing
* `--task-queue`, `-t` (string[]) -
  Which Task Queue(s) to constrain the reachability search to. May be specified multiple times
  Can be passed multiple times.

### temporal task-queue get-build-ids: Show worker Build ID versions on a Task Queue

Fetch the sets of compatible build IDs associated with a Task Queue and associated information

#### Options

* `--task-queue`, `-t` (string) -
  Task queue name. Required
* `--max-sets` (int) -
  Limits how many compatible sets will be returned. Specify 1 to only return the current default major version set. 0 returns all sets. (default: 0).
  Default: 0.

### temporal task-queue list-partition: List a Task Queue's partitions

The temporal task-queue list-partition command displays the partitions of a Task Queue, along with the matching node they are assigned to

#### Options

* `--task-queue`, `-t` (string) -
  Task queue name. Required

### temporal task-queue update-build-ids: Operations to manage Build ID versions on a Task Queue

Provides various commands for adding or changing the sets of compatible build IDs associated with a Task Queue. See the help of each sub-command for more

### temporal task-queue update-build-ids add-new-compatible: Add a new build ID compatible with an existing ID to a Task Queue's version sets

The new build ID will become the default for the set containing the existing ID. See per-flag help for more

#### Options

* `--build-id` (string) -
  The new build ID to be added. Required
* `--task-queue`, `-t` (string) -
  Name of the Task Queue. Required
* `--existing-compatible-build-id` (string) -
  A build ID which must already exist in the version sets known by the task queue. The new ID will be stored in the set containing this ID, marking it as compatible with the versions within. Required
* `--set-as-default` (bool) -
  When set, establishes the compatible set being targeted as the overall default for the queue. If a different set was the current default, the targeted set will replace it as the new default.
  Defaults to false.

### temporal task-queue update-build-ids add-new-default: Add a new default (incompatible) build ID to a Task Queue's 7 sets

Creates a new build ID set which will become the new overall default for the queue with the provided build ID as its only member. This new set is incompatible with all previous sets/versions

#### Options

* `--build-id` (string) -
  The new build ID to be added. Required
* `--task-queue`, `-t` (string) -
  Name of the Task Queue. Required

### temporal task-queue update-build-ids promote-id-in-set: Promote a build ID to become the default for its containing set

New tasks compatible with the set will be dispatched to the default ID

#### Options

* `--build-id` (string) -
  An existing build ID which will be promoted to be the default inside its containing set. Required
* `--task-queue`, `-t` (string) -
  Name of the Task Queue. Required

### temporal task-queue update-build-ids promote-set: Promote a build ID set to become the default for a Task Queue

If the set is already the default, this command has no effect

#### Options

* `--build-id` (string) -
  An existing build ID whose containing set will be promoted. Required
* `--task-queue`, `-t` (string) -
  Name of the Task Queue. Required


### temporal workflow: Start, list, and operate on Workflows

[Workflow](/workflows) commands perform operations on [Workflow Executions](/workflows#workflow-execution)

Workflow commands use this syntax: `temporal workflow COMMAND [ARGS]`

#### Options set for client:

* `--address` (string) -
  Temporal server address.
  Default: 127.0.0.1:7233. Env: TEMPORAL_ADDRESS.
* `--namespace`, `-n` (string) -
  Temporal server namespace.
  Default: default. Env: TEMPORAL_NAMESPACE
* `--api-key` (string) -
  Sets the API key on requests. Env: TEMPORAL_API_KEY
* `--grpc-meta` (string[]) -
  HTTP headers to send with requests (formatted as "KEY=VALUE")
* `--tls` (bool) -
  Enable TLS encryption without additional options such as mTLS or client certificates. Env:
  TEMPORAL_TLS
* `--tls-cert-path` (string) -
  Path to x509 certificate. Env: TEMPORAL_TLS_CERT
* `--tls-key-path` (string) -
  Path to private certificate key. Env: TEMPORAL_TLS_KEY
* `--tls-ca-path` (string) -
  Path to server CA certificate. Env: TEMPORAL_TLS_CA
* `--tls-cert-data` (string) -
  Data for x509 certificate. Exclusive with -path variant. Env: TEMPORAL_TLS_CERT_DATA
* `--tls-key-data` (string) -
  Data for private certificate key. Exclusive with -path variant. Env: TEMPORAL_TLS_KEY_DATA
* `--tls-ca-data` (string) -
  Data for server CA certificate. Exclusive with -path variant. Env: TEMPORAL_TLS_CA_DATA
* `--tls-disable-host-verification` (bool) -
  Disables TLS host-name verification. Env:
  TEMPORAL_TLS_DISABLE_HOST_VERIFICATION
* `--tls-server-name` (string) -
  Overrides target TLS server name. Env: TEMPORAL_TLS_SERVER_NAME
* `--codec-endpoint` (string) -
  Endpoint for a remote Codec Server. Env: TEMPORAL_CODEC_ENDPOINT
* `--codec-auth` (string) -
  Sets the authorization header on requests to the Codec Server. Env: TEMPORAL_CODEC_AUTH

### temporal workflow cancel: Cancel a Workflow Execution

The `temporal workflow cancel` command is used to cancel a [Workflow Execution]([Workflow Executions](/workflows#workflow-execution))
Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History. A new
Command Task will be scheduled, and the Workflow Execution will perform cleanup work

Executions may be cancelled by [Workflow ID](/workflows#workflow-id):
```
temporal workflow cancel --workflow-id YourWorkflowId
```

...or in bulk via a visibility query [list filter](/visibility#list-filter):
```
temporal workflow cancel --query YourQuery
```

Use the options listed below to change the behavior of this command

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch).

### temporal workflow count: Count Workflow Executions

The `temporal workflow count` command returns a count of [Workflow Executions](/workflows#workflow-execution).

Use the options listed below to change the command's behavior

#### Options

* `--query`, `-q` (string) -
  Filter results using a SQL-like query.

### temporal workflow delete: Delete a Workflow Execution

The `temporal workflow delete` command is used to delete a specific [Workflow Executions](/workflows#workflow-execution).
This asynchronously deletes a workflow's [Event History](/workflows#event-history).
If the [Workflow Executions](/workflows#workflow-execution) is Running, it will be terminated before deletion.

```
temporal workflow delete \
		--workflow-id YourWorkflowId \
```

Use the options listed below to change the command's behavior

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch).

### temporal workflow describe: Show information about a Workflow Execution

The `temporal workflow describe` command shows information about a given
[Workflow Executions](/workflows#workflow-execution).

This information can be used to locate Workflow Executions that weren't able to run successfully

`temporal workflow describe --workflow-id meaningful-business-id`

Output can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points

`temporal workflow describe --workflow-id meaningful-business-id --raw true --reset-points true`

Use the command options below to change the information returned by this command

#### Options set for workflow reference:

* `--workflow-id`, `-w` (string) -
  Workflow ID.
  Required.
* `--run-id`, `-r` (string) -
  Run ID.

#### Options

* `--reset-points` (bool) -
  Only show auto-reset points.
* `--raw` (bool) -
  Print properties without changing their format.

### temporal workflow execute: Start a new Workflow Execution and prints its progress

The `temporal workflow execute` command starts a new [Workflow Executions](/workflows#workflow-execution) and
prints its progress. The command completes when the Workflow Execution completes

Single quotes('') are used to wrap input as JSON.

```
temporal workflow execute
		--workflow-id meaningful-business-id \
		--type YourWorkflow \
		--task-queue YourTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options

* `--event-details` (bool) -
  If set when using text output, this will print the event details instead of just the event
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

Use the options listed below to change the command's behavior

#### Options

* `--source`, `-s` (string) -
  Path to the input file.
  Required.
* `--target`, `-t` (string) -
  Path to the output file, or standard output if not set.

### temporal workflow list: List Workflow Executions based on a Query

The `temporal workflow list` command provides a list of [Workflow Executions](/workflows#workflow-execution)
that meet the criteria of a given [Query](/workflows#query).
By default, this command returns up to 10 closed Workflow Executions

`temporal workflow list --query YourQuery`

The command can also return a list of archived Workflow Executions

`temporal workflow list --archived`

Use the command options below to change the information returned by this command

#### Options

* `--query`, `-q` (string) -
  Filter results using a SQL-like query.
* `--archived` (bool) -
  If set, will only query and list archived workflows instead of regular workflows.
* `--limit` (int) -
  Max count to list.

### temporal workflow query: Query a Workflow Execution

The `temporal workflow query` command is used to [Query](/workflows#query) a
[Workflow Execution](/workflows#workflow-execution)
by [Workflow ID](/workflows#workflow-id).

```
temporal workflow query \
		--workflow-id YourWorkflowId \
		--name YourQuery \
		--input '{"YourInputKey": "YourInputValue"}'
```

Use the options listed below to change the command's behavior

#### Options

* `--type` (string) -
  Query Type/Name.
  Required.
* `--reject-condition` (string-enum) -
  Optional flag for rejecting Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [payload input](#options-set-for-payload-input).
Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow reset: Reset a Workflow Execution to an older point in history

The temporal workflow reset command resets a [Workflow Execution](/workflows#workflow-execution)
A reset allows the Workflow to resume from a certain point without losing its parameters or [Event History](/workflows#event-history)

The Workflow Execution can be set to a given [Event Type](/workflows#event):
```
temporal workflow reset --workflow-id meaningful-business-id --type LastContinuedAsNew
```

...or a specific any Event after `WorkflowTaskStarted`
```
temporal workflow reset --workflow-id meaningful-business-id --event-id YourLastEvent
```
For batch reset only FirstWorkflowTask, LastWorkflowTask or BuildId can be used. Workflow ID, run ID and event ID
should not be set
Use the options listed below to change reset behavior

#### Options

* `--workflow-id`, `-w` (string) -
  Workflow ID. Required for non-batch reset operations.
* `--run-id`, `-r` (string) -
  Run ID.
* `--event-id`, `-e` (int) -
  The Event ID for any Event after `WorkflowTaskStarted` you want to reset to (exclusive). It can be `WorkflowTaskCompleted`, `WorkflowTaskFailed` or others.
* `--reason` (string) -
  The reason why this workflow is being reset.
  Required.
* `--reapply-type` (string-enum) -
  Event types to reapply after the reset point. Options: All, Signal, None.
  Default: All.
* `--type`, `-t` (string-enum) -
  Event type to which you want to reset. Options: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.
* `--build-id` (string) -
  Only used if type is BuildId. Reset the first workflow task processed by this build ID. Note that by default, this reset is allowed to be to a prior run in a chain of continue-as-new.
* `--query`, `-q` (string) -
  Start a batch reset to operate on Workflow Executions with given List Filter.
* `--yes`, `-y` (bool) -
  Override confirmation request before reset.
  Only allowed when `--query` is present.

### temporal workflow show: Show Event History for a Workflow Execution

The `temporal workflow show` command provides the [Event History](/workflows#event-history) for a
[Workflow Execution](/workflows#workflow-execution). With JSON output specified, this output can be given to
an SDK to perform a replay

Use the options listed below to change the command's behavior

#### Options

* `--follow`, `-f` (bool) -
  Follow the progress of a Workflow Execution in real time (does not apply
  to JSON output).

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow signal: Signal a Workflow Execution

The `temporal workflow signal` command is used to [Signal](/workflows#signal) a
[Workflow Execution](/workflows#workflow-execution) by [Workflow ID](/workflows#workflow-id)

```
temporal workflow signal \
		--workflow-id YourWorkflowId \
		--name YourSignal \
		--input '{"YourInputKey": "YourInputValue"}'
```

Use the options listed below to change the command's behavior

#### Options

* `--name` (string) -
  Signal Name. Required

Includes options set for [payload input](#options-set-for-payload-input).

#### Options set for single workflow or batch:

* `--workflow-id`, `-w` (string) -
  Workflow ID. Either this or --query must be set.
* `--run-id`, `-r` (string) -
  Run ID. Cannot be set when --query is set.
* `--query`, `-q` (string) -
  Start a batch to operate on Workflow Executions with given List Filter. Either --query or --workflow-id must be set.
* `--reason` (string) -
  Reason to perform batch. Only allowed if query is present unless the command specifies
  otherwise.
  Defaults to message with the current user's name.
* `--yes`, `-y` (bool) -
  Override confirmation request before signalling.
  Only allowed when `--query` is present.

### temporal workflow stack: Show the stack trace of a Workflow Execution

The `temporal workflow stack` command [Queries](/workflows#query) a
[Workflow Execution](/workflows#workflow-execution) with `__stack_trace` as the query type
This returns a stack trace of all the threads or routines currently used by the workflow, and is
useful for troubleshooting

```
temporal workflow stack --workflow-id YourWorkflowId
```

Use the options listed below to change the command's behavior

#### Options

* `--reject-condition` (string-enum) -
  Optional flag for rejecting Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow start: Start a new Workflow Execution

The `temporal workflow start` command starts a new [Workflow Execution](/workflows#workflow-execution). The
Workflow and Run IDs are returned after starting the [Workflow](/workflows).


```
temporal workflow start \
		--workflow-id meaningful-business-id \
		--type YourWorkflow \
		--task-queue YourTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options set for shared workflow start:

* `--workflow-id`, `-w` (string) -
  Workflow ID.
* `--type` (string) -
  Workflow Type name.
  Required.
* `--task-queue`, `-t` (string) -
  Workflow Task queue.
  Required.
* `--run-timeout` (duration) -
  Fail a Workflow Run if it takes longer than `DURATION`.
* `--execution-timeout` (duration) -
  Fail a WorkflowExecution if it takes longer than `DURATION`, including retries and ContinueAsNew tasks.
* `--task-timeout` (duration) -
  Fail a Workflow Task if it takes longer than `DURATION`. (Start-to-close timeout for a Workflow Task.) Default: 10s.
* `--search-attribute` (string[]) -
  Passes Search Attribute in "KEY=VALUE" format. Use valid JSON formats for value.
* `--memo` (string[]) -
  Passes Memo in "KEY=VALUE" format. Use valid JSON formats for value.

#### Options set for workflow start:

* `--cron` (string) -
  Cron schedule for the Workflow. Deprecated -
  use schedules instead.
* `--fail-existing` (bool) -
  Fail if the Workflow already exists.
* `--start-delay` (duration) -
  Wait before starting the Workflow. Cannot be used with a cron schedule. If the
  Workflow receives a signal or update before the delay has elapsed, it will start immediately.
* `--id-reuse-policy` (string) -
  Allow the same Workflow ID to be used in a new Workflow Execution. Options:
  AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.

#### Options set for payload input:

* `--input`, `-i` (string[]) -
  Input value (default JSON unless --input-meta is non-JSON encoding). Can
  be passed multiple times for multiple arguments. Cannot be combined with --input-file.
  Can be passed multiple times. 
* `--input-file` (string[]) -
  Read `PATH` as the input value (JSON by default unless --input-meta is non-JSON
  encoding). Can be passed multiple times for multiple arguments. Cannot be combined with --input.
  Can be passed multiple times.
* `--input-meta` (string[]) -
  Metadata for the input payload, specified as a "KEY=VALUE" pair. If KEY is "encoding", overrides the
  default of "json/plain". Pass multiple --input-meta options to set multiple pairs.
  Can be passed multiple times.
* `--input-base64` (bool) -
  Assume inputs are base64-encoded and attempt to decode them.

### temporal workflow terminate: Terminate a Workflow Execution

The `temporal workflow terminate` command is used to terminate a
[Workflow Execution](/workflows#workflow-execution). Canceling a running Workflow Execution records a
`WorkflowExecutionTerminated` event as the closing Event in the workflow's Event History. Workflow code is oblivious to
termination. Use `temporal workflow cancel` if you need to perform cleanup in your workflow

Executions may be terminated by [Workflow ID](/workflows#workflow-id) with an optional reason:
```
temporal workflow terminate [--reason my-reason] --workflow-id YourWorkflowId
```

...or in bulk via a visibility query [list filter](/visibility#list-filter):
```
temporal workflow terminate --query YourQuery
```

Use the options listed below to change the behavior of this command.

#### Options

* `--workflow-id`, `-w` (string) -
  Workflow ID. Either this or query must be set.
* `--run-id`, `-r` (string) -
  Run ID. Cannot be set when query is set.
* `--query`, `-q` (string) -
  Start a batch to terminate Workflow Executions with the `QUERY` List Filter. Either this or
  Workflow ID must be set.
* `--reason` (string) -
  Reason for termination.
  Defaults to message with the current user's name.
* `--yes`, `-y` (bool) -
  Override confirmation request before termination.
  Only allowed when `--query` is present.

### temporal workflow trace: Interactively show the progress of a Workflow Execution

The `temporal workflow trace` command displays the progress of a [Workflow Execution](/workflows#workflow-execution) and its child workflows with a real-time trace.
This view provides a great way to understand the flow of a workflow.

Use the options listed below to change the behavior of this command.

#### Options

* `--fold` (string[]) -
  Fold Child Workflows with the specified `STATUS`. To specify multiple statuses, pass --fold multiple times. This will reduce the amount of information fetched and displayed. Case-insensitive. Ignored if --no-fold supplied. Available values: running, completed, failed, canceled, terminated, timedout, continueasnew.
  Can be passed multiple times.
* `--no-fold` (bool) -
  Disable folding. All Child Workflows within the set depth will be fetched and displayed.
* `--depth` (int) -
  Fetch up to N Child Workflows deep. Use -1 to fetch child workflows at any depth.
  Default: -1.
* `--concurrency` (int) -
  Fetch up to N Workflow Histories at a time.
  Default: 10.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow update: Update a running workflow synchronously

The `temporal workflow update` command is used to synchronously [Update](/workflows#update) a
[Workflow Execution](/workflows#workflow-execution) by [Workflow ID](/workflows#workflow-id)

```
temporal workflow update \
		--workflow-id YourWorkflowId \
		--name YourUpdate \
		--input '{"Input": "As-JSON"}'
```

Use the options listed below to change the command's behavior

#### Options

* `--name` (string) -
  Update Name.
  Required.
* `--workflow-id`, `-w` (string) -
  Workflow `ID`.
  Required.
* `--update-id` (string) -
  Update `ID`. If unset, default to a UUID.
* `--run-id`, `-r` (string) -
  Run `ID`. If unset, the currently running Workflow Execution receives the Update.
* `--first-execution-run-id` (string) -
  Send the Update to the last Workflow Execution in the chain that started
  with `ID`.

Includes options set for [payload input](#options-set-for-payload-input).
