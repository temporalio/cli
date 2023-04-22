package common

const (
	// Shared flag definitions
	FlagEnvDefinition = "Environment to read environmental variables from."
	FlagAddrDefinition = "The host and port (formatted as host:port) for the Temporal Frontend Service."
	FlagNSAliasDefinition = "Identifies a Namespace in the Temporal Workflow."
	FlagMetadataDefinition = "Contains gRPC metadata to send with requests (formatted as key=value). Values must be in a valid JSON format."
	FlagTLSDefinition = "Enable TLS"
	FlagTLSCertPathDefinition = "Path to x509 certificate."
	FlagTLSKeyPathDefinition = "Path to private certificate key."
	FlagTLSCaPathDefinition = "Path to server CA certificate."
	FlagTLSDisableHVDefinition = "Disables TLS host-name verification."
	FlagTLSServerNameDefinition = "Overrides target TLS server name."
	FlagContextTimeoutDefinition = "Optional timeout for the context of an RPC call (in seconds)."
	FlagCodecEndpointDefinition = "Endpoint for a remote Codec Server."
	FlagCodecAuthDefinition = "Sets the authorization header on requests to the Codec Server."
	FlagArchiveDefinition = "List archived Workflow Executions. Currently an experimental feature."

	// Execution flags
	FlagWorkflowId = "Workflow Id"
	FlagRunIdDefinition = "Run Id"
	FlagJobIDDefinition = "Batch Job Id"
	FlagScheduleIDDefinition = "Schedule Id"

	// ShowWorkflow flags
	FlagOutputFilenameDefinition = "Serializes Event History to a file."
	FlagMaxFieldLengthDefinition = "Maximum length for each attribute field."
	FlagResetPointsOnlyDefinition = "Only show Workflow Events that are eligible for reset."
	FlagFollowAliasDefinition = "Follow the progress of a Workflow Execution."

	// StartWorkflow flags
	FlagWFTypeDefinition = "Workflow Type name."
	FlagTaskQueueDefinition = "Task Queue"
	FlagWorkflowRunTimeoutDefinition = "Timeout (in seconds) of a Workflow Run."
	FlagWorkflowExecutionTimeoutDefinition = "Timeout (in seconds) for a WorkflowExecution, including retries and `ContinueAsNew` tasks."
	FlagWorkflowTaskTimeoutDefinition = "Start-to-close timeout for a Workflow Task (in seconds)."
	FlagCronScheduleDefinition = "Optional Cron Schedule for the Workflow. Cron spec is formatted as: \n" +
	"\t┌───────────── minute (0 - 59) \n" +
	"\t│ ┌───────────── hour (0 - 23) \n" +
	"\t│ │ ┌───────────── day of the month (1 - 31) \n" +
	"\t│ │ │ ┌───────────── month (1 - 12) \n" +
	"\t│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) \n" +
	"\t│ │ │ │ │ \n" +
	"\t* * * * *"
	FlagWorkflowIdReusePolicyDefinition = "Allows the same Workflow Id to be used in a new Workflow Execution. Options are: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning."
	FlagInputDefinition = "Optional JSON input to provide to the Workflow. Pass \"null\" for null values."
	FlagInputFileDefinition = "Passes optional input for the Workflow from a JSON file. If there are multiple JSON files, concatenate them and separate by space or newline. Input from the command line will overwrite file input."
	FlagSearchAttributeDefinition = "Passes Search Attribute in key=value format. Use valid JSON formats for value."
	FlagMemoDefinition = "Passes a memo in key=value format. Use valid JSON formats for value."
	FlagMemoFileDefinition = "Passes a memo as file input, with each line following key=value format. Use valid JSON formats for value."

	// Other Workflow flags
	FlagResetPointsUsage = "Only show auto-reset points."
	FlagPrintRawUsage = "Print properties without changing their format."
	QueryFlagTypeUsage = "The type of Query to run."
	FlagWorkflowSignalUsage = "Signal Workflow Execution by Id."
	FlagQueryDefinition = "Signal Workflow Executions by List Filter."
	FlagSignalName = "Signal Name"
	FlagInputSignal = "Input for the Signal. Formatted in JSON."
	FlagInputFileSignal = "Input for the Signal from file. Formatted in JSON."
	FlagCancelWorkflow = "Cancel Workflow Execution with given Workflow Id."
	FlagWorkflowIDTerminate = "Terminate Workflow Execution with given Workflow Id."
	FlagQueryTerminate = "Terminate Workflow Executions with given List Filter."
	FlagEventIDDefinition = "The Event Id for any Event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others."
	FlagQueryResetBatch = "Visibility Query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query."
	FlagInputFileReset = "Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id."
	FlagExcludeFileDefinition = "Input file that specifies Workflow Executions to exclude from a reset."
	FlagInputSeparatorDefinition = "Separator for the input file. The default is a tab (`\t`)."
	FlagParallelismDefinition = "Number of goroutines to run in parallel. Each goroutine processes one line per second."
	FlagSkipCurrentOpenDefinition = "Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run."
	FlagSkipBaseDefinition =  "Skip a Workflow Execution if the base Run is not the current Workflow Run."
	FlagNonDeterministicDefinition = "Reset Workflow Execution if its last Event is `WorkflowTaskFailed` with a nondeterministic error."
	FlagDryRunDefinition = "Simulate reset without resetting any Workflow Executions."
	FlagDepthDefinition = "Number of Child Workflows to expand. Use -1 to expand all Child Workflows."
	FlagConcurrencyDefinition = "Request concurrency."
	FlagNoFoldDefinition = "Disable folding. All Child Workflows within the set depth will be fetched and displayed."

	// Stack trace query flag definitions
	FlagQueryRejectConditionDefinition = "Optional flag for rejecting Queries based on Workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"."

	// Pagination flag definitions
	FlagLimitDefinition = "Number of items to print."
	FlagPagerDefinition = "Sets the pager for Temporal CLI to use. Options are less, more, and favoritePager."
	FlagNoPagerDefinition = "Disables the interactive pager."
	FlagFieldsDefinition = "Customize fields to print. Set to 'long' to automatically print more information for main fields."

	// Activity flag definitions
	FlagWorkflowIDDefinition = "Identifies the Workflow that the Activity is running on."
	FlagRunIDDefinition = "Identifies the current Workflow Run."
	FlagActivityIDDefinition = "Identifies the Activity Execution."
	FlagResultDefinition = "Set the result value of Activity completion."
	FlagIdentityDefinition = "Specify operator's identity."

	FlagDetailDefinition = "Reason to fail the Activity."

	FlagReasonDefinition = "Reason to perform an operation on the Cluster."

	// Cluster flag definition
	FlagClusterAddressDefinition = "Frontend address of the remote Cluster."
	FlagClusterEnableConnectionDefinition = "Enable cross-cluster connection."
	FlagNameDefinition = "Frontend address of the remote Cluster."

	// Schedule flag definition
	FlagOverlapPolicyDefinition = "Overlap policy. Options are Skip, BufferOne, BufferAll, CancelOther, TerminateOther and AllowAll."
	FlagCalenderDefinition = `Calendar specification in JSON ({"dayOfWeek":"Fri","hour":"17","minute":"5"}) or as a Cron string ("30 2 * * 5" or "@daily").`
	FlagIntervalDefinition = "Interval duration, e.g. 90m, or 90m/13m to include phase offset."
	FlagStartTimeDefinition = "Overall schedule start time."
	FlagEndTimeDefinition = "Overall schedule end time."
	FlagJitterDefinition = "Jitter duration."
	FlagTimeZoneDefinition = "Time zone (IANA name)."
	FlagNotesDefinition = "Initial value of notes field."
	FlagPauseDefinition = "Initial value of paused state."
	FlagRemainingActionsDefinition = "Total number of actions allowed."
	FlagCatchupWindowDefinition = "Maximum allowed catch-up time if server is down."
	FlagPauseOnFailureDefinition = "Pause schedule after any workflow failure."
	FlagSearchAttributeScheduleDefinition = "Set Search Attribute on a schedule (formatted as key=value). Use valid JSON formats for value."
	FlagMemoScheduleDefinition = "Set a memo on a schedule (formatted as key=value). Use valid JSON formats for value."
	FlagMemoFileScheduleDefinition = "Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value."
	FlagPauseScheduleDefinition = "Pauses the Schedule."
	FlagUnpauseDefinition = "Unpauses the Schedule."
	FlagBackfillStartTime = "Backfill start time."
	FlagBackfillEndTime = "Backfill end time."
	FlagPrintRawDefinition = "Print raw data in JSON format. Recommended to use this over -o json for scripting."

	// Search Attribute flags
	FlagNameSearchAttribute = "Search Attribute name."
	FlagYesDefinition = "Confirm all prompts."

	// Task Queue flags
	FlagTaskQueueName = "Name of the Task Queue."
	FlagTaskQueueTypeDefinition = "Task Queue type [workflow|activity]"
)
