package common

const (
	// Main command definitions
	WorkflowDefinition  = "Operations performed on Workflows."
	ActivityDefinition  = "Operations performed on Workflow Activities."
	TaskQueueDefinition = "Operations performed on Task Queues."
	ScheduleDefinition  = "Operations performed on Schedules."
	BatchDefinition     = "Operations performed on Batch jobs."
	OperatorDefinition  = "Operations performed on the Temporal Server."
	EnvDefinition       = "Manage environmental configurations on Temporal Client."

	// Workflow subcommand definitions
	StartWorkflowDefinition     = "Starts a new Workflow Execution."
	ExecuteWorkflowDefinition   = "Start a new Workflow Execution and prints its progress."
	DescribeWorkflowDefinition  = "Show information about a Workflow Execution."
	ListWorkflowDefinition      = "List Workflow Executions based on a Query."
	ShowWorkflowDefinition      = "Show Event History for a Workflow Execution."
	QueryWorkflowDefinition     = "Query a Workflow Execution."
	StackWorkflowDefinition     = "Query a Workflow Execution with __stack_trace as the query type."
	SignalWorkflowDefinition    = "Signal Workflow Execution by Id or List Filter."
	CountWorkflowDefinition     = "Count Workflow Executions (requires ElasticSearch to be enabled)."
	CancelWorkflowDefinition    = "Cancel a Workflow Execution."
	TerminateWorkflowDefinition = "Terminate Workflow Execution by ID or List Filter."
	DeleteWorkflowDefinition    = "Deletes a Workflow Execution."
	ResetWorkflowDefinition     = "Resets a Workflow Execution by Event ID or reset type."
	TraceWorkflowDefinition     = "Trace progress of a Workflow Execution and its children."
	UpdateWorkflowDefinition    = "Updates a running workflow synchronously."

	// Activity subcommand definitions
	CompleteActivityDefinition = "Completes an Activity Execution."
	FailActivityDefinition     = "Fails an Activity Execution."

	// Task Queue subcommand definitions
	DescribeTaskQueueDefinition      = "Provides information for Workers that have recently polled on this Task Queue."
	ListPartitionTaskQueueDefinition = "Lists the Task Queue's partitions and the matching nodes they are assigned to."

	// Batch subcommand definitions
	DescribeBatchJobDefinition  = "Provide information about a Batch operation job."
	ListBatchJobsDefinition     = "List all Batch operation jobs on the Temporal Client."
	TerminateBatchJobDefinition = "Stop an ongoing Batch operation job."

	NamespaceDefinition       = "Operations performed on Namespaces."
	SearchAttributeDefinition = "Operations applying to Search Attributes."
	ClusterDefinition         = "Operations for running a Temporal Cluster."

	// Namespace subcommand definitions
	DescribeNamespaceDefinition = "Describe a Namespace by its name or ID."
	ListNamespacesDefinition    = "List all Namespaces."
	CreateNamespaceDefinition   = "Registers a new Namespace."
	UpdateNamespaceDefinition   = "Updates a Namespace."
	DeleteNamespaceDefinition   = "Deletes an existing Namespace."

	// Search Attribute subcommand defintions
	CreateSearchAttributeDefinition  = "Adds one or more custom Search Attributes."
	ListSearchAttributesDefinition   = "Lists all Search Attributes that can be used in list Workflow Queries."
	RemoveSearchAttributesDefinition = "Removes custom search attribute metadata only (Elasticsearch index schema is not modified)."

	// Cluster subcommand defintions
	HealthDefinition   = "Checks the health of the Frontend Service."
	DescribeDefinition = "Show information about the Cluster."
	SystemDefinition   = "Shows information about the system and its capabilities."
	UpsertDefinition   = "Add or update a remote Cluster."
	ListDefinition     = "List all remote Clusters."
	RemoveDefinition   = "Remove a remote Cluster."

	// Env subcommand definitions
	ListEnvDefinition = "Print all local configuration envs."
	GetDefinition     = "Print environmental properties."
	SetDefinition     = "Set environmental properties."
	DeleteDefinition  = "Delete an environment or environmental property."

	// Schedule definitions
	ScheduleCreateDefinition   = "Create a new Schedule."
	ScheduleCreateDescription  = "Takes a Schedule specification plus all the same args as starting a Workflow."
	ScheduleUpdateDefinition   = "Updates a Schedule with a new definition (full replacement, not patch)."
	ScheduleUpdateDescription  = "Takes a Schedule specification plus all the same args as starting a Workflow."
	ScheduleToggleDefinition   = "Pauses or unpauses a Schedule."
	ScheduleTriggerDefinition  = "Triggers an immediate action."
	ScheduleBackfillDefinition = "Backfills a past time range of actions."
	ScheduleDescribeDefinition = "Get Schedule configuration and current state."
	ScheduleDeleteDefinition   = "Deletes a Schedule."
	ScheduleListDefinition     = "Lists Schedules."
)

