package tracer

import (
	"github.com/fatih/color"
	"go.temporal.io/api/enums/v1"
)

var (
	StatusRunning              string = color.BlueString("▷")
	StatusCompleted            string = color.GreenString("✓")
	StatusTerminated           string = color.RedString("x")
	StatusCanceled             string = color.YellowString("x")
	StatusFailed               string = color.RedString("!")
	StatusContinueAsNew        string = color.GreenString("»")
	StatusTimedOut             string = color.RedString("⏱")
	StatusUnspecifiedScheduled string = "•"
	StatusCancelRequested      string = color.YellowString("▷")
	StatusTimerWaiting         string = color.BlueString("⧖")
	StatusTimerFired           string = color.GreenString("⧖")
	StatusTimerCanceled        string = color.YellowString("⧖")
)

var workflowIcons = map[enums.WorkflowExecutionStatus]string{
	enums.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED:      StatusUnspecifiedScheduled,
	enums.WORKFLOW_EXECUTION_STATUS_RUNNING:          StatusRunning,
	enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:        StatusCompleted,
	enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:       StatusTerminated,
	enums.WORKFLOW_EXECUTION_STATUS_CANCELED:         StatusCanceled,
	enums.WORKFLOW_EXECUTION_STATUS_FAILED:           StatusFailed,
	enums.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW: StatusContinueAsNew,
	enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:        StatusTimedOut,
}
var activityIcons = map[ActivityExecutionStatus]string{
	ACTIVITY_EXECUTION_STATUS_UNSPECIFIED:      StatusUnspecifiedScheduled,
	ACTIVITY_EXECUTION_STATUS_SCHEDULED:        StatusUnspecifiedScheduled,
	ACTIVITY_EXECUTION_STATUS_RUNNING:          StatusRunning,
	ACTIVITY_EXECUTION_STATUS_COMPLETED:        StatusCompleted,
	ACTIVITY_EXECUTION_STATUS_CANCEL_REQUESTED: StatusCancelRequested,
	ACTIVITY_EXECUTION_STATUS_CANCELED:         StatusCanceled,
	ACTIVITY_EXECUTION_STATUS_FAILED:           StatusFailed,
	ACTIVITY_EXECUTION_STATUS_TIMED_OUT:        StatusTimedOut,
}
var timerIcons = map[TimerExecutionStatus]string{
	TIMER_STATUS_WAITING:  StatusTimerWaiting,
	TIMER_STATUS_FIRED:    StatusTimerFired,
	TIMER_STATUS_CANCELED: StatusTimerCanceled,
}

// ExecutionStatus returns the icon (with color) for a given ExecutionState's status.
func ExecutionStatus(exec ExecutionState) string {
	switch e := exec.(type) {
	case *WorkflowExecutionState:
		if icon, ok := workflowIcons[e.Status]; ok {
			return icon
		}
	case *ActivityExecutionState:
		if icon, ok := activityIcons[e.Status]; ok {
			return icon
		}
	case *TimerExecutionState:
		if icon, ok := timerIcons[e.Status]; ok {
			return icon
		}
	}
	return "?"
}

// StatusIcon has names for each status (useful for help messages).
type StatusIcon struct {
	Name string
	Icon string
}

var StatusIconsLegend = []StatusIcon{
	{
		Name: "Unspecified or Scheduled", Icon: StatusUnspecifiedScheduled,
	},
	{
		Name: "Running", Icon: StatusRunning,
	},
	{
		Name: "Completed", Icon: StatusCompleted,
	},
	{
		Name: "Continue As New", Icon: StatusContinueAsNew,
	},
	{
		Name: "Failed", Icon: StatusFailed,
	},
	{
		Name: "Timed Out", Icon: StatusTimedOut,
	},
	{
		Name: "Cancel Requested", Icon: StatusCancelRequested,
	},
	{
		Name: "Canceled", Icon: StatusCanceled,
	},
	{
		Name: "Terminated", Icon: StatusTerminated,
	},
}
