package common

const (
	// Shared flag definitions
	FlagEnvDefinition            = "Environment to read environmental variables from."
	FlagAddrDefinition           = "The host and port (formatted as host:port) for the Temporal Frontend Service."
	FlagNSAliasDefinition        = "Identifies a Namespace in the Temporal Workflow."
	FlagMetadataDefinition       = "Contains gRPC metadata to send with requests (formatted as key=value)."
	FlagTLSDefinition            = "Enable TLS encryption without additional options such as mTLS or client certificates."
	FlagTLSCertPathDefinition    = "Path to x509 certificate."
	FlagTLSKeyPathDefinition     = "Path to private certificate key."
	FlagTLSCaPathDefinition      = "Path to server CA certificate."
	FlagTLSDisableHVDefinition   = "Disables TLS host-name verification."
	FlagTLSServerNameDefinition  = "Overrides target TLS server name."
	FlagContextTimeoutDefinition = "Optional timeout for the context of an RPC call (in seconds)."
	FlagCodecEndpointDefinition  = "Endpoint for a remote Codec Server."
	FlagCodecAuthDefinition      = "Sets the authorization header on requests to the Codec Server."
	FlagArchiveDefinition        = "List archived Workflow Executions. Currently an experimental feature."

	// Execution flags
	FlagWorkflowId           = "Workflow Id"
	FlagRunIdDefinition      = "Run Id"
	FlagJobIDDefinition      = "Batch Job Id"
	FlagScheduleIDDefinition = "Schedule Id"

	// ShowWorkflow flags
	FlagOutputFilenameDefinition  = "Serializes Event History to a file."
	FlagMaxFieldLengthDefinition  = "Maximum length for each attribute field."
	FlagResetPointsOnlyDefinition = "Only show Workflow Events that are eligible for reset."
	FlagFollowAliasDefinition     = "Follow the progress of a Workflow Execution."

	// StartWorkflow flags
	FlagWFTypeDefinition                   = "Workflow Type name."
	FlagTaskQueueDefinition                = "Task Queue"
	FlagWorkflowRunTimeoutDefinition       = "Timeout (in seconds) of a Workflow Run."
	FlagWorkflowExecutionTimeoutDefinition = "Timeout (in seconds) for a WorkflowExecution, including retries and `ContinueAsNew` tasks."
	FlagWorkflowTaskTimeoutDefinition      = "Start-to-close timeout for a Workflow Task (in seconds)."
	FlagCronScheduleDefinition             = "Optional Cron Schedule for the Workflow. Cron spec is formatted as: \n" +
		"\t┌───────────── minute (0 - 59) \n" +
		"\t│ ┌───────────── hour (0 - 23) \n" +
		"\t│ │ ┌───────────── day of the month (1 - 31) \n" +
		"\t│ │ │ ┌───────────── month (1 - 12) \n" +
		"\t│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) \n" +
		"\t│ │ │ │ │ \n" +
		"\t* * * * *"
	FlagWorkflowIdReusePolicyDefinition = "Allows the same Workflow Id to be used in a new Workflow Execution. Options are: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning."
	FlagInputDefinition                 = "JSON value to provide to the Workflow or Query. You may use --input multiple times to pass multiple arguments. May not be combined with --input-file."
	FlagInputFileDefinition             = "Reads a JSON file and provides the JSON as input to the Workflow or Query. The file must contain a single JSON value (typically an object). Each file is passed as a separate argument; you may use --input-file multiple times to pass multiple arguments. May not be combined with --input."
	FlagSearchAttributeDefinition       = "Passes Search Attribute in key=value format. Use valid JSON formats for value."
	FlagMemoDefinition                  = "Passes a memo in key=value format. Use valid JSON formats for value."
	FlagMemoFileDefinition              = "Passes a memo as file input, with each line following key=value format. Use valid JSON formats for value."

	// Other Workflow flags
	FlagResetPointsUsage           = "Only show auto-reset points."
	FlagPrintRawUsage              = "Print properties without changing their format."
	QueryFlagTypeUsage             = "The type of Query to run."
	FlagWorkflowSignalUsage        = "Signal Workflow Execution by Id."
	FlagSignalName                 = "Signal Name"
	FlagInputSignal                = "Input for the Signal. Formatted in JSON."
	FlagInputFileSignal            = "Input for the Signal from file. Formatted in JSON."
	FlagUpdateHandlerName          = "Update handler Name"
	FlagUpdateHandlerInput         = "Args for the Update handler. Formatted in JSON."
	FlagUpdateIDDefinition         = "UpdateID to check the result of an update (either UpdateID or Update handler name should be passed)"
	FlagCancelWorkflow             = "Cancel Workflow Execution with given Workflow Id."
	FlagWorkflowIDTerminate        = "Terminate Workflow Execution with given Workflow Id."
	FlagQueryCancel                = "Cancel Workflow Executions with given List Filter."
	FlagQueryDelete                = "Delete Workflow Executions with given List Filter."
	FlagQuerySignal                = "Signal Workflow Executions with given List List Filter."
	FlagQueryTerminate             = "Terminate Workflow Executions with given List Filter."
	FlagEventIDDefinition          = "The Event Id for any Event after WorkflowTaskStarted you want to reset to (exclusive). It can be WorkflowTaskCompleted, WorkflowTaskFailed or others."
	FlagQueryResetBatch            = "Visibility Query of Search Attributes describing the Workflow Executions to reset. See https://docs.temporal.io/docs/tctl/workflow/list#--query."
	FlagInputFileReset             = "Input file that specifies Workflow Executions to reset. Each line contains one Workflow Id as the base Run and, optionally, a Run Id."
	FlagExcludeFileDefinition      = "Input file that specifies Workflow Executions to exclude from a reset."
	FlagInputSeparatorDefinition   = "Separator for the input file. The default is a tab (`\t`)."
	FlagParallelismDefinition      = "Number of goroutines to run in parallel. Each goroutine processes one line per second."
	FlagSkipCurrentOpenDefinition  = "Skip a Workflow Execution if the current Run is open for the same Workflow Id as the base Run."
	FlagSkipBaseDefinition         = "Skip a Workflow Execution if the base Run is not the current Workflow Run."
	FlagNonDeterministicDefinition = "Reset Workflow Execution if its last Event is `WorkflowTaskFailed` with a nondeterministic error."
	FlagDryRunDefinition           = "Simulate reset without resetting any Workflow Executions."
	FlagDepthDefinition            = "Depth of child workflows to fetch. Use -1 to fetch child workflows at any depth."
	FlagConcurrencyDefinition      = "Request concurrency."
	FlagNoFoldDefinition           = "Disable folding. All Child Workflows within the set depth will be fetched and displayed."

	// Stack trace query flag definitions
	FlagQueryRejectConditionDefinition = "Optional flag for rejecting Queries based on Workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"."

	// Pagination flag definitions
	FlagLimitDefinition   = "Number of items to print."
	FlagPagerDefinition   = "Sets the pager for Temporal CLI to use. Options are less, more, and favoritePager."
	FlagNoPagerDefinition = "Disables the interactive pager."
	FlagFieldsDefinition  = "Customize fields to print. Set to 'long' to automatically print more information for main fields."

	// Activity flag definitions
	FlagWorkflowIDDefinition = "Identifies the Workflow that the Activity is running on."
	FlagRunIDDefinition      = "Identifies the current Workflow Run."
	FlagActivityIDDefinition = "Identifies the Activity Execution."
	FlagResultDefinition     = "Set the result value of Activity completion."
	FlagIdentityDefinition   = "Specify operator's identity."

	FlagDetailDefinition = "Reason to fail the Activity."

	FlagReasonDefinition = "Reason to perform an operation on the Cluster."

	// Cluster flag definition
	FlagClusterAddressDefinition          = "Frontend address of the remote Cluster."
	FlagClusterEnableConnectionDefinition = "Enable cross-cluster connection."
	FlagNameDefinition                    = "Frontend address of the remote Cluster."

	// Schedule flag definition
	FlagOverlapPolicyDefinition           = "Overlap policy. Options are Skip, BufferOne, BufferAll, CancelOther, TerminateOther and AllowAll."
	FlagCalenderDefinition                = `Calendar specification in JSON ({"dayOfWeek":"Fri","hour":"17","minute":"5"}) or as a Cron string ("30 2 * * 5" or "@daily").`
	FlagIntervalDefinition                = "Interval duration, e.g. 90m, or 90m/13m to include phase offset."
	FlagStartTimeDefinition               = "Overall schedule start time."
	FlagEndTimeDefinition                 = "Overall schedule end time."
	FlagJitterDefinition                  = "Jitter duration."
	FlagTimeZoneDefinition                = "Time zone (IANA name)."
	FlagNotesDefinition                   = "Initial value of notes field."
	FlagPauseDefinition                   = "Initial value of paused state."
	FlagRemainingActionsDefinition        = "Total number of actions allowed."
	FlagCatchupWindowDefinition           = "Maximum allowed catch-up time if server is down."
	FlagPauseOnFailureDefinition          = "Pause schedule after any workflow failure."
	FlagSearchAttributeScheduleDefinition = "Set Search Attribute on a schedule (formatted as key=value). Use valid JSON formats for value."
	FlagMemoScheduleDefinition            = "Set a memo on a schedule (formatted as key=value). Use valid JSON formats for value."
	FlagMemoFileScheduleDefinition        = "Set a memo from a file. Each line should follow the format key=value. Use valid JSON formats for value."
	FlagPauseScheduleDefinition           = "Pauses the Schedule."
	FlagUnpauseDefinition                 = "Unpauses the Schedule."
	FlagBackfillStartTime                 = "Backfill start time."
	FlagBackfillEndTime                   = "Backfill end time."
	FlagPrintRawDefinition                = "Print raw data in JSON format. Recommended to use this over -o json for scripting."

	// Search Attribute flags
	FlagNameSearchAttribute = "Search Attribute name."
	FlagYesDefinition       = "Confirm all prompts."

	// Task Queue flags
	FlagTaskQueueName           = "Name of the Task Queue."
	FlagTaskQueueTypeDefinition = "Task Queue type [workflow|activity]"
	FlagPartitionsDefinition    = "Query for all partitions up to this number (experimental+temporary feature)"

	// Namespace update flags
	FlagActiveClusterDefinition           = "Active cluster name."
	FlagClusterDefinition                 = "Cluster name."
	FlagDescriptionDefinition             = "Namespace description."
	FlagHistoryArchivalStateDefinition    = "History archival state, valid values are \"disabled\" and \"enabled\""
	FlagHistoryArchivalURIDefinition      = "Optionally specify history archival URI (cannot be changed after first time archival is enabled)"
	FlagIsGlobalNamespaceDefinition       = "Whether the namespace is a global namespace."
	FlagNamespaceDataDefinition           = "Namespace data in key=value format. Use JSON for values."
	FlagNamespaceVerboseDefinition        = "Print applied namespace changes"
	FlagOwnerDefinition                   = "Owner email."
	FlagPromoteNamespaceDefinition        = "Promote local namespace to global namespace"
	FlagRetentionDefinition               = "Length of time (in days) a closed Workflow is preserved before deletion."
	FlagVisibilityArchivalStateDefinition = "Visibility archival state, valid values are \"disabled\" and \"enabled\""
	FlagVisibilityArchivalURIDefinition   = "Optionally specify visibility archival URI (cannot be changed after first time archival is enabled)"

	// Build id based versioning flags
	FlagNewBuildIDUsage                = "The new build id to be added."
	FlagExistingCompatibleBuildIDUsage = "A build id which must already exist in the version sets known by the task queue. The new id will be stored in the set containing this id, marking it as compatible with the versions within."
	FlagSetBuildIDAsDefaultUsage       = "When set, establishes the compatible set being targeted as the overall default for the queue. If a different set was the current default, the targeted set will replace it as the new default."
	FlagPromoteSetBuildIDUsage         = "An existing build id whose containing set will be promoted."
	FlagPromoteBuildIDUsage            = "An existing build id which will be promoted to be the default inside its containing set."
	FlagMaxBuildIDSetsUsage            = "Limits how many compatible sets will be returned. Specify 1 to only return the current default major version set. 0 returns all sets."
	FlagBuildIDReachabilityUsage       = "Which Build ID to get reachability information for. May be specified multiple times."
	FlagTaskQueueForReachabilityUsage  = "Which Task Queue(s) to constrain the reachability search to. May be specified multiple times."
	FlagReachabilityTypeUsage          = "Specify how you'd like to filter the reachability of Build IDs. Valid choices are `open` (reachable by one or more open workflows), `closed` (reachable by one or more closed workflows), or `existing` (reachable by either). If a Build ID is reachable by new workflows, that is always reported."
)