const BatchUsageText = `Batch commands change multiple [Workflow Executions](/concepts/what-is-a-workflow-execution)
by providing a [List Filter](/concepts/what-is-visibility) and the type of Batch job to execute.

The List Filter identifies the Workflow Executions that will be affected by the Batch job.
The Batch type determines the required parameters, along with what is affected on the Workflow Executions.

There are three types of Batch Jobs:
	- Signal: sends a [Signal](/concepts/what-is-a-signal) to the Workflow Executions specified by the List Filter.
	- Cancel: cancels the Workflow Executions specified by the List Filter.
	- Terminate: terminates the Workflow Executions specified by the List Filter.

A successfully started Batch job will return a Job ID.
Use this Job ID to execute other actions on the Batch job. 

Use the command options below to change the information returned by this command.
`

const DescribeBatchUsageText = `The ` + "`" + `temporal batch describe` + "`" + ` command shows the progress of an ongoing Batch job.

Pass a valid Job ID to return a Batch Job's information.
` + "`" + `temporal batch describe --jobid=MyJobId` + "`" + `

Use the command options below to change the information returned by this command.`

const ListBatchUsageText = `The ` + "`" + `temporal batch list` + "`" + ` command returns all Batch jobs. 
Batch Jobs can be returned for an entire Cluster or a single Namespace.
` + "`" + `temporal batch list --namespace=MyNamespace` + "`" + `

Use the command options below to change the information returned by this command.`

const TerminateBatchUsageText = `The ` + "`" + `temporal batch terminate` + "`" + ` command terminates a Batch job with the provided Job ID. 
For future reference, provide a reason for terminating the Batch Job.

` + "`" + `temporal batch terminate --job-id=MyJobId --reason=JobReason` + "`" + `

Use the command options below to change the information returned by this command.`

const WorkflowUsageText = `[Workflow](/concepts/what-is-a-workflow) commands perform operations on on [Workflow Executions](/concepts/what-is-a-workflow-execution).

Workflow commands use this syntax:
` + "`" + `temporal workflow COMMAND [ARGS]` + "`" + `
`

const StartWorkflowUsageText = `The ` + "`" + `temporal workflow start` + "`" + ` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution).
The Workflow and Run IDs are returned after starting the [Workflow](/concepts/what-is-a-workflow).

` + "`" + `temporal workflow start --task-queue=MyTaskQueue --type=MyWorkflow` + "`" + `

Use the command options below to change the information returned by this command.`

const ExecuteWorkflowUsageText = `The ` + "`" + `temporal workflow execute` + "`" + ` command starts a new [Workflow Execution](/concepts/what-is-a-workflow-execution) and prints its progress.
The command completes when the Workflow Execution completes.

[Workflows](/concepts/what-is-a-workflow) are executed with the following syntax:
` + "`" + `temporal workflow execute --workflow-id=meaningful-business-id --type=MyWorkflow --task-queue=MyTaskQueue` + "`" + `

Single quotes('') are used to wrap input as JSON.

` + "`" + `temporal workflow execute --workflow-id=meaningful-business-id --type-MyWorkflow --task-queue-MyTaskQueue --input='{"JSON": "Input"}'` + "`" + `

Use the command options below to change the information returned by this command.`

