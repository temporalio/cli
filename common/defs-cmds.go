package common



const (
	// Main command definitions
	WorkflowDefinition = "Operations that can be performed on Workflows."
	ActivityDefinition = "Operations that can be performed on Workflow Activities."
	TaskQueueDefinition = "Operations performed on Task Queues."
	ScheduleDefinition = "Operations performed on Schedules."
	BatchDefinition = "Operations performed on Batch jobs."
	OperatorDefinition = "Operations performed on the Temporal Server."
	EnvDefinition = "Manage environmental configurations on Temporal Client."

	// Workflow subcommand definitions
	StartWorkflowDefinition = "Starts a new Workflow Execution."
	ExecuteWorkflowDefinition = "Start a new Workflow Execution and prints its progress."
	DescribeWorkflowDefinition = "Show information about a Workflow Execution."
	ListWorkflowDefinition = "List Workflow Executions based on a Query."
	ShowWorkflowDefinition = "Show Event History for a Workflow Execution."
	QueryWorkflowDefinition = "Query a Workflow Execution."
	StackWorkflowDefinition = "Query a Workflow Execution with __stack_trace as the query type."
	SignalWorkflowDefinition = "Signal Workflow Execution by Id or List Filter."
	CountWorkflowDefinition = "Count Workflow Executions (requires ElasticSearch to be enabled)."
	CancelWorkflowDefinition = "Cancel a Workflow Execution."
	TerminateWorkflowDefinition = "Terminate Workflow Execution by Id or List Filter."
	DeleteWorkflowDefinition = "Deletes a Workflow Execution."
	ResetWorkflowDefinition = "Resets a Workflow Execution by Event Id or reset type."
	TraceWorkflowDefinition = "Trace progress of a Workflow Execution and its children."

	// Activity subcommand definitions
	CompleteActivityDefinition = "Completes an Activity."
	FailActivityDefinition = "Fails an Activity."

	// Task Queue subcommand definitions
	DescribeTaskQueueDefinition = "Describes the Workers that have recently polled on this Task Queue."
	ListPartitionTaskQueueDefinition = "Lists the Task Queue's partitions and which matching node they are assigned to."

	// Batch subcommand definitions
	DescribeBatchJobDefinition = "Describe a Batch operation job." 
	ListBatchJobsDefinition = "List Batch operation jobs."
	TerminateBatchJobDefinition = "Stop a Batch operation job."

	NamespaceDefinition = "Operations applying to Namespaces."
	SearchAttributeDefinition = "Operations applying to Search Attributes."
	ClusterDefinition = "Operations for running a Temporal Cluster."

	// Namespace subcommand definitions
	DescribeNamespaceDefinition =  "Describe a Namespace by its name or Id."
	ListNamespacesDefinition = "List all Namespaces."
	RegisterNamespaceDefinition = "Registers a new Namespace."
	UpdateNamespaceDefinition = "Updates a Namespace."
	DeleteNamespaceDefinition = "Deletes an existing Namespace."

	// Search Attribute subcommand defintions
	CreateSearchAttributeDefinition = "Adds one or more custom Search Attributes."
	ListSearchAttributesDefinition = "Lists all Search Attributes that can be used in list Workflow Queries."
	RemoveSearchAttributesDefinition = "Removes custom search attribute metadata only (Elasticsearch index schema is not modified)."

	// Cluster subcommand defintions
	HealthDefinition = "Checks the health of the Frontend Service."
	DescribeDefinition = "Show information about the Cluster."
	SystemDefinition = "Shows information about the system and its capabilities."
	UpsertDefinition = "Add or update a remote Cluster."
	ListDefinition = "List all remote Clusters."
	RemoveDefinition = "Remove a remote Cluster."

	// Env subcommand definitions
	GetDefinition = "Prints environmental properties."
	SetDefinition = "Set environmental properties."
	DeleteDefinition = "Delete an environment or environmental property."
)

//TODO: get information checked for all UsageTexts

