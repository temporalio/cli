package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/taskqueue/v1"
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
			updatedTime = updateTime.Format(time.RFC3339)
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

// printTaskQueueConfig is a shared function to print task queue configuration
// This can be used by both the config get command and the describe command
func printTaskQueueConfig(cctx *CommandContext, config *taskqueue.TaskQueueConfig) error {
	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(config, printer.StructuredOptions{})
	}

	// For text, we will use a table
	var configRows []configRow

	// Queue Rate Limit
	if config.QueueRateLimit != nil {
		configRows = append(configRows, buildRateLimitConfigRow("Queue Rate Limit", config.QueueRateLimit, "%.2f rps"))
	}

	// Fairness Key Rate Limit Default
	if config.FairnessKeysRateLimitDefault != nil {
		configRows = append(configRows, buildRateLimitConfigRow("Fairness Key Rate Limit Default", config.FairnessKeysRateLimitDefault, "%.2f rps"))
	}

	// Print the config table
	if len(configRows) > 0 {
		// Always show truncation note, regardless of actual truncation
		cctx.Printer.Println(color.YellowString("Note: Long content may be truncated. Use --output json for full details."))

		return cctx.Printer.PrintStructured(configRows, printer.StructuredOptions{
			Table: &printer.TableOptions{},
		})
	}

	return nil
}

func printTaskQueueVersioningInfo(cctx *CommandContext, versioningInfo *taskqueue.TaskQueueVersioningInfo) error {
	if versioningInfo == nil {
		return nil
	}

	// For JSON, we'll just dump the proto
	if cctx.JSONOutput {
		return cctx.Printer.PrintStructured(versioningInfo, printer.StructuredOptions{})
	}

	// For text, we will use a structured display
	var currentVersionDeploymentName, currentVersionBuildID string
	currentVersionDeploymentName = versioningInfo.GetCurrentDeploymentVersion().GetDeploymentName()
	currentVersionBuildID = versioningInfo.GetCurrentDeploymentVersion().GetBuildId()

	var rampingVersionDeploymentName, rampingVersionBuildID string
	if versioningInfo.RampingDeploymentVersion != nil {
		rampingVersionDeploymentName = versioningInfo.RampingDeploymentVersion.DeploymentName
		rampingVersionBuildID = versioningInfo.RampingDeploymentVersion.BuildId
	}

	var updateTime time.Time
	if versioningInfo.UpdateTime != nil {
		updateTime = versioningInfo.UpdateTime.AsTime()
	}

	printMe := struct {
		CurrentVersionDeploymentName string    `cli:",cardOmitEmpty"`
		CurrentVersionBuildID        string    `cli:",cardOmitEmpty"`
		RampingVersionDeploymentName string    `cli:",cardOmitEmpty"`
		RampingVersionBuildID        string    `cli:",cardOmitEmpty"`
		RampingVersionPercentage     float32   `cli:",cardOmitEmpty"`
		UpdateTime                   time.Time `cli:",cardOmitEmpty"`
	}{
		CurrentVersionDeploymentName: currentVersionDeploymentName,
		CurrentVersionBuildID:        currentVersionBuildID,
		RampingVersionDeploymentName: rampingVersionDeploymentName,
		RampingVersionBuildID:        rampingVersionBuildID,
		RampingVersionPercentage:     versioningInfo.RampingVersionPercentage,
		UpdateTime:                   updateTime,
	}

	return cctx.Printer.PrintStructured(printMe, printer.StructuredOptions{})
}