const DescribeWorkflowUsageText = `The ` + "`" + `temporal workflow describe` + "`" + ` command shows information about a given [Workflow Execution](/concepts/what-is-a-workflow-execution).

This information can be used to locate Workflow Executions that weren't able to run successfully.

` + "`" + `temporal workflow describe --workflow-id=meaningful-business-id` + "`" + `

Output can be shown as printed ('raw') or formatted to only show the Workflow Execution's auto-reset points.

` + "`" + `temporal workflow describe --workflow-id=meaningful-business-id --raw=true --reset-points=true` + "`" + `

Use the command options below to change the information returned by this command.`

const ListWorkflowUsageText = `The ` + "`" + `temporal workflow list` + "`" + ` command provides a list of [Workflow Executions](/concepts/what-is-a-workflow-execution) that meet the criteria of a given [Query](/concepts/what-is-a-query).
By default, this command returns up to 10 closed Workflow Executions.

` + "`" + `temporal workflow list --query=MyQuery` + "`" + `

The command can also return a list of archived Workflow Executions.

` + "`" + `temporal workflow list --archived=true` + "`" + `

Use the command options below to change the information returned by this command.`

const QueryWorkflowUsageText = `The ` + "`" + `temporal workflow query` + "`" + ` command sends a [Query](/concepts/what-is-a-query) to a [Workflow Execution](/concepts/what-is-a-workflow-execution).

Queries can retrieve all or part of the Workflow state within given parameters.
Queries can also be used on completed [Workflows](/concepts/what-is-a-workflow-execution).

` + "`" + `temporal workflow query --workflow-id=meaningful-business-id --type=MyQueryType` + "`" + `

Use the command options listed below to change the information returned by this command.`

const CancelWorkflowUsageText = `The ` + "`" + `temporal workflow cancel` + "`" + ` command cancels a [Workflow Execution](/concepts/what-is-a-workflow-execution).

Canceling a running Workflow Execution records a [` + "`" + `WorkflowExecutionCancelRequested` + "`" + ` event](/references/events#workflow-execution-cancel-requested) in the [Event History](/concepts/what-is-an-event-history).
A new [Command Task](/concepts/what-is-a-command) will be scheduled, and the Workflow Execution performs cleanup work.

` + "`" + `temporal workflow cancel --workflow-id=meaningful-business-id` + "`" + `

In addition to Workflow IDs, Workflows can also be [Signaled](/concepts/what-is-a-signal) by a [Query](/concepts/what-is-a-query).
` + "`" + `temporal workflow cancel --query=MyQuery` + "`" + `

Use the options listed below to change the behavior of this command.`

const TerminateWorkflowUsageText = `The ` + "`" + `temporal workflow terminate` + "`" + ` command terminates a [Workflow Execution](/concepts/what-is-a-workflow-execution)

Terminating a running Workflow Execution records a [` + "`" + `WorkflowExecutionTerminated` + "`" + ` event](/references/events#workflowexecutionterminated) as the closing Event in the [Event History](/concepts/what-is-an-event-history).
Any further [Command](/concepts/what-is-a-command) Tasks cannot be scheduled after running this command.

Workflow terminations require a valid [Workflow ID](/concepts/what-is-a-workflow-id) to function.
` + "`" + `temporal workflow terminate --workflow-id=meaningful-business-id` + "`" + `

Use the options listed below to change termination behavior.`

const ResetWorkflowUsageText = `The ` + "`" + `temporal workflow reset` + "`" + ` command resets a [Workflow Execution](/concepts/what-is-a-workflow-execution).
A reset allows the Workflow to resume from a certain point without losing its parameters or [Event History](/concepts/what-is-an-event-history).

The Workflow Execution can be set to a given [Event Type](/concepts/what-is-an-event).
` + "`" + `temporal workflow reset --workflow-id=meaningful-business-id --type=LastContinuedAsNew` + "`" + `

The Workflow Execution can also be reset to any Event after []` + "`" + `WorkflowTaskStarted` + "`" + `](/references/events#workflowtaskstarted).
` + "`" + `temporal workflow reset --workflow-id=meaningful-business-id --event-id=MyLastEvent` + "`" + `

Use the options listed below to change reset behavior.`

