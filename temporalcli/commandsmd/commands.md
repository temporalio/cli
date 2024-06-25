# Temporal CLI Commands

Commands for the Temporal CLI

<!--

This document is automatically parsed.
Follow these rules.

IN-HOUSE STYLE

* Wording and grammar.
  * Run a spell check.
  * Be clear and concise.
  * Don't reword the command in the short description. Use distinct supplementary language.
    * Yes: temporal workflow delete: Remove Workflow Execution
    * No: temporal workflow delete: Delete Workflow
  * Re-use and adapt existing wording and phrases wherever possible.
  * Word command short descriptions as if they began "This command will..."
    Use sentence casing for the short description.
  * ID is fully capitalized in text ("the Workflow ID") and Id in [metasyntax](https://en.wikipedia.org/wiki/Metasyntactic_variable) (YourWorkflowId).
* Avoid parentheticals unless absolutely necessary.

* Wrapping:
  * Assume a user-visible line length of 80 characters max.
  * When declaring Options follow the wrapping style in this file.
    Splitting the items into multiple lines like this improves maintainability.
    It provides clearer diffs for changes.
  * Hand-wrap your long descriptions as they appear for users.
    NOTE: this may change in the future with automatic wrapping.

* Ordering:
  * Commands: Use alphabetical order for commands.
  * Options: Most commonly used options first.

* Punctuation:
  * With the exception of short descriptions and triple-quoted code-fenced samples, end everything with a period.
    Do _not_ use periods for short descriptions.
    Skipping required periods may cause errors in parsing this file.
    Caution: Don't forget the period at the end of "Includes options set" lines.
  * Introduce triple-quoted code-fenced samples with a colons.
    Avoid using 'for example' unless there's no other way to introduce code.

* Code, flags, and keys
  * Demonstrate at least one example invocation of the command in every long description.
  * Include the most commonly used patterns in long descriptions so users don't have to call help at multiple invocation levels.
  * Avoid deprecated period-delineated versions of environment-specific keys.
    * Yes:
          ```
          temporal env set \
              --env prod \
              --key tls-cert-path \
              --value /home/my-user/certs/cluster.cert`
          ```
    * No: `temporal env set prod.tls-cert-path /home/my-user/certs/cluster.cert`.
  * Split invocation samples to multiple lines.
    Use one option or flag per line, as in the above example.
    Use a single space and a backslash to continue the invocation.
    Use 4-space indentation.
  * Always use long options and flags for invocation examples.
    * Yes: `--namespace`
    * No:  `-n`
  * When commands have a single command-level option, include it the mandatory example.
  * Use square bracket overviews to present how complex commands will be used.
    * Yes: temporal operator [command] [subcommand] [options]
    Commands with subcommands can't be run on their own.
    Because of this, always use full command examples.
  * Use square brackets to highlight optional elements, especially when long descriptions would suffer from two very similar command invocations.
    * Yes: temporal operator cluster describe [--detail]
  * Use YourEnvironment, YourNamespace, etc as unquoted metasyntactic variable stand-ins.
    Respectful metasyntax describes the role of the stand-in.
    * Yes: --workflow-id YourWorkflowId
    * No: --workflow-id your-work-id, --workflow-id "
  * For JSON input, use single quotes to encase interior double quotes.
    Otherwise, in the rare case it is needed, prefer double quotes.
  * When presenting options use a space rather than equal to set them.
    This is more universally supported and consistent with POSIX guidelines.
    * Yes: `temporal command --namespace YourNamespace`.
    * No: `temporal command --namespace=YourNamespace`.
    Note: in this utility's current incarnation, Boolean options must be set with an equal sign.
    Since Booleans can be treated like flags, avoid using assigned values in samples.
    * Yes: `--detail`
    * No: `--detail=true`

For options and flags:

* When options and flags can be passed multiple times, say so explicitly in the usage text: "Can be passed multiple times."
* Never rely on the flag type (e.g. `string`, `bool`, etc.) being shown to the user.
  It is replaced/hidden when a `META-VARIABLE` is used.
* Where possible, use a `META-VARIABLE` (all caps and wrapped in `\``s) to describe/reference content passed to an option.
* Limit `code spans` to meta-variables.
  To reference other options or specify literal values, use double quotes.

COMMAND ENTRY OVERVIEW

A Command entry uses the following format:

    ### <command>: <short-description>

    <long description>

    (optional command implementation configuration)

    #### Options 
    (or)
    #### Options set for options set name:

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
* End H4 Options set for declarations with a colon.

COMMAND LISTING

* Use H3 `### <command>: <short-description>` headings for each command.
  * One line only. Use the complete command path with parent commands.
  * Short descriptions do not repeat the command literally.
  * Use square-bracketed delimited arguments to document positional arguments.
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
  * [space-holder for new element being introduced]

OPTIONS SECTION

* Start the optional Options section with an H4 `#### Options` header line.
* Follow the header declaration with a list of options.
  The individual option definition syntax follows below the header declaration.
* You must include at least one option.
  Otherwise, `gen-commands` will complete but the CLI utility will not run.
* To incorporate an existing options set, add a line below options like this, remembering to end every `Includes options set for` line with a period:

  ```
  Includes options set for [<options-set-name>](#options-set-for-<options-set-link-name>).
  ```

  For example:

  ```
  Includes options set for [client](#options-set-for-client).
  ```
  
  You may include multiple option sets.

  An options set declaration is the equivalent of pasting those options into the bulleted options list.

  * Options that are similar but slightly different don't need to be in option sets.
  * Every "Option Set for" declaration links to the H4 entry that supplies the inherited options.
  * Reserve option sets for when the behavior of the option is the same across commands.
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
  In reality, the short option `-a` does not actually exist with `--address`:

  ```
  * `--address`, `-a` (string) -
    Temporal Service endpoint.
    Connect to the Temporal Service at `HOST:PORT`.
    Default: 127.0.0.1:7233.
    Env: TEMPORAL_ADDRESS.
  ```

* Each option listing includes a long option with a double dash and a meaningful name.
* [Optional] A short option uses a single dash and a short string.
  When used, separate the long and short option with a comma and a space.
