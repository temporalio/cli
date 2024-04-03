package trace

import (
	"embed"
	"fmt"
	"io"
	"strings"
	"text/template"
	"time"

	"github.com/fatih/color"
	"go.temporal.io/api/enums/v1"
)

const (
	MinFoldingDepth = 1
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

var (
	faint = color.New(color.Faint).SprintFunc()
)

// ExecutionTemplate contains the necessary templates and utilities to render WorkflowExecutionState and its child states.
type ExecutionTemplate struct {
	tmpl       *template.Template
	shouldFold func(*WorkflowExecutionState, int) bool
}

// NewExecutionTemplate initializes the templates with the necessary functions.
func NewExecutionTemplate(foldStatus []enums.WorkflowExecutionStatus, noFold bool) (*ExecutionTemplate, error) {
	shouldFold := ShouldFoldStatus(foldStatus, noFold)
	templateFunctions := template.FuncMap{
		"statusIcon": ExecutionStatus,
		"blue":       color.BlueString,
		"yellow":     color.YellowString,
		"red":        color.RedString,
		"faint":      faint,
		"shouldFold": shouldFold,
		"indent": func(depths ...int) string {
			var sum int
			for _, d := range depths {
				sum += d
			}
			return faint(strings.Repeat(" │  ", sum))
		},
		"splitLines": func(str string) []string {
			return strings.Split(str, "\n")
		},
		"timeSince": FmtTimeSince,
	}

	tmpl, err := template.New("output").Funcs(templateFunctions).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		return nil, err
	}

	return &ExecutionTemplate{
		tmpl:       tmpl,
		shouldFold: shouldFold,
	}, nil
}

type StateTemplate struct {
	State ExecutionState
	Depth int
}

// Execute executes the templates for a given Execution state and writes it into the ExecutionTemplate's writer.
func (t *ExecutionTemplate) Execute(writer io.Writer, state ExecutionState, depth int) error {
	if state == nil {
		return nil
	}

	var templateName string
	switch state.(type) {
	case *WorkflowExecutionState:
		templateName = "workflow.tmpl"
	case *ActivityExecutionState:
		templateName = "activity.tmpl"
	case *TimerExecutionState:
		templateName = "timer.tmpl"
	default:
		return fmt.Errorf("no template available for %s", state)
	}

	if err := t.tmpl.ExecuteTemplate(writer, templateName, &StateTemplate{
		State: state,
		Depth: depth,
	}); err != nil {
		return err
	}

	if workflow, isWorkflow := state.(*WorkflowExecutionState); isWorkflow && !t.shouldFold(workflow, depth) {
		for _, child := range workflow.ChildStates {
			if err := t.Execute(writer, child, depth+1); err != nil {
				return err
			}

		}
	}

	return nil
}

// ShouldFoldStatus returns a predicate that will return true when the workflow status can be folded for a given depth.
// NOTE: Depth starts at 0 (i.e. the root workflow will be at depth 0).
func ShouldFoldStatus(foldStatus []enums.WorkflowExecutionStatus, noFold bool) func(*WorkflowExecutionState, int) bool {
	return func(state *WorkflowExecutionState, currentDepth int) bool {
		if noFold || currentDepth < MinFoldingDepth {
			return false
		}
		for _, s := range foldStatus {
			if s == state.Status {
				return true
			}
		}
		return false
	}
}

// FmtTimeSince returns a string representing the difference it time between start and close (or start and now).
func FmtTimeSince(start time.Time, duration time.Duration) string {
	if start.IsZero() {
		return ""
	}
	if duration == 0 {
		return fmt.Sprintf("%s ago", FmtDuration(time.Since(start)))
	}
	return FmtDuration(duration)
}

// FmtDuration produces a string for a given duration, rounding to the most reasonable timeframe.
func FmtDuration(duration time.Duration) string {
	if duration < time.Second {
		return duration.Round(time.Millisecond).String()
	} else if duration < time.Hour {
		return duration.Round(time.Second).String()
	} else if duration < 24*time.Hour {
		return duration.Round(time.Minute).String()
	} else if duration < 7*24*time.Hour {
		days := int(duration.Hours() / 24)
		hours := int(duration.Hours()) - days*24 // % doesn't work for float
		return fmt.Sprintf("%dd%dh", days, hours)
	} else {
		days := int(duration.Hours() / 24)
		weeks := int(duration.Hours() / (7 * 24))
		return fmt.Sprintf("%dw%dd", weeks, days)
	}
}
