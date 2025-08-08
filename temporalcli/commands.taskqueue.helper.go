package temporalcli

import (
	"fmt"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
)

// Create a structured table for config display.
type configRow struct {
	Setting     string
	Value       string
	Reason      string
	UpdatedBy   string
	UpdatedTime string
}

func parseTaskQueueType(input string) (enums.TaskQueueType, error) {
	switch input {
	case "", "workflow":
		return enums.TASK_QUEUE_TYPE_WORKFLOW, nil
	case "activity":
		return enums.TASK_QUEUE_TYPE_ACTIVITY, nil
	case "nexus":
		return enums.TASK_QUEUE_TYPE_NEXUS, nil
	default:
		return enums.TASK_QUEUE_TYPE_WORKFLOW, fmt.Errorf(
			"invalid task queue type: %s. Must be one of: workflow, activity, nexus", input)
	}
}

func buildRateLimitConfigRow(setting string, rl *taskqueue.RateLimitConfig, format string) configRow {
	value := "Not Set"
	reason := ""
	updatedBy := ""
	updatedTime := ""

	if rl.RateLimit != nil && rl.RateLimit.RequestsPerSecond > 0 {
		value = fmt.Sprintf(format, rl.RateLimit.RequestsPerSecond)
	}

	if rl.Metadata != nil {
		if rl.Metadata.Reason != "" {
			reason = truncateString(rl.Metadata.Reason, 50)
		}
		if rl.Metadata.UpdateIdentity != "" {
			updatedBy = truncateString(rl.Metadata.UpdateIdentity, 50)
		}
		if rl.Metadata.UpdateTime != nil {
			updateTime := rl.Metadata.UpdateTime.AsTime()
			updatedTime = updateTime.Format("2006-01-02 15:04:05")
		}
	}

	return configRow{
		Setting:     setting,
		Value:       value,
		Reason:      reason,
		UpdatedBy:   updatedBy,
		UpdatedTime: updatedTime,
	}
}

func buildRateLimitUpdate(
	rateLimitSet bool,
	rateLimit float32,
	reason string,
) *workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate {
	if reason == "" && !rateLimitSet {
		return nil
	}
	if rateLimit == UnsetRateLimit {
		return &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
			Reason: reason,
		}
	}
	return &workflowservice.UpdateTaskQueueConfigRequest_RateLimitUpdate{
		RateLimit: &taskqueue.RateLimit{
			RequestsPerSecond: rateLimit,
		},
		Reason: reason,
	}
}
