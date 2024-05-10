package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/sdk/client"
)

type assignmentRowType struct {
	Index         int       `json:"-"` // omit index with JSON
	TargetBuildID string    `json:"targetBuildID"`
	Percentage    float32   `json:"percentage"`
	CreateTime    time.Time `json:"createTime"`
}

type redirectRowType struct {
	SourceBuildID string    `json:"sourceBuildID"`
	TargetBuildID string    `json:"targetBuildID"`
	CreateTime    time.Time `json:"createTime"`
}

type formattedRulesType struct {
	AssignmentRules []assignmentRowType `json:"assignmentRules"`
	RedirectRules   []redirectRowType   `json:"redirectRules"`
}

func versioningRulesToRows(rules *client.WorkerVersioningRules) *formattedRulesType {
	var aRules []assignmentRowType
	for i, r := range rules.AssignmentRules {
		var percentage float32 = 100.0
		switch ramp := r.Rule.Ramp.(type) {
		case *client.VersioningRampByPercentage:
			percentage = ramp.Percentage
		}
		row := assignmentRowType{
			Index:         i,
			TargetBuildID: r.Rule.TargetBuildID,
			Percentage:    percentage,
			CreateTime:    r.CreateTime,
		}
		aRules = append(aRules, row)
	}

	var rRules []redirectRowType
	for _, r := range rules.RedirectRules {
		row := redirectRowType{
			SourceBuildID: r.Rule.SourceBuildID,
			TargetBuildID: r.Rule.TargetBuildID,
			CreateTime:    r.CreateTime,
		}
		rRules = append(rRules, row)
	}

	return &formattedRulesType{
		AssignmentRules: aRules,
		RedirectRules:   rRules,
	}
}

func printBuildIdRules(cctx *CommandContext, rules *client.WorkerVersioningRules) error {
	fRules := versioningRulesToRows(rules)

	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString("Assignment Rules:"))
		err := cctx.Printer.PrintStructured(fRules.AssignmentRules, printer.StructuredOptions{Table: &printer.TableOptions{}})
		if err != nil {
			return fmt.Errorf("displaying rules failed: %w", err)
		}
		// Separate newline
		cctx.Printer.Println()
		cctx.Printer.Println(color.MagentaString("Redirection Rules:"))
		return cctx.Printer.PrintStructured(fRules.RedirectRules, printer.StructuredOptions{Table: &printer.TableOptions{}})
	}
	// json output
	return cctx.Printer.PrintStructured(fRules, printer.StructuredOptions{})
}

func (c *TemporalTaskQueueGetBuildIdRulesCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()

	rules, err := cl.GetWorkerVersioningRules(cctx, &client.GetWorkerVersioningOptions{
		TaskQueue: c.TaskQueue,
	})
	if err != nil {
		return fmt.Errorf("unable to get task queue build ID rules: %w", err)
	}

	return printBuildIdRules(cctx, rules)
}