const ResetBatchUsageText = `The ` + "`" + `temporal workflow reset-batch` + "`" + ` command resets multiple [Workflow Executions](/concepts/what-is-a-workflow-execution) by ` + "`" + `resetType` + "`" + `.
Resetting a [Workflow](/concepts/what-is-a-workflow) resumes it from a certain point without losing its parameters or [Event History](/concepts/what-is-an-event-history).

The set of Workflow Executions to reset can be specified in an input file.
The input file must have a [Workflow ID](/concepts/what-is-a-workflow-id) on each line.

` + "`" + `temporal workflow reset-batch --input-file=MyInput --input-separator="\t"` + "`" + `

Workflow Executions can also be found by [Query](/concepts/what-is-a-query).
` + "`" + `temporal workflow reset-batch --query=MyQuery

Use the options listed below to change reset behavior.`

const TaskQueueUsageText = `Task Queue commands allow operations to be performed on [Task Queues](/concepts/what-is-a-task-queue).
To run a Task Queue command, run ` + "`" + `temporal task-queue [command] [command options]` + "`" + `
`

const DescribeTaskQueueUsageText = `The ` + "`" + `temporal task-queue describe` + "`" + ` command provides [poller](/application-development/worker-performance#poller-count) information for a given [Task Queue](/concepts/what-is-a-task-queue).

The [Server](/concepts/what-is-the-temporal-server) records the last time of each poll request.
A ` + "`" + `LastAccessTime` + "`" + ` value in excess of one minute can indicate the Worker is at capacity (all Workflow and Activity slots are full) or that the Worker has shut down.
[Workers](/concepts/what-is-a-worker) are removed if 5 minutes have passed since the last poll request.

Information about the Task Queue can be returned to troubleshoot server issues.

` + "`" + `temporal task-queue describe --task-queue=MyTaskQueue --task-queue-type="activity"` + "`" + `

Use the options listed below to modify what this command returns.`

const ScheduleUsageText = `Schedule commands allow the user to create, use, and update [Schedules](/concepts/what-is-a-schedule).
Schedules control when certain Actions for a [Workflow Execution](/concepts/what-is-a-workflow-execution) are performed, making it a useful tool for automation.

Schedule commands follow this syntax:

` + "`" + `temporal schedule [command] [command options]` + "`" + `.
`
const OperatorUsageText = `Operator commands enable actions on [Namespaces](/concepts/what-is-a-namespace), [Search Attributes](/concepts/what-is-a-search-attribute), and [Temporal Clusters](/concepts/what-is-a-temporal-cluster).
These actions are performed through subcommands.

To run an Operator command, run ` + "`" + `temporal operator [command] [subcommand] [command options]` + "`" + `.
`

const ActivityUsageText = `Activity commands operate on [Activity Executions](/concepts/what-is-an-activity-execution).

Activity commands follow this syntax:
` + "`" + `temporal activity [command] [command options]` + "`" + `
`

const CompleteActivityUsageText = `The ` + "`" + `temporal activity complete` + "`" + ` command completes an [Activity Execution](/concepts/what-is-an-activity-execution).
The result given upon return can also be set with the command.

` + "`" + `temporal activity complete --activity-id=MyActivity --result=ActivityComplete` + "`" + `

Use the options listed below to change the behavior of this command.`

const FailActivityUsageText = `The ` + "`" + `temporal activity fail` + "`" + ` command fails an [Activity Execution](/concepts/what-is-an-activity-execution).
The Activity must be running on a valid [Workflow](/concepts/what-is-a-workflow).
` + "`" + `temporal fail --workflow-id=meaningful-business-id --activity-id=MyActivity` + "`" + ` 

Use the options listed below to change the behavior of this command.`

const HealthUsageText = `The ` + "`" + `temporal operator cluster health` + "`" + ` command checks the health of the [Frontend Service](/concepts/what-is-a-frontend-service).
A successful execution returns a list of [Cluster](/concepts/what-is-a-temporal-cluster) metrics.

Use the options listed below to change the behavior and output of this command.`

const ClusterUsageText = `Cluster commands enable operations on [Temporal Clusters](/concepts/what-is-a-temporal-cluster).

Cluster commands follow this syntax:
` + "`" + `temporal operator cluster COMMAND [ARGS]` + "`" + ``

const ClusterDescribeUsageText = `The ` + "`" + `temporal operator cluster describe` + "`" + ` command shows information about the [Cluster](/concepts/what-is-a-temporal-cluster).
This information can include information about other connected services, such as a remote [Codec Server](/concepts/what-is-a-codec-server).

Use the options listed below to change the output of this command.`

