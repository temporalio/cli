package common

// TODO: add links to Temporal concepts after confirming the format won't screw with the command line
const (
	// Main command definitions
	WorkflowDefinition = "Operations that can be performed on Workflows."
	ActivityDefinition = "Operations that can be performed on Workflow Activities."
	TaskQueueDefinition = "Operations performed on Task Queues."
	ScheduleDefinition = "Operations performed on Schedules."
	BatchDefinition = "Operations performed on Batch jobs. Use Workflow commands with --query flag to start batch jobs."
	OperatorDefinition = "Operations on Temporal Server."
	EnvDefinition = "Manage environmental configurations on Temporal Client."

	// Workflow subcommand definitions
	StartWorkflowDefinition = "Starts a new Workflow Execution."
	StartWorkflowUsageText = "When invoked successfully, the Workflow and Run Ids of the recently started Workflow are returned."

	ExecuteWorkflowDefinition = "Start a new Workflow Execution and prints its progress."
	ExecuteWorkflowUsageText = "Single quotes('') are used to wrap input as JSON."

	DescribeWorkflowDefinition = "Show information about a Workflow Execution."
	DescribeWorkflowUsageText = "This information can be used to locate a Workflow Execution that failed."

	ListWorkflowDefinition = "List Workflow Executions based on a Query."
	ListWorkflowUsageText = "By default, this command lists up to 10 closed Workflow Executions."

	ShowWorkflowDefinition = "Show Event History for a Workflow Execution."

	QueryWorkflowDefinition = "Query a Workflow Execution."
	QueryWorkflowUsageText = "Queries can retrieve all or part of the Workflow state within given parameters.\nQueries can also be used on completed Workflows."

	StackWorkflowDefinition = "Query a Workflow Execution with __stack_trace as the query type."
	SignalWorkflowDefinition = "Signal Workflow Execution by Id or List Filter."
	CountWorkflowDefinition = "Count Workflow Executions (requires ElasticSearch to be enabled)."

	CancelWorkflowDefinition = "Cancel a Workflow Execution."
	CancelWorkflowUsageText = "Canceling a running Workflow Execution records a `WorkflowExecutionCancelRequested` event in the Event History.\n\nAfter cancellation, the Workflow Execution can perform cleanup work,and a new command task will be scheduled."

	TerminateWorkflowDefinition = "Terminate Workflow Execution by Id or List Filter."
	TerminateWorkflowUsageText = "Terminating a running Workflow records a `WorkflowExecutionTerminated` event as the closing event.\n\nAny further command tasks cannot be scheduled after running `terminate`."

	DeleteWorkflowDefinition = "Deletes a Workflow Execution."

	ResetWorkflowDefinition = "Resets a Workflow Execution by Event Id or reset type."
	ResetWorkflowUsageText = "A reset allows the Workflow to be resumed from a certain point without losing your parameters or Event History."

	ResetBatchUsageText = "Resetting a Workflow allows the process to resume from a certain point without losing your parameters or Event History."

	TraceWorkflowDefinition = "Trace progress of a Workflow Execution and its children."

	// Activity subcommand definitions
	CompleteActivityDefinition = "Completes an Activity."
	FailActivityDefinition = "Fails an Activity."

	// Task Queue subcommand definitions
	DescribeTaskQueueDefinition = "Describes the Workers that have recently polled on this Task Queue"
	DescribeTaskQueueUsageText =  "The Server records the last time of each poll request.\n\nPoll requests can last up to a minute, so a LastAccessTime under a minute is normal.\nIf it's over a minute, then likely either the Worker is at capacity (all Workflow and Activity slots are full) or it has shut down.\nOnce it has been 5 minutes since the last poll request, the Worker is removed from the list.\n\nRatePerSecond is the maximum Activities per second the Worker will execute."

	ListPartitionTaskQueueDefinition = "Lists the Task Queue's partitions and which matching node they are assigned to."

	// Schedule subcommand definitions
	ScheduleUsageText = "These commands allow Schedules to be created, used, and updated."

	// Batch subcommand definitions
	BatchUsageText = "Batch jobs run in the background and affect Workflow Executions one at a time.\n\nIn `cli`, the batch commands are used to view the status of Batch jobs, and to terminate them.\nA successfully started Batch job returns a Job Id, which is needed to execute batch commands.\n\nTerminating a batch job does not roll back the operations already performed by the job itself."

	DescribeBatchJobDefinition = "Describe a Batch operation job." 
	DescribeBatchUsageText = "This command shows the progress of an ongoing Batch job."

	ListBatchJobsDefinition = "List Batch operation jobs."
	ListBatchUsageText = "When used, all Batch operation jobs within the system are listed."

	TerminateBatchJobDefinition = "Stop a Batch operation job."
	TerminateBatchUsageText =  "When used, the Batch job with the provided Batch Id is terminated."


	// Operator subcommands and additional text
	OperatorUsageText = "These commands enable operations on Namespaces, Search Attributes, and Temporal Clusters."

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
	DeleteDefinition = "Delete environment or environmental property."
)