const BatchUsageText = `Batch commands allow you to change multiple [Workflow Executions](/workflows#workflow-execution) without having to repeat yourself on the command line. 
In order to do this, you provide the command with a [List Filter](/visibility#list-filter) and the type of Batch job to execute.

The List Filter identifies the Workflow Executions that will be affected by the Batch job.
The Batch type determines the other parameters that need to be provided, along with what is being affected on the Workflow Executions.

To start the Batch job, run `+"`"+`temporal workflow query`+"`"+`.
A successfully started Batch job will return a Job ID.
Use this Job ID to execute other actions on the Batch job.
`
const DescribeBatchUsageText = `The `+"`"+`temporal batch describe`+"`"+` command shows the progress of an ongoing Batch job.

Use the command options listed below to change the information returned by this command.
Make sure to write the command in this format:
`+"`"+`temporal batch describe [command options] [arguments]`+"`"+`

## OPTIONS
`
const ListBatchUsageText = `When used, `+"`"+`temporal batch list`+"`"+` returns all Batch jobs. 

Use the command options listed below to change the information returned by this command.
Make sure to write the command in this format:
`+"`"+`temporal batch list [command options] [arguments]`+"`"+`

## OPTIONS
`
const TerminateBatchUsageText = `The `+"`"+`temporal batch terminate`+"`"+` command terminates a Batch job with the provided Job ID. 

Use the command options listed below to change the behavior of this command.
Make sure to write the command as follows:
`+"`"+`temporal batch terminate [command options] [arguments]`+"`"+`

## OPTIONS
`
const StartWorkflowUsageText = `The `+"`"+`temporal workflow start`+"`"+` command starts a new [Workflow Execution](/workflows#workflow-execution).
When invoked successfully, the Workflow and Run ID are returned immediately after starting the [Workflow](/workflows).

Use the command options listed below to change how the Workflow Execution behaves upon starting.
Make sure to write the command in this format:
`+"`"+`temporal workflow start [command options] [arguments]`+"`"+`

## OPTIONS
`
const ExecuteWorkflowUsageText = `The `+"`"+`temporal workflow execute`+"`"+` command starts a new [Workflow Execution](/workflows#workflow-execution) and prints its progress.
The command doesn't finish until the [Workflow](/workflows) completes.

Single quotes('') are used to wrap input as JSON.

Use the command options listed below to change how the Workflow Execution behaves during its run.
Make sure to write the command in this format:
`+"`"+`temporal workflow execute [command options] [arguments]`+"`"+`

## OPTIONS
`
const DescribeWorkflowUsageText = `The `+"`"+`temporal workflow describe`+"`"+` command shows information about a given [Workflow Execution](/workflows#workflow-execution).
This information can be used to locate Workflow Executions that weren't able to run successfully.

Use the command options listed below to change the information returned by this command.
Make sure to write the command in this format:
`+"`"+`temporal workflow describe [command options] [arguments]`+"`"+`

## OPTIONS
`
const ListWorkflowUsageText = `The `+"`"+`temporal workflow list`+"`"+` command provides a list of [Workflow Executions](/workflows#workflow-execution) that meet the criteria of a given [Query](/workflows#query).
By default, this command returns a list of up to 10 closed Workflow Executions.

Use the command options listed below to change the information returned by this command.
Make sure to write the command as follows:
`+"`"+`temporal workflow list [command options] [arguments]`+"`"+`

## OPTIONS
`
const QueryWorkflowUsageText = `The `+"`"+`temporal workflow query`+"`"+` command sends a [Query](/workflows#query) to a [Workflow Execution](/workflows#workflow-execution).

Queries can retrieve all or part of the Workflow state within given parameters.
Queries can also be used on completed [Workflows](/workflows#workflow-execution).

Use the command options listed below to change the information returned by this command.
Make sure to write the command as follows:
`+"`"+`temporal workflow query [command options] [arguments]`+"`"+`

## OPTIONS
`
const CancelWorkflowUsageText = `The `+"`"+`temporal workflow cancel`+"`"+` command cancels a [Workflow Execution](/workflows#workflow-execution).

Canceling a running Workflow Execution records a [`+"`"+`WorkflowExecutionCancelRequested`+"`"+` event](/events#workflowexecutioncancelrequested) in the [Event History](/workflows#event-history).
A new [Command](/workflows#command) Task will be scheduled, and the Workflow Execution performs cleanup work.

Use the options listed below to change the behavior of this command.
Make sure to write the command as follows:
`+"`"+`temporal workflow cancel [command options] [arguments]`+"`"+`

## OPTIONS
`
const TerminateWorkflowUsageText = `The `+"`"+`temporal workflow terminate`+"`"+` command terminates a [Workflow Execution](/workflows#workflow-execution)

Terminating a running Workflow Execution records a [`+"`"+`WorkflowExecutionTerminated`+"`"+` event](/events#workflowexecutionterminated) as the closing Event in the [Event History](/workflows#event-history).
Any further [Command](/workflows#command) Tasks cannot be scheduled after running this command.

Use the options listed below to change termination behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow terminate [command options] [arguments]`+"`"+`

## OPTIONS
`
const ResetWorkflowUsageText = `The `+"`"+`temporal workflow reset`+"`"+` command resets a [Workflow Execution](/workflows#workflow-execution).
A reset allows the Workflow to be resumed from a certain point without losing your parameters or [Event History](/workflows#event-history).

Use the options listed below to change reset behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow reset [command options] [arguments]`+"`"+`

## OPTIONS
`
const ResetBatchUsageText = `The `+"`"+`temporal workflow reset-batch`+"`"+` command resets a batch of [Workflow Executions](/workflows#workflow-execution) by `+"`"+`resetType`+"`"+`.
Resetting a [Workflow](/workflows) allows the process to resume from a certain point without losing your parameters or [Event History](/workflows#event-history).

Use the options listed below to change reset behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow reset-batch [command options] [arguments]`+"`"+`

## OPTIONS
`
const DescribeTaskQueueUsageText = `The `+"`"+`temporal task-queue describe`+"`"+` command provides [poller](/applcation-development/worker-performance#poller-count) information for a given [Task Queue](/tasks#task-queue).

The [Server](/clusters#temporal-server) records the last time of each poll request.
Should `+"`"+`LastAccessTime`+"`"+` exceeds one minute, it's likely that the Worker is at capacity (all Workflow and Activity slots are full) or that the Worker has shut down.
[Workers](/workers) are removed if 5 minutes have passed since the last poll request.

Use the options listed below to modify what this command returns.
Make sure to write the command as follows:
`+"`"+`temporal task-queue describe [command options] [arguments]`+"`"+`

## OPTIONS
`
const ScheduleUsageText = `Schedule commands allow the user to create, use, and update [Schedules](/workflows#schedule).
Schedules control when certain Actions for a Workflow Execution are performed, making it a useful tool for automation.

To run a Schedule command, run `+"`"+`temporal schedule [command] [command options] [arguments]`+"`"+`.
`
const OperatorUsageText = `Operator commands enable actions on [Namespaces](/namespaces), [Search Attributes](/visibility#search-attribute), and [Temporal Clusters](/clusters).
These actions are performed through subcommands for each Operator area.

To run an Operator command, run `+"`"+`temporal operator [command] [subcommand] [command options] [arguments]`+"`"+`.
`
const CompleteActivityUsageText = `The `+"`"+`temporal activity complete`+"`"+` command completes an [Activity Execution](/activities#activity-execution).

Use the options listed below to change the behavior of this command.
Make sure to write the command as follows:
`+"`"+`temporal activity complete [command options] [arguments]`+"`"+`

## OPTIONS
`
const FailActivityUsageText = `The `+"`"+`temporal activity fail`+"`"+` command fails an [Activity Execution](/activities#activity-execution).

Use the options listed below to change the behavior of this command.
Make sure to write the command as follows:
`+"`"+`temporal activity fail [command options] [arguments]`+"`"+`

## OPTIONS
`
const HealthUsageText = `The `+"`"+`temporal operator cluster health`+"`"+` command checks the health of the [Frontend Service](/clusters#frontend-service).

Use the options listed below to change the behavior and output of this command.
Make sure to write the command as follows:
`+"`"+`temporal operator cluster health [command options] [arguments]`+"`"+`

## OPTIONS
`
const ClusterDescribeUsageText = `The `+"`"+`temporal operator cluster describe`+"`"+` command shows information about the [Cluster](/clusters).

Use the options listed below to change the output of this command.
Make sure to write the command as follows:
`+"`"+`temporal operator cluster describe [command options] [arguments]`+"`"+`

## OPTIONS
`
const ClusterSystemUsageText = `The `+"`"+`temporal operator cluster system`+"`"+` command provides information about the system the Cluster is running on.

Use the options listed below to change this command's output.
Make sure to write the command as follows:
`+"`"+`temporal operator cluster system [command options] [arguments]`+"`"+`

## OPTIONS
`
const ClusterUpsertUsageText = `The `+"`"+`temporal operator cluster upsert`+"`"+` command allows the user to add or update a remote [Cluster](/clusters).

Use the options listed below to change the behavior of this command.
Make sure to write the command as follows:
`+"`"+`temporal operator cluster upsert [command options] [arguments]`+"`"+`

## OPTIONS
`

