package temporalcli

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/internal/printer"
	"go.temporal.io/sdk/client"
)

type assignmentRowType struct {
	Index          int       `json:"-"` // omit index with JSON
	TargetBuildID  string    `json:"targetBuildID"`
	RampPercentage float32   `json:"rampPercentage"`
	CreateTime     time.Time `json:"createTime"`
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
			Index:          i,
			TargetBuildID:  r.Rule.TargetBuildID,
			RampPercentage: percentage,
			CreateTime:     r.CreateTime,
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

type getConflictTokenOptions struct {
	safeMode        bool
	safeModeMessage string
	taskQueue       string
	showAssignment  bool
}

func (c *TemporalTaskQueueVersioningCommand) getConflictToken(cctx *CommandContext, options *getConflictTokenOptions) (client.VersioningConflictToken, error) {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return client.VersioningConflictToken{}, err
	}
	defer cl.Close()

	rules, err := cl.GetWorkerVersioningRules(cctx, client.GetWorkerVersioningOptions{
		TaskQueue: options.taskQueue,
	})
	if err != nil {
		return client.VersioningConflictToken{}, fmt.Errorf("unable to get versioning conflict token: %w", err)
	}

	if options.safeMode {
		// duplicate `cctx.promptYes` check to avoid printing current rules with json
		if cctx.JSONOutput {
			return client.VersioningConflictToken{}, fmt.Errorf("must bypass prompts when using JSON output")
		}
		fRules := versioningRulesToRows(rules)

		if options.showAssignment {
			cctx.Printer.Println(color.MagentaString("Current Assignment Rules:"))
			err = cctx.Printer.PrintStructured(fRules.AssignmentRules, printer.StructuredOptions{Table: &printer.TableOptions{}})
		} else {
			//!showAssigment == showRedirect
			cctx.Printer.Println(color.MagentaString("Current Redirect Rules:"))
			err = cctx.Printer.PrintStructured(fRules.RedirectRules, printer.StructuredOptions{Table: &printer.TableOptions{}})
		}
		if err != nil {
			return client.VersioningConflictToken{}, fmt.Errorf("displaying rules failed: %w", err)
		}

		yes, err := cctx.promptYes(
			fmt.Sprintf("Continue with rules update %v? y/N", options.safeModeMessage), false)
		if err != nil {
			return client.VersioningConflictToken{}, err
		} else if !yes {
			return client.VersioningConflictToken{}, fmt.Errorf("user denied confirmation")
		}
	}

	return rules.ConflictToken, nil
}

func (c *TemporalTaskQueueVersioningCommand) updateBuildIdRules(cctx *CommandContext, options client.UpdateWorkerVersioningRulesOptions) error {
	cl, err := dialClient(cctx, &c.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	rules, err := cl.UpdateWorkerVersioningRules(cctx, options)
	if err != nil {
		return fmt.Errorf("error updating task queue build ID rules: %w", err)
	}

	err = printBuildIdRules(cctx, rules)
	if err != nil {
		return err
	}

	cctx.Printer.Println("Successfully updated task queue build ID rules")
	return nil
}

func (c *TemporalTaskQueueVersioningAddRedirectRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "adding a redirect rule",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  false,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationAddRedirectRule{
			Rule: client.VersioningRedirectRule{
				SourceBuildID: c.SourceBuildId,
				TargetBuildID: c.TargetBuildId,
			},
		},
	})
}

func (c *TemporalTaskQueueVersioningCommitBuildIdCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "committing a Build ID",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationCommitBuildID{
			TargetBuildID: c.BuildId,
			Force:         c.Force,
		},
	})
}

func (c *TemporalTaskQueueVersioningDeleteAssignmentRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "deleting an assignment rule",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationDeleteAssignmentRule{
			RuleIndex: int32(c.RuleIndex),
			Force:     c.Force,
		},
	})
}

func (c *TemporalTaskQueueVersioningDeleteRedirectRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "deleting a redirect rule",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  false,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationDeleteRedirectRule{
			SourceBuildID: c.SourceBuildId,
		},
	})
}

func (c *TemporalTaskQueueVersioningInsertAssignmentRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "inserting an assignment rule",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	rule := client.VersioningAssignmentRule{
		TargetBuildID: c.BuildId,
	}
	if c.Percentage != 100 {
		rule.Ramp = &client.VersioningRampByPercentage{
			Percentage: float32(c.Percentage),
		}
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationInsertAssignmentRule{
			RuleIndex: int32(c.RuleIndex),
			Rule:      rule,
		},
	})
}

func (c *TemporalTaskQueueVersioningReplaceAssignmentRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "replacing an assignment rule",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  true,
	})
	if err != nil {
		return err
	}

	rule := client.VersioningAssignmentRule{
		TargetBuildID: c.BuildId,
	}
	if c.Percentage != 100 {
		rule.Ramp = &client.VersioningRampByPercentage{
			Percentage: float32(c.Percentage),
		}
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationReplaceAssignmentRule{
			RuleIndex: int32(c.RuleIndex),
			Rule:      rule,
			Force:     c.Force,
		},
	})
}

func (c *TemporalTaskQueueVersioningReplaceRedirectRuleCommand) run(cctx *CommandContext, args []string) error {
	token, err := c.Parent.getConflictToken(cctx, &getConflictTokenOptions{
		safeMode:        !c.Yes,
		safeModeMessage: "replacing a redirect rule",
		taskQueue:       c.Parent.TaskQueue,
		showAssignment:  false,
	})
	if err != nil {
		return err
	}

	return c.Parent.updateBuildIdRules(cctx, client.UpdateWorkerVersioningRulesOptions{
		TaskQueue:     c.Parent.TaskQueue,
		ConflictToken: token,
		Operation: &client.VersioningOperationReplaceRedirectRule{
			Rule: client.VersioningRedirectRule{
				SourceBuildID: c.SourceBuildId,
				TargetBuildID: c.TargetBuildId,
			},
		},
	})
}

func (c *TemporalTaskQueueVersioningGetRulesCommand) run(cctx *CommandContext, args []string) error {
	cl, err := dialClient(cctx, &c.Parent.Parent.ClientOptions)
	if err != nil {
		return err
	}
	defer cl.Close()

	rules, err := cl.GetWorkerVersioningRules(cctx, client.GetWorkerVersioningOptions{
		TaskQueue: c.Parent.TaskQueue,
	})
	if err != nil {
		return fmt.Errorf("unable to get task queue build ID rules: %w", err)
	}

	return printBuildIdRules(cctx, rules)
}
