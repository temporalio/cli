package common

import (
	"time"

	enumspb "go.temporal.io/api/enums/v1"
)

const (
	LocalHostPort = "127.0.0.1:7233"

	maxOutputStringLength = 200 // max length for output string
	maxWorkflowTypeLength = 32  // max item length for output workflow type in table
	defaultMaxFieldLength = 500 // default max length for each attribute field

	// regex expression for parsing time durations, shorter, longer notations and numeric value respectively
	defaultDateTimeRangeShortRE = "^[1-9][0-9]*[smhdwMy]$"                                // eg. 1s, 20m, 300h etc.
	defaultDateTimeRangeLongRE  = "^[1-9][0-9]*(second|minute|hour|day|week|month|year)$" // eg. 1second, 20minute, 300hour etc.
	defaultDateTimeRangeNum     = "^[1-9][0-9]*"                                          // eg. 1, 20, 300 etc.

	// time ranges
	day   = 24 * time.Hour
	week  = 7 * day
	month = 30 * day
	year  = 365 * day

	defaultContextTimeout                        = defaultContextTimeoutInSeconds * time.Second
	DefaultContextTimeoutForListArchivedWorkflow = 3 * time.Minute
	defaultContextTimeoutForLongPoll             = 2 * time.Minute
	defaultContextTimeoutInSeconds               = 5

	defaultDateTimeFormat               = time.RFC3339 // used for converting UnixNano to string like 2018-02-15T16:16:36-08:00
	DefaultNamespaceRetention           = 3 * 24 * time.Hour
	defaultTimeFormat                   = "15:04:05" // used for converting UnixNano to string like 16:16:36 (only time)
	DefaultWorkflowIDReusePolicy        = enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
	defaultWorkflowTaskTimeoutInSeconds = 10

	ShowErrorStackEnv = `TEMPORAL_CLI_SHOW_STACKS`
)

var envKeysForUserName = []string{
	"USER",
	"LOGNAME",
	"HOME",
}

