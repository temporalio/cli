package common
const (
	// Shared flag definitions
	FlagEnvDefinition = "Name of the environment to read environmental variables from."
	FlagAddrDefinition = "The host and port (formatted as host:port) for the Temporal Frontend Service."
	FlagNSAliasDefinition = "Identifies a Namespace in the Temporal Workflow."
	FlagMetadataDefinition = "Contains gRPC metadata to send with requests (Format: key=value). Values must be in a valid JSON format."
	FlagTLSCertPathDefinition = "Path to x509 certificate."
	FlagTLSKeyPathDefinition = "Path to private certificate key."
	FlagTLSCaPathDefinition = "Path to server CA certificate."
	FlagTLSDisableHVDefinition = "Disables TLS host name verification if already enabled."
	FlagTLSServerNameDefinition = "Provides an override for the target TLS server name."
	FlagContextTimeoutDefinition = "An optional timeout for the context of an RPC call (in seconds)."
	FlagCodecEndpointDefinition = "Endpoint for a remote Codec Server."
	FlagCodecAuthDefinition = "Sets the authorization header on requests to the Codec Server."

	// Execution flags
	FlagWorkflowIdDefinition = "Workflow Id"
	FlagRunIdDefinition = "Run Id"

	// ShowWorkflow flags
	FlagOutputFilenameDefinition = "Serializes Event History to a file."
	FlagMaxFieldLengthDefinition = "Maximum length for each attribute field."
	FlagResetPointsOnlyDefinition = "Only show Workflow Events that are eligible for reset."
	FlagFollowAliasDefinition = "Follow the progress of a Workflow Execution."

	// StartWorkflow flags
	FlagWFTypeDefinition = "Workflow type name."
	FlagTaskQueueDefinition = "Task Queue"
	FlagWorkflowRunTimeoutDefinition = "Timeout (in seconds) of a single Workflow run."
	FlagWorkflowExecutionTimeoutDefinition = "Timeout (in seconds) for a WorkflowExecution, including retries and continue-as-new tasks."
	FlagWorkflowTaskTimeoutDefinition = "Start-to-close timeout for a Workflow Task (in seconds)."
	FlagCronScheduleDefinition = "Optional Cron Schedule for the Workflow. Cron spec is formatted as: \n" +
	"\t┌───────────── minute (0 - 59) \n" +
	"\t│ ┌───────────── hour (0 - 23) \n" +
	"\t│ │ ┌───────────── day of the month (1 - 31) \n" +
	"\t│ │ │ ┌───────────── month (1 - 12) \n" +
	"\t│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday) \n" +
	"\t│ │ │ │ │ \n" +
	"\t* * * * *"
	FlagWorkflowIdReusePolicyDefinition = "Allows the same Workflow Id to be used in a new Workflow Execution. " +
	"Options: AllowDuplicate, AllowDuplicateFailedOnly, RejectDuplicate, TerminateIfRunning."
	FlagInputDefinition = "Optional JSON input to provide to the Workflow.\nPass \"null\" for null values."
	FlagInputFileDefinition = "Passes optional input for the Workflow from a JSON file.\n" +
	"If there are multiple JSON files, concatenate them and separate by space or newline.\n" +
	"Input from the command line will overwrite file input."
	FlagSearchAttributeDefinition = "Passes Search Attribute in key=value format. Use valid JSON formats for value."
	FlagMemoDefinition = "Passes a memo in key=value format. Use valid JSON formats for value."
	FlagMemoFileDefinition = "Passes a memo as file input, with each line following key=value format. Use valid JSON formats for value."

	// Stack trace query flag definitions
	FlagInputSTQDefinition = "Optional query input, in JSON format. For multiple parameters, concatenate them and separate by space."
	FlagInputFileSTQDefinition = "Passes optional Query input from a JSON file.\nIf there are multiple JSON, concatenate them and separate by space or newline.\n" + "Input from the command line will overwrite file input."
	FlagQueryRejectConditionDefinition = "Optional flag for rejecting Queries based on Workflow state. Valid values are \"not_open\" and \"not_completed_cleanly\"."

	// Pagination flag definitions
	FlagLimitDefinition = "Number of items to print."
	FlagPagerDefinition = "Sets the pager for Temporal CLI to use.\nOptions: less, more, favoritePager."
	FlagNoPagerDefinition = "Disables the interactive pager."
)