* Backtick every option and short description.
  Include the dash or dashes within the ticks.
  For example: `` `--workflow-id`, `-w` ``.
* A data type follows option names indicating the required value type for the option.
  The type is `bool`, `duration`, `int`, `string`, `string[]`, `string-enum`, or `timestamp`. (_TODO: more_.).
  Always parenthesize data types.
  For example: `` `--raw` (bool) ``.
* A dash follows the data type, with a space on either side.
* The short description is free-form text and follows the dash.
  Take care not to match trailing attributes.
  Take care not to parrot the command invocation.
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
run a local Temporal Service, start [Workflow Executions](/workflows#workflow-execution), pass messages to
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

Update an [Activity](/activities)'s state to as completed or failed. This marks an
Activity as successfully finished or as having encountered an error:

```
temporal activity complete \
    --activity-id=YourActivityId \
    --workflow-id=YourWorkflowId \
    --result='{"YourResultKey": "YourResultValue"}'
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal activity complete: Complete an Activity

Complete an [Activity](/activities), marking it as successfully finished. Specify the 
Activity ID and include a JSON result for the returned value:

```
temporal activity complete \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId \
    --result '{"YourResultKey": "YourResultVal"}'
```

#### Options

* `--activity-id` (string) -
  Activity ID to complete.
  Required.
* `--result` (string) -
  Result `JSON` to return.
  Required.
* `--identity` (string) -
  Identity of the user submitting this request.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal activity fail: Fail an Activity

Fail an [Activity](/activities), marking it as having encountered an error. Specify the
Activity and Workflow IDs:

```
temporal activity fail \
    --activity-id YourActivityId \
    --workflow-id YourWorkflowId
```

#### Options

* `--activity-id` (string) -
  Activity ID to fail.
  Required.
* `--detail` (string) -
  Reason for failing the Activity (JSON).
* `--identity` (string) -
  Identity of the user submitting this request.
* `--reason` (string) -
  Reason for failing the Activity.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal batch: Manage batch jobs

A batch job executes a command on multiple [Workflow Executions](/workflows#workflow-execution) at once. 
Use an SQL-like `--query` to select which Workflow Executions to include:

```
temporal batch workflow cancel \
  --query 'ExecutionStatus = "Running" AND WorkflowType="YourWorkflow"' \
  --reason "Testing"
```

The `batch` keyword is optional when using a `--query`. The following command
is functionally identical to the previous one:

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
temporal batch describe \
    --job-id YourJobId
```

#### Options

* `--job-id` (string) -
  Batch job ID.
  Required.

### temporal batch list: List all batch jobs

Return a list of batch jobs on the [Service](/clusters) or within a 
single [Namespace](/namespaces). For example, list the batch jobs for
"YourNamespace":

```
temporal batch list \
    --namespace YourNamespace
```

#### Options

* `--limit` (int) -
  Maximum number of batch jobs to display.

### temporal batch terminate: Forcefully end a batch job

Terminate a batch job with the provided job ID. You must provide a reason
for the termination. The [Service](/clusters) stores this explanation as
metadata for the termination event for later reference:

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
for you. You can set up distinct environments like "dev" and "prod" for 
convenience:

```
temporal env set \
    --env prod \
    --key address \
    --value production.f45a2.tmprl.cloud:7233
```

Each environment is isolated. Changes to "prod" presets won't affect "dev".

For easiest use, set a `TEMPORAL_ENV` environment variable in your shell.
The Temporal CLI checks for an `--env` option first, then checks the shell.
If neither is set, most commands use `default` environment presets if
available.

### temporal env delete: Delete an environment or environment property

Remove a presets environment entirely _or_ remove a key-value pair within
an environment. If you don't specify an environment (with `--env` or by
setting the `TEMPORAL_ENV` variable), this command updates the default 
environment:

```
temporal env delete \
    --env YourEnvironment
```

or

```
temporal env delete \
    --env prod \
    --key tls-key-path
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
temporal env get \
    --env YourEnvironment
```

Print a single property:

```
temporal env get \
    --env YourEnvironment \
    --key YourPropertyKey
```

If you don't specify an environment (with `--env` or by setting the 
`TEMPORAL_ENV` variable), this command lists properties of the "default"
environment.

<!--
* maximum-args=1
-->

#### Options

* `--key`, `-k` (string) -
  Property name.

### temporal env list: Show environment names

List the environments you have set up on your local computer. Environments 
are stored in "$HOME/.config/temporalio/temporal.yaml".

### temporal env set: Set environment properties

Assign a value to a property key and store it to an environment:

```
temporal env set \
    --env environment \
    --key property \
    --value value
```

If you don't specify an environment (with `--env` or by setting the 
`TEMPORAL_ENV` variable), this command sets properties in the default 
environment.

Storing keys with CLI option names lets the CLI automatically set those
options for you. This reduces effort and helps avoid typos when issuing
commands.

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