const ClusterListUsageText = `The `+"`"+`temporal operator cluster list`+"`"+` command prints a list of all remote [Clusters](/clusters) on the system.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator cluster list [command options] [arguments]`+"`"+`

## OPTIONS
`
const ClusterRemoveUsageText = `The `+"`"+`temporal operator cluster remove`+"`"+` command removes a remote [Cluster](/clusters) from the system.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator cluster remove [command options] [arguments]`+"`"+`

## OPTIONS
`
const EnvGetUsageText = `The `+"`"+`temporal env get`+"`"+` command prints the environmental properties for the environment in use.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal env get [command options] [arguments]`+"`"+`

## OPTIONS
`
const EnvSetUsageText = `The `+"`"+`temporal env set`+"`"+` command sets the value for an environmental property.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal env set [command options] [arguments]`+"`"+`

## OPTIONS
`
const EnvDeleteUsageText = `The `+"`"+`temporal env delete`+"`"+` command deletes a given environment or environmental property.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal env delete [command options] [arguments]`+"`"+`

## OPTIONS
`
const NamespaceDescribeUsageText = `The `+"`"+`temporal operator namespace describe`+"`"+` command provides a description of a [Namespace](/namespaces).
Namespaces can be identified by name or Namespace ID.

Use the options listed below to change the command's output.
Make sure to write the command as follows:
`+"`"+`temporal operator namespace describe [command options] [arguments]`+"`"+`

## OPTIONS
`
const NamespaceListUsageText = `The `+"`"+`temporal operator namespace list`+"`"+` command lists all [Namespaces](/namespaces) on the [Server](/clusters#frontend-server).

Use the options listed below to change the command's output.
Make sure to write the command as follows:
`+"`"+`temporal operator namespace list [command options] [arguments]`+"`"+`

## OPTIONS
`
const NamespaceCreateUsageText = `The `+"`"+`temporal operator namespace create`+"`"+` command creates a new [Namespace](/namespaces).

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator namespace create [command options] [arguments]`+"`"+`

## OPTIONS
`
const NamespaceUpdateUsageText = `The `+"`"+`temporal operator namespace update`+"`"+` command updates a given [Namespace](/namespaces).

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator namespace update [command options] [arguments]`+"`"+`

## OPTIONS
`
const NamespaceDeleteUsageText = `The `+"`"+`temporal operator namespace delete`+"`"+` command deletes a given [Namespace](/namespaces) from the system.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator namespace delete [command options] [arguments]`+"`"+`

## OPTIONS
`
const ScheduleCreateUsageText = `The `+"`"+`temporal schedule create`+"`"+` command creates a new [Schedule](/workflows#schedule).
Newly created Schedules return a Schedule ID to be used in other Schedule commands.

Schedules need to follow a format like the example shown here:
`+"`"+``+"`"+``+"`"+`
temporal schedule create \
		--sid 'your-schedule-id' \
		--cron '3 11 * * Fri' \
		--wid 'your-workflow-id' \
		--tq 'your-task-queue' \
		--type 'YourWorkflowType' 
`+"`"+``+"`"+``+"`"+`

Any combination of `+"`"+`--cal`+"`"+`, `+"`"+`--interval`+"`"+`, and `+"`"+`--cron`+"`"+` is supported.
Actions will be executed at any time specified in the Schedule.

Use the options provided below to change the command's behavior.

## OPTIONS
`
const ScheduleUpdateUsageText = `The `+"`"+`temporal schedule update`+"`"+` command updates an existing [Schedule](/workflows#schedule).

Like `+"`"+`temporal schedule create`+"`"+`, updated Schedules need to follow a certain format:
`+"`"+``+"`"+``+"`"+`
temporal schedule update 			\
		--sid 'your-schedule-id' 	\
		--cron '3 11 * * Fri' 		\
		--wid 'your-workflow-id' 	\
		--tq 'your-task-queue' 		\
		--type 'YourWorkflowType' 
`+"`"+``+"`"+``+"`"+`

Updating a Schedule takes the given options and replaces the entire configuration of the Schedule with what's provided. 
If you only change one value of the Schedule, be sure to provide the other unchanged fields to prevent them from being overwritten.

Use the options provided below to change the command's behavior.

## OPTIONS
`

