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
	FlagInputDefinition = "Optional JSON input to provide to the Workflow.\nPass Pass \"null\" for null values."

)