Operator commands manage and fetch information about [Namespace](/namespaces), 
[Search Attributes](/visibility#search-attribute), and [Temporal Services](/clusters):

```
temporal operator [command] [subcommand] [options]
```

For example, to show information about the Temporal Service at the
default address (localhost):

```
temporal operator cluster describe
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal operator cluster: Manage a Temporal Cluster

Perform operator actions on [Temporal Services](/clusters) (also known as Clusters).

```
temporal operator cluster [subcommand] [options]
```

For example to check Service/Cluster health:

```
temporal operator cluster health
```

### temporal operator cluster describe: Show Temporal Cluster information

View information about a [Temporal Cluster](/clusters) (Service), including Cluster Name,
persistence store, and visibility store. Add `--detail` for additional info:

```
temporal operator cluster describe [--detail]
```

#### Options

* `--detail` (bool) -
  Show history shard count and Cluster/Service version information.

### temporal operator cluster health: Check Temporal Service health

View information about the health of a [Temporal Service](/clusters):

```
temporal operator cluster health
```

### temporal operator cluster list: Show Temporal Clusters

Print a list of [Temporal Clusters](/clusters) (Services) registered to this
system. Report details include the Cluster's name, ID, address, History Shard
count, Failover version, and availability:

```
temporal operator cluster list [--limit max-count]
```

#### Options

* `--limit` (int) -
    Maximum number of Clusters to display.

### temporal operator cluster remove: Remove a Temporal Cluster

Remove a registered [Temporal Cluster](/clusters) (Service) from this system.

```
temporal operator cluster remove \
    --name YourClusterName
```

#### Options

* `--name` (string) -
  Cluster/Service name.
  Required.

### temporal operator cluster system: Show Temporal Cluster info


Show Temporal Server information for [Temporal Clusters](/clusters) (Service):
Server version, scheduling support, and more. This information helps diagnose
problems with the Temporal Server.

The command defaults to the local Service. Otherwise, use the 
`--frontend-address` option to specify a Cluster (Service) endpoint:

```
temporal operator cluster system \ 
    --frontend-address "YourRemoteEndpoint:YourRemotePort"
``` 

### temporal operator cluster upsert: Add/update a Temporal Cluster

Add, remove, or update a registered ("remote") [Temporal Cluster](/clusters) (Service).

```
temporal operator cluster upsert [options]
```

For example:

```
temporal operator cluster upsert \
    --frontend-address "YourRemoteEndpoint:YourRemotePort"
    --enable-connection false
```

#### Options

* `--frontend-address` (string) -
  Remote endpoint.
  Required.
* `--enable-connection` (bool) -
  Set the connection to "enabled".

### temporal operator namespace: Namespace operations

Manage [Temporal Cluster](/clusters) (Service) [Namespaces](/namespaces):

```
temporal operator namespace [command] [command options]
```

For example:

```
temporal operator namespace create \
    --namespace YourNewNamespaceName
```

### temporal operator namespace create: Register a new Namespace

Create a new [Namespace](/namespaces) on the [Temporal Service](/clusters):

```
temporal operator namespace create \
    --namespace YourNewNamespaceName \
    [options]
````

Create a Namespace with multi-region data replication:

```
temporal operator namespace create \
    --global \
    --namespace YourNewNamespaceName
```

Configure settings like retention and Visibility Archival State as needed.
For example, the Visibility Archive can be set on a separate URI:

```
temporal operator namespace create \
    --retention 5d \
    --visibility-archival-state enabled \
    --visibility-uri YourURI \
    --namespace YourNewNamespaceName
```

Note: URI values for archival states can't be changed once enabled.

<!--
* maximum-args=1
-->

#### Options

* `--active-cluster` (string) -
  Active Cluster (Service) name.
* `--cluster` (string[]) -
  Cluster (Service) names for Namespace creation.
  Can be passed multiple times.
* `--data` (string) -
  Namespace data.
* `--description` (string) -
  Namespace description.
  Comma-separated 'KEY="VALUE"' string pairs.
* `--email` (string) -
  Owner email.
* `--global` (bool) -
  Enable multi-region data replication.
* `--history-archival-state` (string-enum) -
  History archival state.
  Options: disabled, enabled.
  Default: disabled.
* `--history-uri` (string) -
  Archive history to this `URI`.
  Once enabled, can't be changed.
* `--retention` (duration) -
  Time to preserve closed Workflows before deletion.
  Default: 72h.
* `--visibility-archival-state` (string-enum) -
  Visibility archival state.
  Options: disabled, enabled.
  Default: disabled
* `--visibility-uri` (string) -
  Archive visibility data to this `URI`.
  Once enabled, can't be changed.

### temporal operator namespace delete [namespace]: Delete a Namespace

Removes a [Namespace](/namespaces) from the [Service](/clusters).

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

### temporal operator namespace describe [namespace]: Describe a Namespace

Provide long-form information about a [Namespace](/namespaces) identified by its ID or name:

```
temporal operator namespace describe \
    --namespace-id YourNamespaceId
```

or

```
temporal operator namespace describe \
    --namespace YourNamespaceName
```

<!--
* maximum-args=1
-->

#### Options

* `--namespace-id` (string) -
  Namespace ID.

### temporal operator namespace list: List Namespaces

Display a detailed listing for all [Namespaces](/namespaces) on the [Service](/clusters):

```
temporal operator namespace list
``` 

### temporal operator namespace update: Update a Namespace

Update a [Namespace](/namespaces) using properties you specify.

```
temporal operator namespace update [options]
``` 

Assign a Namespace's active Cluster (Service):

```
temporal operator namespace update \
    --namespace YourNamespaceName \
    --active-cluster NewActiveCluster
```

Promote a Namespace for multi-region data replication:

```
temporal operator namespace update \
    --namespace YourNamespaceName \
    --promote-global
```

You may update archives that were previously enabled or disabled.
Note: URI values for archival states can't be changed once enabled.

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
  Active Cluster (S) name.
* `--cluster` (string[]) -
  Cluster names.
* `--data` (string[]) -
  Set a 'KEY="VALUE"' string pair in Namespace data.
  KEY is a string, VALUE is JSON.
  Can be passed multiple times.
* `--description` (string) -
  Namespace description.
* `--email` (string) -
  Owner email.
* `--promote-global` (bool) -
  Enable multi-region data replication.
* `--history-archival-state` (string-enum) -
  History archival state.
  Options: disabled, enabled.
* `--history-uri` (string) -
  Archive history to this `URI`.
  Once enabled, can't be changed.
* `--retention` (duration) -
  Length of time a closed Workflow is preserved before deletion
* `--visibility-archival-state` (string-enum) -
  Visibility archival state. Options: disabled, enabled
* `--visibility-uri` (string) -
  Archive visibility data to this `URI`.
  Once enabled, can't be changed.

### temporal operator search-attribute: Search Attribute operations

Create, list, or remove [Search Attributes](/visibility#search-attribute)
fields stored in a Workflow Execution's metadata:

```
temporal operator search-attribute create \
    --name YourAttributeName \
    --type Keyword
```

Supported types include: Text, Keyword, Int, Double, Bool, Datetime, and
KeywordList.

### temporal operator search-attribute create: Add custom Search Attributes

Add one or more custom [Search Attributes](/visibility#search-attribute):

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

Display a list of active [Search Attributes](/visibility#search-attribute) that can be 
assigned or used with Workflow Queries. You can manage this list and add
attributes as needed:

```
temporal operator search-attribute list
```

### temporal operator search-attribute remove: Remove custom Search Attributes

Remove custom [Search Attributes](/visibility#search-attribute) from the options that can be
assigned or used with Workflow Queries.

```
temporal operator search-attribute remove \
    --name YourAttributeName
```

Remove attributes without confirmation:

```
temporal operator search-attribute remove \
    --name YourAttributeName \
    --yes
```

#### Options

* `--name` (string[]) -
  Search Attribute name.
  Required.
* `--yes`, `-y` (bool) -
  Don't prompt to confirm removal.
  

### temporal schedule: Perform operations on Schedules

Create, use, and update [Schedules](/workflows#schedule) that allow 
[Workflow Executions](/workflows#workflow-execution) to be created at specified times:

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
Use this command to fill in Workflow runs from when a [Schedule](/workflows#schedule)
was paused, before a Schedule was created, from the future, or to re-process a
previously executed interval.

Backfills require a Schedule ID and the time period covered by the request.
It's best to use the `BufferAll` or `AllowAll` policies to avoid conflicts
and ensure no [Workflow Executions](/workflows#workflow-execution) are
skipped.

For example:

```
  temporal schedule backfill \
    --schedule-id "YourScheduleId" \
    --start-time "2022-05-01T00:00:00Z" \
    --end-time "2022-05-31T23:59:59Z" \
    --overlap-policy BufferAll
```

The policies include:

* **AllowAll**: Allow unlimited concurrent Workflow Executions. This
  significantly speeds up the backfilling process on systems that support
  concurrency. Ensure running Workflow Executions do not interfere with each
  other.
* **BufferAll**: Buffer all incoming Workflow Executions while waiting for the
  running Workflow Execution to complete.
* **Skip**: If a previous Workflow Execution is still running, discard new
  Workflow Executions.
* **BufferOne**: Same as 'Skip' but buffer a single Workflow Execution to be
  run after the previous Execution completes. Discard other Workflow Executions.
* **CancelOther**: Cancel the running Workflow Execution and replace it with
  the incoming new Workflow Execution.
* **TerminateOther**: Terminate the running Workflow Execution and replace
  it with the incoming new Workflow Execution.

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

Create a new [Schedule](/workflows#schedule) on the Temporal Service. A Schedule automatically
starts new [Workflow Executions](/workflows#workflow-execution) at the times you specify.

For example:

```
  temporal schedule create \
    --schedule-id "YourScheduleId" \
    --calendar '{"dayOfWeek":"Fri","hour":"3","minute":"30"}' \
    --workflow-id "YourBaseWorkflowId" \
    --task-queue "YourTaskQueue" \
    --workflow-type "YourWorkflowType"
```

Schedules support any combination of `--calendar`, `--interval`, and `--cron`:

* Shorthand [`--interval`](https://docs.temporal.io/workflows#spec) strings.
  For example: 45m (every 45 minutes) or 6h/5h (every 6 hours, at the top of
  the 5th hour).
* JSON [`--calendar`](https://docs.temporal.io/workflows#spec), as in the
  preceding example.
* Unix-style [`--cron`](https://docs.temporal.io/workflows#cron-schedules) strings and [robfig](https://pkg.go.dev/github.com/robfig/cron/v3)
  declarations (@daily/@weekly/@every X/etc). For example, every Friday at 
  12:30 PM: `30 12 * * Fri`.

#### Options set for schedule configuration:

* `--calendar` (string[]) -
  [Calendar specification](https://docs.temporal.io/workflows#spec) in JSON.
  For example: `{"dayOfWeek":"Fri","hour":"17","minute":"5"}`.
* `--catchup-window` (duration) -
  Maximum catch-up time for when the Service is unavailable.
* `--cron` (string[]) -
  Calendar specification in [cron string format](https://docs.temporal.io/workflows#cron-schedules).
  For example: `"30 12 * * Fri"`.
* `--end-time` (timestamp) -
  Schedule end time
* `--interval` (string[]) -
  [Interval duration](https://docs.temporal.io/workflows#spec).
  For example, 90m, or 60m/15m to include phase offset.
* `--jitter` (duration) -
  Max difference in time from the specification.
  Vary the start time randomly within this amount.
* `--notes` (string) -
  Initial notes field value.
* `--paused` (bool) -
  Pause the Schedule immediately on creation.
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
  Set schedule Search Attributes using 'KEY="VALUE"' format.
  KEY is a string, VALUE is JSON.
  Can be passed multiple times.
* `--schedule-memo` (string[]) -
  Set a schedule memo using 'KEY="VALUE"' format.
  KEY is a string, VALUE is JSON.
  Can be passed multiple times.

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).
Includes options set for [shared-workflow-start](#options-set-for-shared-workflow-start).
Includes options set for [payload-input](#options-set-for-payload-input).

### temporal schedule delete: Remove a Schedule

Deletes a [Schedule](/workflows#schedule) on the front end Service:

```
temporal schedule delete \
    --schedule-id YourScheduleId
```

Removing a Schedule won't affect the [Workflow Executions](/workflows#workflow-execution) it started
that are still running. To cancel or terminate these Workflow Executions, use 
`temporal workflow delete` with the `TemporalScheduledById` Search Attribute instead.
 
#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule describe: Display Schedule state

Show a [Schedule](/workflows#schedule) configuration, including information about past,
current, and future Workflow runs:

```
temporal schedule describe \
    --schedule-id YourScheduleId
```

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule list: Display hosted Schedules

Lists the [Schedules](/workflows#schedule) hosted by a [Namespace](/namespaces):

```
temporal schedule list \
    --namespace YourNamespace
```

#### Options

* `--long`, `-l` (bool) -
  Show detailed information
* `--really-long` (bool) -
  Show extensive information in non-table form.

### temporal schedule toggle: Pause or unpause a Schedule

Pause or unpause a [Schedule](/workflows#schedule) by passing a flag with your
desired state:

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

The `--reason` text updates the Schedule's `notes` field for operations 
communication. It defaults to "(no reason provided)" if omitted. This field
is also visible on the Service Web UI. 

#### Options

* `--pause` (bool) -
  Pause the schedule.
* `--reason` (string) -
  Reason for pausing or unpausing a Schedule.
  Default: "(no reason provided)"
* `--unpause` (bool) -
  Unpause the schedule.

Includes options set for [schedule-id](#options-set-for-schedule-id).

### temporal schedule trigger: Immediately run a Schedule

Trigger a [Schedule](/workflows#schedule) to run immediately:

```
temporal schedule trigger \
    --schedule-id "YourScheduleId"
```

#### Options

Includes options set for [schedule-id](#options-set-for-schedule-id).
Includes options set for [overlap-policy](#options-set-for-overlap-policy).

### temporal schedule update: Update Schedule details

Update an existing [Schedule](/workflows#schedule) with new configuration details, including
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

Add persistence for [Workflow Executions](/workflows#workflow-execution) across runs:

```
temporal server start-dev \
    --db-filename path-to-your-local-persistent-store
```

Set the port from the front-end gRPC Service (7233 default):

```
temporal server start-dev \
    --port 7234 \
    --ui-port 8234 \
    --metrics-port 57271
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

Add persistence for [Workflow Executions](/workflows#workflow-execution) across runs:

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
  Port for the HTTP API service.
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
  Dynamic configuration value in 'KEY="VALUE"' format.
  KEY is a string, VALUE is JSON.
  For example, "YourKey=\"YourStringValue\"".
* `--log-config` (bool) -
  Log the server config to stderr.

### temporal task-queue: Manage Task Queues

Inspect and update [Task Queues](/workers#task-queue), the queues that
Workers poll for Workflow and Activity tasks:

```
temporal task-queue [command] [command options]
```

For example:

```
temporal task-queue describe \
    --task-queue "YourTaskQueueName"
```

#### Options

Includes options set for [client](#options-set-for-client).

### temporal task-queue describe: Show active Workers

Display a list of active Workers that have recently [polled](/dev-guide/worker-performance#poller-count)
a [Task Queue](/workers#task-queue). The [Temporal Server](/clusters) records each poll
request time. A `LastAccessTime` over one minute may indicate the Worker is
at capacity or has shut down. Temporal [Workers](/workers) are removed if 5 minutes
have passed since the last poll request:

```
temporal task-queue describe \
   --task-queue "YourTaskQueueName"
```

Workflow and Activity polling use separate Task Queues:

```
temporal task-queue describe \
    --task-queue YourTaskQueue \
    --task-queue-type "activity"`
```

#### Options

* `--task-queue`, `-t` (string) -
  Task queue name.
  Required.
* `--task-queue-type` (string-enum) -
  Task Queue type.
  Options: workflow, activity.
  Default: workflow.
* `--partitions` (int) -
  Query partitions 1 through `N`.
  Experimental/Temporary feature.
  Default: 1

### temporal task-queue get-build-id-reachability: Show Build ID availability

Show if a given Build ID can be used for new, existing, or closed workflows
in Namespaces that support Worker versioning:

```
temporal task-queue get-build-id-reachability \
    --build-id "2.0"
```

You can specify the `--build-id` and `--task-queue` flags multiple times.
If `--task-queue` is omitted, the command checks Build ID reachability
against all Task Queues.

#### Options

* `--build-id` (string[]) -
  One or more Build ID strings.
  Can be passed multiple times.
* `--reachability-type` (string-enum) -
  Reachability filter.
  `open`: reachable by one or more open workflows.
  `closed`: reachable by one or more closed workflows.
  `existing`: reachable by either.
  New Workflow Executions reachable by a Build ID are always reported.
  Options: open, closed, existing.
  Default: existing.
* `--task-queue`, `-t` (string[]) -
  Search only the specified task queue(s).
  Can be passed multiple times.

### temporal task-queue get-build-ids: Fetch Build ID versions

Fetch sets of compatible Build IDs for specified [Task Queues](/workers#task-queue) and
display their information:

```
temporal task-queue get-build-ids \
    --task-queue "YourTaskQueue"
```

This command is limited to Namespaces that enable Worker versioning.

#### Options

* `--task-queue`, `-t` (string) -
  Task queue name.
  Required.
* `--max-sets` (int) -
  Max return count.
  Use 1 for default major version.
  Use 0 for all sets.
  Default: 0.

### temporal task-queue list-partition: List Task Queue partitions

Display a [Task Queue](/workers#task-queue)'s partition list with assigned matching nodes:

```
temporal task-queue list-partition \
    --task-queue "YourTaskQueue"
```

#### Options

* `--task-queue`, `-t` (string) -
  Task Queue name.
  Required

### temporal task-queue update-build-ids: Manage Build IDs

Add or change a [Task Queue](/workers#task-queue)'s compatible Build IDs for Namespaces using
Worker versioning.

### temporal task-queue update-build-ids add-new-compatible: Add compatible Build ID

Add a compatible Build ID to a [Task Queue](/workers#task-queue)'s existing version set.
Provide an existing Build ID and a new Build ID:

```
temporal task-queue update-build-ids add-new-compatible \
    --task-queue YourTaskQueue \
    --existing-compatible-build-id YourExistingBuildId \
    --build-id YourNewBuildId
```

The new ID is stored in the set containing the existing ID and becomes the
new default for that set.

This command is limited to Namespaces that support Worker versioning.

#### Options

* `--build-id` (string) -
  Build ID to be added.
  Required
* `--task-queue`, `-t` (string) -
  Task Queue name.
  Required.
* `--existing-compatible-build-id` (string) -
  Pre-existing Build ID in this Task Queue.
  Required.
* `--set-as-default` (bool) -
  Set the expanded Build ID set as the Task Queue default.
  Defaults to false.

### temporal task-queue update-build-ids add-new-default: Set new default Build ID set

Create a new [Task Queue](/workers#task-queue) Build ID set, add a Build ID to it, and make it the
overall Task Queue default. The new set will be incompatible with previous 
sets and versions.

```
temporal task-queue update-build-ids add-new-default \
    --task-queue YourTaskQueue \
    --build-id YourNewBuildId
```

This command is limited to Namespaces that support Worker versioning.

#### Options

* `--build-id` (string) -
  Build ID to be added.
  Required
* `--task-queue`, `-t` (string) -
  Task Queue name.
  Required.

### temporal task-queue update-build-ids promote-id-in-set: Set Build ID as set default

Establish an existing Build ID as the default in its [Task Queue](/workers#task-queue) set. New
tasks compatible with this set will now be dispatched to this ID:

```
temporal task-queue update-build-ids promote-id-in-set \
    --task-queue YourTaskQueue \
    --build-id YourBuildId
```

This command is limited to Namespaces that support Worker versioning.

#### Options

* `--build-id` (string) -
  Build ID to set as default.
  Required
* `--task-queue`, `-t` (string) -
  Task Queue name.
  Required.

### temporal task-queue update-build-ids promote-set: Promote Build ID set

Promote a Build ID set to be the default on a [Task Queue](/workers#task-queue). Identify the
set by providing a Build ID within it. If the set is already the default,
this command has no effect:

```
temporal task-queue update-build-ids promote-set \
    --task-queue YourTaskQueue \
    --build-id YourBuildId
```

This command is limited to Namespaces that support Worker versioning.

#### Options

* `--build-id` (string) -
  Build ID within the promoted set.
  Required
* `--task-queue`, `-t` (string) -
  Task Queue name.
  Required.

### temporal workflow: Start, list, and operate on Workflows

[Workflow](/workflows) commands perform operations on [Workflow Executions](/workflows#workflow-execution):

```
temporal workflow [command] [options]
```

For example:

```
temporal workflow list
```

#### Options set for client:

* `--address` (string) -
  Temporal Service gRPC endpoint.
  Default: 127.0.0.1:7233.
  Env: TEMPORAL_ADDRESS.
* `--namespace`, `-n` (string) -
  Temporal Service Namespace.
  Default: default.
  Env: TEMPORAL_NAMESPACE.
* `--api-key` (string) -
  API key for request.
  Env: TEMPORAL_API_KEY
* `--grpc-meta` (string[]) -
  HTTP headers for requests.
  Format as comma-separated "KEY=VALUE" pairs.
* `--tls` (bool) -
  Enable base TLS encryption.
  Does not have additional options like mTLS or client certs. 
  Env: TEMPORAL_TLS.
* `--tls-cert-path` (string) -
  Path to x509 certificate.
  Can't be used with --tls-cert-data.
  Env: TEMPORAL_TLS_CERT.
* `--tls-cert-data` (string) -
  Data for x509 certificate.
  Can't be used with --tls-cert-path.
  Env: TEMPORAL_TLS_CERT_DATA.
* `--tls-key-path` (string) -
  Path to x509 private key.
  Can't be used with --tls-key-data.
  Env: TEMPORAL_TLS_KEY.
* `--tls-key-data` (string) -
  Private certificate key data.
  Can't be used with --tls-key-path.
  Env: TEMPORAL_TLS_KEY_DATA.
* `--tls-ca-path` (string) -
  Path to server CA certificate.
  Can't be used with --tls-ca-data.
  Env: TEMPORAL_TLS_CA.
* `--tls-ca-data` (string) -
  Data for server CA certificate.
  Can't be used with --tls-ca-path.
  Env: TEMPORAL_TLS_CA_DATA.
* `--tls-disable-host-verification` (bool) -
  Disable TLS host-name verification.
  Env: TEMPORAL_TLS_DISABLE_HOST_VERIFICATION.
* `--tls-server-name` (string) -
  Override target TLS server name.
  Env: TEMPORAL_TLS_SERVER_NAME.
* `--codec-endpoint` (string) -
  Remote Codec Server endpoint.
  Env: TEMPORAL_CODEC_ENDPOINT.
* `--codec-auth` (string) -
  Authorization header for Codec Server requests.
  Env: TEMPORAL_CODEC_AUTH.

### temporal workflow cancel: Send cancellation to Workflow Execution

Canceling a running [Workflow Execution](/workflows#workflow-execution) records a 
`WorkflowExecutionCancelRequested` event in the Event History. The [Service](/clusters)
schedules a new Command Task, and the Workflow Execution performs any cleanup
work supported by its implementation.

Use the [Workflow ID](/workflows#workflow-id) to cancel an Execution:

```
temporal workflow cancel \
    --workflow-id YourWorkflowId
```

A visibility query lets you send bulk cancellations to Workflow Executions
matching the results:

```
temporal workflow cancel \
    --query YourQuery
```

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch).

### temporal workflow count: Number of Workflow Executions

Show a count of [Workflow Executions](/workflows#workflow-execution), regardless of execution
state (running, terminated, etc). Use `--query` to select a subset of Workflow
Executions:

```
temporal workflow count \
    --query YourQuery
```

#### Options

* `--query`, `-q` (string) -
  Content for an SQL-like `QUERY` List Filter.

### temporal workflow delete: Remove Workflow Execution

Delete a [Workflow Executions](/workflows#workflow-execution) and its [Event History](/workflows#event-history):

```
temporal workflow delete \
    --workflow-id YourWorkflowId
```

The removal executes asynchronously. If the Execution is Running, the Service
terminates it before deletion.

#### Options

Includes options set for [single workflow or batch](#options-set-single-workflow-or-batch).

### temporal workflow describe: Show Workflow Execution info

Display information about a specific [Workflow Execution](/workflows#workflow-execution):

```
temporal workflow describe \
    --workflow-id YourWorkflowId
```

Show the Workflow Execution's auto-reset points:

```
temporal workflow describe \
    --workflow-id YourWorkflowId \
    --reset-points true
```

#### Options set for workflow reference:

* `--workflow-id`, `-w` (string) -
  Workflow ID.
  Required.
* `--run-id`, `-r` (string) -
  Run ID.

#### Options

* `--reset-points` (bool) -
  Show auto-reset points only.
* `--raw` (bool) -
  Print properties without changing their format.
  Defaults to true.

### temporal workflow execute: Start new Workflow Execution

Establish a new [Workflow Execution](/workflows#workflow-execution) and direct its
progress to stdout. The command blocks and returns when the Workflow Execution
completes. If your Workflow requires input, pass valid JSON:

```
temporal workflow execute
		--workflow-id YourWorkflowId \
		--type YourWorkflow \
		--task-queue YourTaskQueue \
		--input '{"Input": "As-JSON"}'
```

Use `--event-details` to relay updates to the command-line output in JSON
format. When using JSON output (`--output json`), this includes the entire
"history" JSON key for the run.

#### Options

* `--event-details` (bool) -
  Show event details during run.

Includes options set for [shared workflow start](#options-set-for-shared-workflow-start).
Includes options set for [workflow start](#options-set-for-workflow-start).
Includes options set for [payload input](#options-set-for-payload-input).

### temporal workflow fix-history-json: Updates an event history JSON file

Reserialize an [Event History](/workflows#event-history) JSON file:

```
temporal workflow fix-history-json \
	--source /path/to/original.json \
	--target /path/to/reserialized.json
```

#### Options

* `--source`, `-s` (string) -
  Path to the original file.
  Required.
* `--target`, `-t` (string) -
  Path to the results file.
  When omitted, output is sent to stdout.

### temporal workflow list: Show Workflow Executions

List [Workflow Executions](/workflows#workflow-execution). By default, this command returns
up to 10 closed Workflow Executions. The optional `--query` limits the output
to Workflows matching a [Query](/encyclopedia/application-message-passing#queries):

```
temporal workflow list \
    --query YourQuery`
```

View a list of archived Workflow Executions:

```
temporal workflow list \
    --archived
```

#### Options

* `--query`, `-q` (string) -
  Content for an SQL-like `QUERY` List Filter.
* `--archived` (bool) -
  Limit output to archived Workflow Executions.
* `--limit` (int) -
  Maximum number of Workflow Executions to display.

### temporal workflow query: Retrieve Workflow Execution state

Send a [Query](/encyclopedia/application-message-passing#queries) to a [Workflow Execution](/workflows#workflow-execution) by
[Workflow ID](/workflows#workflow-id) to retrieve its state. This synchronous
operation exposes the internal state of a running Workflow Execution, which
constantly changes. You can query both running and completed Workflow
Executions:

```
temporal workflow query \
    --workflow-id YourWorkflowId
    --type YourQueryType
    --input '{"YourInputKey": "YourInputValue"}'
```

Query implementations must never mutate Workflow Execution state and must
not contain blocking code.

#### Options

* `--type` (string) -
  Query Type/Name.
  Required.
* `--reject-condition` (string-enum) -
  Optional flag to reject Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [payload input](#options-set-for-payload-input).
Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow reset: Move Workflow Execution history point

Reset a [Workflow Execution](/workflows#workflow-execution) so it can resume from a point in
its [Event History](/workflows#event-history) without losing its progress up to that point:

```
temporal workflow reset \
    --workflow-id YourWorkflowId
    --event-id YourLastEvent
```

Start from where the Workflow Execution last continued as new:

```
temporal workflow reset \
    --workflow-id YourWorkflowId \
    --type LastContinuedAsNew
```

For batch resets, limit your resets to FirstWorkflowTask, LastWorkflowTask,
or BuildId. Do not use Workflow IDs, run IDs, or event IDs with this
command.

#### Options

* `--workflow-id`, `-w` (string) -
  Workflow ID.
  Required for non-batch reset operations.
* `--run-id`, `-r` (string) -
  Run ID.
* `--event-id`, `-e` (int) -
  Event ID to reset to. 
  Event must occur after `WorkflowTaskStarted`.
  `WorkflowTaskCompleted`, `WorkflowTaskFailed`, etc are valid.
* `--reason` (string) -
  Reason for reset.
  Required.
* `--reapply-type` (string-enum) -
  Types of events to re-apply after reset point.
  Options: All, Signal, None.
  Default: All.
* `--type`, `-t` (string-enum) -
  The event type for the reset. 
  Options: FirstWorkflowTask, LastWorkflowTask, LastContinuedAsNew, BuildId.
* `--build-id` (string) -
  A Build ID.
  Use only with the BuildId `--type`. 
  Resets the first Workflow task processed by this ID.
  By default, this reset may be in a prior run, earlier than a Continue
  as New point.
* `--query`, `-q` (string) -
  Content for an SQL-like `QUERY` List Filter.
* `--yes`, `-y` (bool) -
  Don't prompt to confirm.
  Only allowed when `--query` is present.

### temporal workflow show: Display Event History

Show a [Workflow Execution](/workflows#workflow-execution)'s [Event History](/workflows#event-history). 
When using JSON output (`--output JSON`), you may pass the results to an SDK
to perform a replay:

```
temporal workflow show \
    --workflow-id YourWorkflowId
    --output json
```

#### Options

* `--event-details` (bool) -
  Show event details during run.
* `--follow`, `-f` (bool) -
  Direct Workflow Execution progress to stdout in real time.
  Does not apply when JSON output is selected.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow signal: Send a message to a Workflow Execution

Send an asynchronous notification (Signal) to a running [Workflow Execution](/workflows#workflow-execution) 
by its [Workflow ID](/workflows#workflow-id). The Signal is written to the
History. When you include `--input`, that data is available for the Workflow
Execution to consume:

```
temporal workflow signal \
		--workflow-id YourWorkflowId \
		--name YourSignal \
		--input '{"YourInputKey": "YourInputValue"}'
```

#### Options

* `--name` (string) -
  Signal name.
  Required.

Includes options set for [payload input](#options-set-for-payload-input).

#### Options set for single workflow or batch:
* `--workflow-id`, `-w` (string) -
  Workflow ID.
  You must set either --workflow-id or --query.
* `--query`, `-q` (string) -
  Content for an SQL-like `QUERY` List Filter.
  You must set either --workflow-id or --query.
* `--run-id`, `-r` (string) -
  Run ID.
  Only use with --workflow-id.
  Cannot use with --query.
* `--reason` (string) -
  Reason to perform batch.
  Only use with --query.
  Defaults to user name.
* `--yes`, `-y` (bool) -
  Don't prompt to confirm signaling.
  Only allowed when --query is present.

### temporal workflow stack: Trace a Workflow Execution

Perform a [Query](/encyclopedia/application-message-passing#queries) on a [Workflow Execution](/workflows#workflow-execution) using a `__stack_trace`-type
Query. Display a stack trace of the threads and routines currently in use
by the Workflow for troubleshooting:

```
temporal workflow stack \
    --workflow-id YourWorkflowId
```

#### Options

* `--reject-condition` (string-enum) -
  Optional flag to reject Queries based on Workflow state.
  Options: not_open, not_completed_cleanly.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow start: Initiate a Workflow Execution

Start a new [Workflow Execution](/workflows#workflow-execution). Returns the Workflow-
and Run-IDs.

```
temporal workflow start \
		--workflow-id YourWorkflowId \
		--type YourWorkflow \
		--task-queue YourTaskQueue \
		--input '{"Input": "As-JSON"}'
```

#### Options set for shared workflow start:

* `--workflow-id`, `-w` (string) -
  Workflow ID.
  If not supply, the Service generates a unique ID.
* `--type` (string) -
  Workflow Type name.
  Required.
* `--task-queue`, `-t` (string) -
  Workflow Task queue.
  Required.
* `--run-timeout` (duration) -
  Fail a Workflow Run if it takes longer than `DURATION`.
* `--execution-timeout` (duration) -
  Fail a WorkflowExecution if it takes longer than `DURATION`.
  This time-out includes retries and ContinueAsNew tasks.
* `--task-timeout` (duration) -
  Fail a Workflow Task if it takes longer than `DURATION`.
  This is the Start-to-close timeout for a Workflow Task.
  Default: 10s.
* `--search-attribute` (string[]) -
  Search Attribute in "KEY='VALUE'" format.
  KEY is a string, VALUE is JSON.
* `--memo` (string[]) -
  Memo in 'KEY="VALUE"' format.
  KEY is a string, VALUE is JSON. 

#### Options set for workflow start:

* `--cron` (string) -
  Cron schedule for the Workflow.
  Deprecated. 
  Use Schedules instead.
* `--fail-existing` (bool) -
  Fail if the Workflow already exists.
* `--start-delay` (duration) -
  Delay before starting the Workflow Execution.
  Can't be used with cron schedules.
  If the Workflow receives a signal or update prior to this time, the Workflow
  Execution starts immediately.
* `--id-reuse-policy` (string) -
  Re-use policy for the Workflow ID in new Workflow Executions.
  Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning.

#### Options set for payload input:

* `--input`, `-i` (string[]) -
  Input value.
  Use JSON content or set --input-meta to override.
  Can't be combined with --input-file.
  Can be passed multiple times. 
* `--input-file` (string[]) -
  A path or paths for input file(s).
  Use JSON content or set --input-meta to override.
  Can't be combined with --input.
  Can be passed multiple times to concatenate the file contents.
* `--input-meta` (string[]) -
  Input payload metadata as a "KEY=VALUE" pair.
  When the KEY is "encoding", this overrides the default ("json/plain").
  Can be passed multiple times.
* `--input-base64` (bool) -
  Assume inputs are base64-encoded and attempt to decode them.

### temporal workflow terminate: Forcefully end a Workflow Execution

Terminate a [Workflow Execution](/workflows#workflow-execution):

```
temporal workflow terminate \
    --reason YourReasonForTermination \
    --workflow-id YourWorkflowId
```

The reason is optional and defaults to the current user's name. The reason
is stored in the Event History as part of the `WorkflowExecutionTerminated`
event. This becomes the closing Event in the Workflow Execution's history.

Executions may be terminated in bulk via a visibility query [list filter](/visibility#list-filter):

```
temporal workflow terminate \
    --query YourQuery \
    --reason YourReasonForTermination
```

Workflow code cannot see or respond to terminations. To perform clean-up work
in your Workflow code, use `temporal workflow cancel` instead.

#### Options

* `--workflow-id`, `-w` (string) -
  Workflow ID.
  You must set either --workflow-id or --query.
* `--query`, `-q` (string) -
  Content for an SQL-like `QUERY` List Filter.
  You must set either --workflow-id or --query.
* `--run-id`, `-r` (string) -
  Run ID.
  Can only be set with --workflow-id.
  Do not use with --query.
* `--reason` (string) -
  Reason for termination.
  Defaults to message with the current user's name.
* `--yes`, `-y` (bool) -
  Don't prompt to confirm termination.
  Can only be used with --query.

### temporal workflow trace: Workflow Execution live progress

Display the progress of a [Workflow Execution](/workflows#workflow-execution) and its child
workflows with a real-time trace. This view helps you understand how Workflows
are proceeding:

```
temporal workflow trace \
    --workflow-id YourWorkflowId
```

#### Options

* `--fold` (string[]) -
  Status for folding away Child Workflows.
  Case-insensitive.
  Ignored if --no-fold supplied.
  Available values: running, completed, failed, canceled, terminated, timedout, continueasnew.
  Can be passed multiple times.
  Each fold reduces the amount of information fetched and displayed.
* `--no-fold` (bool) -
  Disable folding.
  Fetch and display Child Workflows within set depth.
* `--depth` (int) -
  Set depth for Child Workflow fetches.
  Pass -1 to fetch child workflows at any depth.
  Default: -1.
* `--concurrency` (int) -
  Number of Workflow Histories to fetch at a time.
  Default: 10.

Includes options set for [workflow reference](#options-set-for-workflow-reference).

### temporal workflow update: Synchronously run a Workflow update handler

Send a message to a [Workflow Execution](/workflows#workflow-execution) to invoke an
[update](/encyclopedia/application-message-passing#updates) handler. An update can change the
state of a Workflow Execution and return a response:

```
temporal workflow update \
		--workflow-id YourWorkflowId \
		--name YourUpdate \
		--input '{"Input": "As-JSON"}'
```

#### Options

* `--name` (string) -
  Handler method name.
  Required.
* `--workflow-id`, `-w` (string) -
  Workflow ID.
  Required.
* `--update-id` (string) -
  Update ID.
  If unset, defaults to a UUID.
  Must be unique per Workflow Execution.
* `--run-id`, `-r` (string) -
  Run ID.
  If unset, updates the currently-running Workflow Execution.
* `--first-execution-run-id` (string) -
  Parent Run ID.
  The update is sent to the last Workflow Execution in the chain started
  with this Run ID.

Includes options set for [payload input](#options-set-for-payload-input).