const ScheduleToggleUsageText = `The `+"`"+`temporal schedule toggle`+"`"+` command can pause and unpause a [Schedule](/workflows#schedule).

Toggling a Schedule requires a reason to be entered on the command line. 
Use `+"`"+`--reason`+"`"+` to note the issue leading to the pause or unpause.

Schedule toggles are passed in this format:
`+"`"+` temporal schedule toggle --sid 'your-schedule-id' --pause --reason "paused because the database is down"`+"`"+`
`+"`"+`temporal schedule toggle --sid 'your-schedule-id' --unpause --reason "the database is back up"`+"`"+`

Use the options provided below to change this command's behavior.

## OPTIONS
`

const ScheduleTriggerUsageText = `The `+"`"+`temporal schedule trigger`+"`"+` command triggers an immediate action with a given [Schedule](/workflows#schedule).
By default, this action is subject to the Overlap Policy of the Schedule.

`+"`"+`temporal schedule trigger`+"`"+` can be used to start a Workflow Run immediately.
`+"`"+`temporal schedule trigger --sid 'your-schedule-id'`+"`"+` 

The Overlap Policy of the Schedule can be overridden as well.
`+"`"+`temporal schedule trigger --sid 'your-schedule-id' --overlap-policy 'AllowAll'`+"`"+`

Use the options provided below to change this command's behavior.

## OPTIONS
`

