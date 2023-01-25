package common

const (
	// Main command definitions
	WorkflowDefinition = "Operations that can be performed on Workflows."
	ActivityDefinition = "Operations that can be performed on Workflow Activities."
	TaskQueueDefinition = "Operations performed on Task Queues."
	ScheduleDefinition = "Operations performed on Schedules."
	BatchDefinition = "Operations performed on Batch jobs. Use Workflow commands with --query flag to start batch jobs."
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

const BatchUsageText = `Batch commands allow you to change multiple [Workflow Executions](/workflows#workflow-execution) without having to repeat yourself on the command line. 
In order to do this, you provide the command with a [List Filter](/visibility#list-filter) and the type of Batch job to execute.

The List Filter identifies the Workflow Executions that will be affected by the Batch job.
The Batch type determines the other parameters that need to be provided, along with what is being affected on the Workflow Executions.

To start the Batch job, run `+"`"+`temporal workflow query`+"`"+`.
A successfully started Batch job will return a Job ID.
Use this Job ID to execute other actions on the Batch job.
`
//TODO: get above info in BatchUsageText checked.

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