const ClusterSystemUsageText = `The ` + "`" + `temporal operator cluster system` + "`" + ` command provides information about the system the [Cluster](/concepts/what-is-a-temporal-cluster) is running on.
This information can be used to diagnose problems occurring in the [Temporal Server](/concepts/what-is-the-temporal-server).

` + "`" + `temporal operator cluster system` + "`" + `

Use the options listed below to change this command's output.`

const ClusterUpsertUsageText = `The ` + "`" + `temporal operator cluster upsert` + "`" + ` command allows the user to add or update a remote [Cluster](/concepts/what-is-a-temporal-cluster).
` + "`" + `temporal operator cluster upsert --frontend-address="127.0.2.1"` + "`" + `

Upserting can also be used to enable or disabled cross-cluster connection.
` + "`" + `temporal operator cluster upsert --enable-connection=true` + "`" + `

Use the options listed below to change the behavior of this command.`

const ClusterListUsageText = `The ` + "`" + `temporal operator cluster list` + "`" + ` command prints a list of all remote [Clusters](/concepts/what-is-a-temporal-cluster) on the system.

` + "`" + `temporal operator cluster list` + "`" + `

Use the options listed below to change the command's behavior.`

const ClusterRemoveUsageText = `The ` + "`" + `temporal operator cluster remove` + "`" + ` command removes a remote [Cluster](/concepts/what-is-a-temporal-cluster) from the system.

` + "`" + `temporal operator cluster remove --name=SomeCluster` + "`" + `

Use the options listed below to change the command's behavior.`

const EnvUsageText = `Environment (or 'env') commands allow the user to configure the properties for the environment in use.`

const EnvListUsageText = `List all environments`

const EnvGetUsageText = `The ` + "`" + `temporal env get` + "`" + ` command prints the environmental properties for the environment in use.

Passing the 'local' [Namespace](/concepts/what-is-a-namespace) returns the name, address, and certificate paths for the local environment.
` + "`" + `temporal env get local` + "`" + `
` + "`" + `
Output:
tls-cert-path  /home/my-user/certs/cluster.cert  
tls-key-path   /home/my-user/certs/cluster.key   
address        127.0.0.1:7233                    
namespace      someNamespace 
` + "`" + `

Output can be narrowed down to a specific environmental property.
` + "`" + `temporal env get local.tls-key-path` + "`" + `
` + "`" + `tls-key-path  /home/my-user/certs/cluster.key` + "`" + `

Use the options listed below to change the command's behavior.`

const EnvSetUsageText = `The ` + "`" + `temporal env set` + "`" + ` command sets the value for an environmental property.

Properties (such as the frontend address) can be set for the entire system:
` + "`" + `temporal env set local.address 127.0.0.1:7233` + "`" + `

Use the options listed below to change the command's behavior.`

const EnvDeleteUsageText = `The ` + "`" + `temporal env delete` + "`" + ` command deletes a given environment or environmental property.

Delete an environment (such as 'local') and its saved values by passing a valid [Namespace](/concepts/what-is-a-namespace) name.

` + "`" + `temporal env delete local` + "`" + `

Use the options listed below to change the command's behavior.`

const NamespaceUsageText = `Namespace commands perform operations on [Namespaces](/concepts/what-is-a-namespace) contained in the [Temporal Cluster](/concepts/what-is-a-temporal-cluster).

Namespace commands follow this syntax:
` + "`" + `temporal operator namespace COMMAND [ARGS]` + "`" + `
`
const NamespaceDescribeUsageText = `The ` + "`" + `temporal operator namespace describe` + "`" + ` command provides [Namespace](/concepts/what-is-a-namespace) information.
Namespaces are identified by Namespace ID.

` + "`" + `temporal operator namespace describe --namespace-id=meaningful-business-id` + "`" + `

Use the options listed below to change the command's output.`

const NamespaceListUsageText = `The ` + "`" + `temporal operator namespace list` + "`" + ` command lists all [Namespaces](/namespaces) on the [Server](/concepts/what-is-a-frontend-service).

` + "`" + `temporal operator namespace list` + "`" + `

Use the options listed below to change the command's output.`