const ScheduleBackfillUsageText = `The `+"`"+`temporal schedule backfill`+"`"+` command executes Actions ahead of their specified time range. 
Backfilling can be used to fill in [Workflow Runs](/workflows#run-id) from a time period when the Schedule was paused, or from before the Schedule was created. 

`+"`"+``+"`"+``+"`"+`
temporal schedule backfill --sid 'your-schedule-id' \
		--overlap-policy 'BufferAll' 				\
		--start-time '2022-05-0101T00:00:00Z'		\
		--end-time '2022-05-31T23:59:59Z'
`+"`"+``+"`"+``+"`"+`

Use the options provided below to change this command's behavior.

## OPTIONS
`
const ScheduleDescribeUsageText = `The `+"`"+`temporal schedule describe`+"`"+` command shows the current [Schedule](#workflows#schedule) configuration.
This command also provides information about past, current, and future [Workflow Runs](/workflows#run-id).

`+"`"+`temporal schedule describe --sid 'your-schedule-id' [command options] [arguments]`+"`"+`

Use the options below to change this command's output.

## OPTIONS
`
const ScheduleDeleteUsageText = `The `+"`"+`temporal schedule delete`+"`"+` command deletes a [Schedule](/workflows#schedule).
Deleting a Schedule does not affect any [Workflows](/workflows) started by the Schedule.

[Workflow Executions](/workflows#workflow-execution) started by Schedules can be cancelled or terminated like other Workflow Executions.
However, Workflow Executions started by a Schedule can be identified by their [Search Attributes](/visibility#search-attribute), making them targetable by batch command for termination.

`+"`"+`temporal schedule delete --sid 'your-schedule-id' [command options] [arguments]`+"`"+`

Use the options below to change the behavior of this command.

## OPTIONS
`
const ScheduleListUsageText = `The `+"`"+`temporal schedule list`+"`"+` command lists all [Schedule](/workflows#schedule) configurations.
Listing Schedules in [Standard Visibility](/visibility#standard-visibility) will only provide Schedule IDs.

`+"`"+`temporal schedule list [command options] [arguments]`+"`"+`

Use the options below to change the behavior of this command.

## OPTIONS
`
const SearchAttributeCreateUsageText = `The `+"`"+`temporal operator search-attribute create`+"`"+` command adds one or more custom [Search Attributes](/visibility#search-attribute).
These Search Attributes can be used to [filter a list](/visibility#list-filter) of [Workflow Executions](/workflows#workflow-execution) that contain the given Search Attributes in their metadata.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator search-attribute create [command options] [arguments]`+"`"+`

## OPTIONS
`
const SearchAttributeListUsageText = `The `+"`"+`temporal operator search-attrbute list`+"`"+` command displays a list of all [Search Attributes](/visibility#search-attribute) that can be used in `+"`"+` temporal workflow list --query`+"`"+`.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator search-attribute list [command options] [arguments]`+"`"+`

## OPTIONS
`
const SearchAttributeRemoveUsageText = `The `+"`"+`temporal operator search-attribute remove`+"`"+` command removes custom [Search Attribute](/visibility#search-attribute) metadata.
This command does not remove custom Search Attributes from Elasticsearch.
The index schema is not modified.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal operator search-attribute remove [command options] [arguments]`+"`"+`

## OPTIONS
`
const TaskQueueListPartitionUsageText = `The `+"`"+`temporal task-queue list-partition`+"`"+` command displays the partitions of a [Task Queue](/tasks#task-queue), along with the matching node they are assigned to.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal task-queue list-partition [command options] [arguments]`+"`"+`

## OPTIONS
`
const WorkflowShowUsageText = `The `+"`"+`temporal workflow show`+"`"+` command provides the [Event History](/workflows#event-history) for a specified [Workflow Execution](/workflows#workflow-execution).

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow show [command options] [arguments]`+"`"+`

## OPTIONS
`
const WorkflowStackUsageText = `The `+"`"+`temporal workflow stack`+"`"+` command queries a [Workflow Execution](/workflows#workflow-execution) with `+"`"+`--stack-trace`+"`"+` as the [Query](/workflows#stack-trace-query) type.
Returning the stack trace of all the threads owned by a Workflow Execution can be great for troubleshooting in production.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow stack [command options] [arguments]`+"`"+`

## OPTIONS
`
const WorkflowSignalUsageText = `The `+"`"+`temporal workflow signal`+"`"+` command is used to [Signal](/workflows#signal) a [Workflow Execution](/workflows#workflow-execution) by ID or [List Filter](/visibility#list-filter).

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow signal [command options] [arguments]`+"`"+`

## OPTIONS
`
const WorkflowCountUsageText = `The `+"`"+`temporal workflow count`+"`"+` command returns a count of [Workflow Executions](/workflows#workflow-execution).
This command requires Elasticsearch to be enabled.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow count [command options] [arguments]`+"`"+`

## OPTIONS
`
const WorkflowDeleteUsageText = `The `+"`"+`temporal workflow delete`+"`"+` command deletes the specified [Workflow Execution](/workflows#workflow-execution).

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow delete [command options] [arguments]`+"`"+`

## OPTIONS
`

const WorkflowTraceUsageText = `The `+"`"+`temporal workflow trace`+"`"+` command tracks the progress of a [Workflow Execution](/workflows#workflow-execution) and any  [Child Workflows](/workflows#child-workflow) it generates.

Use the options listed below to change the command's behavior.
Make sure to write the command as follows:
`+"`"+`temporal workflow trace [command options] [arguments]`+"`"+`

## OPTIONS
`