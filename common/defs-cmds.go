package common

const (
	// Main command definitions
	WorkflowDefinition = "Operations that can be performed on [Workflows](https://docs.temporal.io/workflows)."
	ActivityDefinition = "Operations that can be performed on Workflow [Activities](https://docs.temporal.io/activities)."
	TaskQueueDefinition = "Operations performed on [Task Queues](https://docs.temporal.io/tasks/#task-queue)."
	ScheduleDefinition = "Operations performed on [Schedules](https://docs.temporal.io/workflows/#schedule)."
	BatchDefinition = "Operations performed on Batch jobs. Use [Workflows](https://docs.temporal.io/workflows) commands with --query flag to start batch jobs."
	OperatorDefinition = "Operations performed on the [Temporal Server](https://docs.temporal.io/clusters/#temporal-server)."
	EnvDefinition = "Manage environmental configurations on [Temporal Client](https://docs.temporal.io/temporal/#temporal-client)."

	// Workflow subcommand definitions
	StartWorkflowDefinition = "Starts a new [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution)."
	StartWorkflowUsageText = "When invoked successfully, the Workflow and Run Ids of the recently started [Workflow](https://docs.temporal.io/workflows) are returned."

	ExecuteWorkflowDefinition = "Start a new [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) and prints its progress."
	ExecuteWorkflowUsageText = "Single quotes('') are used to wrap input as JSON."

	DescribeWorkflowDefinition = "Show information about a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution)."
	DescribeWorkflowUsageText = "This information can be used to locate a Workflow Execution that failed."

	ListWorkflowDefinition = "List [Workflow Executions](https://docs.temporal.io/workflows/#workflow-execution) based on a [Query](https://docs.temporal.io/workflows/#query)."
	ListWorkflowUsageText = "By default, this command lists up to 10 closed Workflow Executions."

	ShowWorkflowDefinition = "Show [Event History](https://docs.temporal.io/workflows/#event-history) for a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution)."

	QueryWorkflowDefinition = "[Query](https://docs.temporal.io/workflows/#query) a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution)."
	QueryWorkflowUsageText = "Queries can retrieve all or part of the Workflow state within given parameters.\nQueries can also be used on completed [Workflows](https://docs.temporal.io/workflows/)."

	StackWorkflowDefinition = "Query a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) with [__stack_trace](https://docs.temporal.io/workflows/#stack-trace-query) as the query type."
	SignalWorkflowDefinition = "Signal [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) by Id or [List Filter](https://docs.temporal.io/visibility/#list-filter)."
	CountWorkflowDefinition = "Count Workflow Executions (requires ElasticSearch to be enabled)."

	CancelWorkflowDefinition = "Cancel a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution)."
	CancelWorkflowUsageText = "Canceling a running Workflow Execution records a [`WorkflowExecutionCancelRequested` event](https://docs.temporal.io/references/events/#workflowexecutioncanceled) in the [Event History](https://docs.temporal.io/workflows/#event-history).\n\nAfter cancellation, the Workflow Execution can perform cleanup work,and a new [Command](https://docs.temporal.io/workflows/#command) task will be scheduled."

	TerminateWorkflowDefinition = "Terminate [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) by Id or [List Filter](https://docs.temporal.io/visibility/#list-filter)."
	TerminateWorkflowUsageText = "Terminating a running Workflow records a [`WorkflowExecutionTerminated` event](https://docs.temporal.io/references/events/#workflowexecutionterminated) as the closing event.\n\nAny further [Command](https://docs.temporal.io/workflows/#command) tasks cannot be scheduled after running `terminate`."

	DeleteWorkflowDefinition = "Deletes a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution)."

	ResetWorkflowDefinition = "Resets a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) by Event Id or reset type."
	ResetWorkflowUsageText = "A reset allows the Workflow to be resumed from a certain point without losing your parameters or [Event History](https://docs.temporal.io/workflows/#event-history)."

	ResetBatchUsageText = "Resetting a Workflow allows the process to resume from a certain point without losing your parameters or [Event History](https://docs.temporal.io/workflows/#event-history)."

	TraceWorkflowDefinition = "Trace progress of a [Workflow Execution](https://docs.temporal.io/workflows/#workflow-execution) and its [children](https://docs.temporal.io/workflows/#child-workflow)."

	// Activity subcommand definitions
	CompleteActivityDefinition = "Completes an [Activity](https://docs.temporal.io/activities)."
	FailActivityDefinition = "Fails an [Activity](https://docs.temporal.io/activities)."

	// Task Queue subcommand definitions
	DescribeTaskQueueDefinition = "Describes the [Workers](https://docs.temporal.io/workers) that have recently polled on this [Task Queue](https://docs.temporal.io/tasks/#task-queue)"
	DescribeTaskQueueUsageText =  "The [Server](https://docs.temporal.io/clusters/#temporal-server) records the last time of each poll request.\n\nPoll requests can last up to a minute, so a LastAccessTime under a minute is normal.\nIf it's over a minute, then likely either the Worker is at capacity (all [Workflow](https://docs.temporal.io/workflows/) and [Activity](https://docs.temporal.io/activities) slots are full) or it has shut down.\nOnce it has been 5 minutes since the last poll request, the Worker is removed from the list.\n\nRatePerSecond is the maximum Activities per second the Worker will execute."

	ListPartitionTaskQueueDefinition = "Lists the [Task Queue's](https://docs.temporal.io/tasks/#task-queue) partitions and which matching node they are assigned to."

	// Schedule subcommand definitions
	ScheduleUsageText = "These commands allow [Schedules](https://docs.temporal.io/workflows/#schedule) to be created, used, and updated."

	// Batch subcommand definitions
	BatchUsageText = "Batch Jobs run in the background and affect [Workflow Executions](https://docs.temporal.io/workflows/#workflow-execution) one at a time.\n\nIn `cli`, the Batch Commands are used to view the status of Batch jobs, and to terminate them.\nA successfully started Batch job returns a Job Id, which is needed to execute Batch Commands.\n\nTerminating a Batch Job does not roll back the operations already performed by the job itself."

	DescribeBatchJobDefinition = "Describe a Batch operation job." 
	DescribeBatchUsageText = "This command shows the progress of an ongoing Batch job."

	ListBatchJobsDefinition = "List Batch operation jobs."
	ListBatchUsageText = "When used, all Batch operation jobs within the system are listed."

	TerminateBatchJobDefinition = "Stop a Batch operation job."
	TerminateBatchUsageText =  "When used, the Batch job with the provided Batch Id is terminated."


	// Operator subcommands and additional text
	OperatorUsageText = "These commands enable operations on [Namespaces](https://docs.temporal.io/namespaces), [Search Attributes](https://docs.temporal.io/visibility#search-attribute), and [Temporal Clusters](https://docs.temporal.io/clusters)."

	NamespaceDefinition = "Operations applying to [Namespaces](https://docs.temporal.io/namespaces)."
	SearchAttributeDefinition = "Operations applying to [Search Attributes](https://docs.temporal.io/visibility#search-attribute)."
	ClusterDefinition = "Operations for running a [Temporal Cluster](https://docs.temporal.io/clusters)."

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
	DeleteDefinition = "Delete environment or environmental property."
)