const NamespaceCreateUsageText = `The ` + "`" + `temporal operator namespace create` + "`" + ` command creates a new [Namespace](/concepts/what-is-a-namespace) on the [Server](/concepts/what-is-a-frontend-service).
Namespaces can be created on the active [Cluster](/concepts/what-is-a-temporal-cluster), or any named Cluster.
` + "`" + `temporal operator namespace --cluster=MyCluster` + "`" + `

Global Namespaces can also be created.
` + "`" + `temporal operator namespace create --global` + "`" + `

Other settings, such as [retention](/concepts/what-is-a-retention-period) and [Visibility Archival State](/concepts/what-is-visibility), can be configured as needed.
For example, the Visibility Archive can be set on a separate URI.
` + "`" + `temporal operator namespace create --retention=RetentionMyWorkflow --visibility-archival-state="enabled" --visibility-uri="some-uri"` + "`" + `

Use the options listed below to change the command's behavior.`

const NamespaceUpdateUsageText = `The ` + "`" + `temporal operator namespace update` + "`" + ` command updates a [Namespace](/concepts/what-is-a-namespace).

Namespaces can be assigned a different active [Cluster](/concepts/what-is-a-temporal-cluster).
` + "`" + `temporal operator namespace update --active-cluster=NewActiveCluster` + "`" + `

Namespaces can also be promoted to global Namespaces.
` + "`" + `temporal operator namespace --promote-global=true` + "`" + `

Any [Archives](/concepts/what-is-archival) that were previously enabled or disabled can be changed through this command.
However, URI values for archival states cannot be changed after the states are enabled.
` + "`" + `temporal operator namespace update --history-archival-state="enabled" --visibility-archival-state="disabled"` + "`" + `

Use the options listed below to change the command's behavior.`

const NamespaceDeleteUsageText = `The ` + "`" + `temporal operator namespace delete` + "`" + ` command deletes a given [Namespace](/concepts/what-is-a-namespace) from the system.

Its syntax is:
` + "`" + `temporal operator namespace delete [ARGS]` + "`" + `

Use the command options below to change the information returned by this command.`

const ScheduleCreateUsageText = `The ` + "`" + `temporal schedule create` + "`" + ` command creates a new [Schedule](/concepts/what-is-a-schedule).
Newly created Schedules return a Schedule ID to be used in other Schedule commands.

Schedules are passed in the following format:
` + "`" + `` + "`" + `` + "`" + `
temporal schedule create \
		--sid 'your-schedule-id' \
		--cron '3 11 * * Fri' \
		--wid 'your-workflow-id' \
		--tq 'your-task-queue' \
		--type 'YourWorkflowType' 
` + "`" + `` + "`" + `` + "`" + `

Any combination of ` + "`" + `--cal` + "`" + `, ` + "`" + `--interval` + "`" + `, and ` + "`" + `--cron` + "`" + ` is supported.
Actions will be executed at any time specified in the Schedule.

Use the options provided below to change the command's behavior.`

const ScheduleUpdateUsageText = `The ` + "`" + `temporal schedule update` + "`" + ` command updates an existing [Schedule](/concepts/what-is-a-schedule).

Updated Schedules need to follow a certain format:
` + "`" + `` + "`" + `` + "`" + `
temporal schedule update 			\
		--sid 'your-schedule-id' 	\
		--cron '3 11 * * Fri' 		\
		--wid 'your-workflow-id' 	\
		--tq 'your-task-queue' 		\
		--type 'YourWorkflowType' 
` + "`" + `` + "`" + `` + "`" + `

Updating a Schedule takes the given options and replaces the entire configuration of the Schedule with what's provided. 
If you only change one value of the Schedule, be sure to provide the other unchanged fields to prevent them from being overwritten.

Use the command options below to change the information returned by this command.`

const ScheduleToggleUsageText = `The ` + "`" + `temporal schedule toggle` + "`" + ` command can pause and unpause a [Schedule](/concepts/what-is-a-schedule).

Toggling a Schedule requires a reason. 
Use ` + "`" + `--reason` + "`" + ` to note the issue leading to the pause or unpause.

Schedule toggles follow this syntax:
` + "`" + ` temporal schedule toggle --sid 'your-schedule-id' --pause --reason "paused because the database is down"` + "`" + `
` + "`" + `temporal schedule toggle --sid 'your-schedule-id' --unpause --reason "the database is back up"` + "`" + `

Use the command options below to change the information returned by this command.`

const ScheduleTriggerUsageText = `The ` + "`" + `temporal schedule trigger` + "`" + ` command triggers an immediate action with a given [Schedule](/concepts/what-is-a-schedule).
By default, this action is subject to the Overlap Policy of the Schedule.

Schedule triggers follow this syntax:
` + "`" + `temporal schedule trigger` + "`" + ` can be used to start a Workflow Run immediately.
` + "`" + `temporal schedule trigger --sid 'your-schedule-id'` + "`" + ` 

The Overlap Policy of the Schedule can be overridden as well.
` + "`" + `temporal schedule trigger --sid 'your-schedule-id' --overlap-policy 'AllowAll'` + "`" + `

Use the options provided below to change this command's behavior.`

const ScheduleBackfillUsageText = `The ` + "`" + `temporal schedule backfill` + "`" + ` command executes Actions ahead of their specified time range. 
Backfilling can fill in [Workflow Runs](/concepts/what-is-a-run-id) from a time period when the Schedule was paused, or from before the Schedule was created. 

Schedule backfills require a valid Schedule ID, along with the time in which to run the Schedule and a change to the overlap policy.
` + "`" + `` + "`" + `` + "`" + `
temporal schedule backfill --sid 'your-schedule-id' \
		--overlap-policy 'BufferAll' 				\
		--start-time '2022-05-0101T00:00:00Z'		\
		--end-time '2022-05-31T23:59:59Z'
` + "`" + `` + "`" + `` + "`" + `

Use the options provided below to change this command's behavior.`

const ScheduleDescribeUsageText = `The ` + "`" + `temporal schedule describe` + "`" + ` command shows the current [Schedule](/concepts/what-is-a-schedule) configuration.
This command also provides information about past, current, and future [Workflow Runs](/concepts/what-is-a-run-id).

` + "`" + `temporal schedule describe --sid 'your-schedule-id' [command options] ` + "`" + `

Use the options below to change this command's output.`

const ScheduleDeleteUsageText = `The ` + "`" + `temporal schedule delete` + "`" + ` command deletes a [Schedule](/concepts/what-is-a-schedule).
Deleting a Schedule does not affect any [Workflows](/concepts/what-is-a-workflow) started by the Schedule.

[Workflow Executions](/concepts/what-is-a-workflow-execution) started by Schedules can be cancelled or terminated.
In additon, Workflow Executions started by a Schedule can be identified by their [Search Attributes](/concepts/what-is-a-search-attribute), making them targetable by batch command for termination.

` + "`" + `temporal schedule delete --sid 'your-schedule-id' [command options] ` + "`" + `

Use the options below to change the behavior of this command.`

const ScheduleListUsageText = `The ` + "`" + `temporal schedule list` + "`" + ` command lists all [Schedule](/concepts/what-is-a-schedule) configurations.
Listing Schedules in [Standard Visibility](/concepts/what-is-standard-visibility) will only provide Schedule IDs.

` + "`" + `temporal schedule list` + "`" + `

Use the options below to change the behavior of this command.`

const SearchAttributeUsageText = `Search Attribute commands enable operations for the creation, listing, and removal of [Search Attributes](/concepts/what-is-a-search-attribute) for [Workflow Executions](/concepts/what-is-a-workflow-execution).

Search Attribute commands follow this syntax:
` + "`" + `temporal operator search-attribute COMMAND [ARGS]` + "`" + `
`

const SearchAttributeCreateUsageText = `The ` + "`" + `temporal operator search-attribute create` + "`" + ` command adds one or more custom [Search Attributes](/concepts/what-is-a-search-attribute) to a [Workflow Execution](/concepts/what-is-a-workflow-execution).
These Search Attributes can be used to [filter a list](/concepts/what-is-a-list-filter) of Workflow Executions that contain the given Search Attributes in their metadata.

Use the options listed below to change the command's behavior.`

const SearchAttributeListUsageText = `The ` + "`" + `temporal operator search-attribute list` + "`" + ` command displays a list of all [Search Attributes](/concepts/what-is-a-search-attribute) for a [Workflow Execution](/concepts/what-is-a-workflow-execution).
Attributes on this list can be used in [Queries](/concepts/what-is-a-query).

` + "`" + ` temporal workflow list --query` + "`" + `.

Use the options listed below to change the command's behavior.`

const SearchAttributeRemoveUsageText = `The ` + "`" + `temporal operator search-attribute remove` + "`" + ` command removes custom [Search Attribute](/concepts/what-is-a-search-attribute) metadata from a [Workflow Execution](/concepts/what-is-a-workflow-execution).
This command does not remove custom Search Attributes from Elasticsearch or change the index schema.

Use the options listed below to change the command's behavior.`

const TaskQueueListPartitionUsageText = `The ` + "`" + `temporal task-queue list-partition` + "`" + ` command displays the partitions of a [Task Queue](/concepts/what-is-a-task-queue), along with the matching node they are assigned to.

Use the options listed below to change the command's behavior.`

const WorkflowShowUsageText = `The ` + "`" + `temporal workflow show` + "`" + ` command provides the [Event History](/concepts/what-is-an-event-history) for a [Workflow Execution](/concepts/what-is-a-workflow-execution).

Use the options listed below to change the command's behavior.`

const WorkflowStackUsageText = `The ` + "`" + `temporal workflow stack` + "`" + ` command queries a [Workflow Execution](/concepts/what-is-a-workflow-execution) with ` + "`" + `--stack-trace` + "`" + ` as the [Query](/concepts/what-is-a-query#stack-trace-query) type.
Returning the stack trace of all the threads owned by a Workflow Execution can be great for troubleshooting in production.

Use the options listed below to change the command's behavior.`

const WorkflowSignalUsageText = `The ` + "`" + `temporal workflow signal` + "`" + ` command is used to [Signal](/concepts/what-is-a-signal) a [Workflow Execution](/concepts/what-is-a-workflow-execution) by [ID](/concepts/what-is-a-workflow-id) or [List Filter](/concepts/what-is-a-list-filter).

Use the options listed below to change the command's behavior.`

const WorkflowCountUsageText = `The ` + "`" + `temporal workflow count` + "`" + ` command returns a count of [Workflow Executions](/concepts/what-is-a-workflow-execution).
This command requires Elasticsearch to be enabled.

Use the options listed below to change the command's behavior.`

const WorkflowDeleteUsageText = `The ` + "`" + `temporal workflow delete` + "`" + ` command deletes the specified [Workflow Execution](/concepts/what-is-a-workflow-execution).

Use the options listed below to change the command's behavior.`

const WorkflowTraceUsageText = `The ` + "`" + `temporal workflow trace` + "`" + ` command tracks the progress of a [Workflow Execution](/concepts/what-is-a-workflow-execution) and any  [Child Workflows](/concepts/what-is-a-child-workflow-execution) it generates.

Use the options listed below to change the command's behavior.`

const WorkflowUpdateUsageText = `The ` + "`" + `temporal workflow update` + "`" + ` command synchronously updates a running [Workflow Execution](/concepts/what-is-a-workflow-execution).

Use the options listed below to change the command's behavior.`

const ServerUsageText = `Server commands allow you to start and manage the [Temporal Server](/concepts/what-is-the-temporal-server) from the command line.

Currently, ` + "`" + `cli` + "`" + ` server functionality extends to starting the Server. 

Server commands follow this syntax:
` + "`" + `temporal server COMMAND` + "`" + `
`

const StartDevUsageText = `The ` + "`" + `temporal server start-dev` + "`" + ` command starts the Temporal Server on ` + "`" + `localhost:7233` + "`" + `.
The results of any command run on the Server can be viewed at http://localhost:7233.
`

const CustomTemplateHelpCLI = `NAME:
   {{template "helpNameTemplate" .}}{{if .Description}}

DESCRIPTION:
   {{template "descriptionTemplate" .}}{{end}}{{if .UsageText}}

USAGE:
   {{wrap .UsageText 3 | markdown2Text}}{{end}}{{if .VisibleFlagCategories}}
   {{template "visibleFlagCategoryTemplate" .}}{{else if .VisibleFlags}}

DISPLAY OPTIONS:
   {{template "visibleFlagTemplate" .}}{{end}}
`
