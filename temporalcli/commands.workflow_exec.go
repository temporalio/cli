package temporalcli

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
	stringify "github.com/temporalio/cli/temporalcli/internal"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/protobuf/encoding/protojson"
)

func (c *TemporalWorkflowExecuteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	run, err := c.Parent.startWorkflow(cctx, cl, &c.WorkflowStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}

	// Print history only if not JSON
	if !cctx.JSONOutput {
		iter := cl.GetWorkflowHistory(cctx, run.GetID(), run.GetRunID(), true, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				return fmt.Errorf("failed getting event: %w", err)
			}
			// Print the event
			fields := map[string]any{
				"ID": event.EventId,
				// We pre-format time here because the default --time-format makes no
				// sense if it's "relative"
				// TODO(cretz): Allow user to override?
				"Time": event.EventTime.AsTime().Format(time.RFC3339),
				"Type": coloredEventType(cctx, event.EventType),
			}
			if c.EventDetails {
				// First field in the attributes
				attrs := reflect.ValueOf(event.Attributes).Field(0).Interface()
				fields["Details"] = stringify.AnyToString(attrs, false, 120, dataConverter)
			}
			if err := cctx.Printer.Print(PrintOptions{Table: &PrintTableOptions{}}, fields); err != nil {
				return fmt.Errorf("failed printing: %w", err)
			}
			// We have to follow runs
			var newRunID string
			switch event.EventType {
			case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
				newRunID = event.GetWorkflowExecutionCompletedEventAttributes().NewExecutionRunId
			case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
				newRunID = event.GetWorkflowExecutionFailedEventAttributes().NewExecutionRunId
			case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
				newRunID = event.GetWorkflowExecutionTimedOutEventAttributes().NewExecutionRunId
			case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CONTINUED_AS_NEW:
				newRunID = event.GetWorkflowExecutionContinuedAsNewEventAttributes().NewExecutionRunId
			}
			if newRunID != "" {
				cctx.Logger.Info("Following workflow events to new run", "workflow.runId", newRunID)
				iter = cl.GetWorkflowHistory(cctx, run.GetID(), newRunID, true, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
			}
		}
	}

	// Get result
	var res any
	fail := run.Get(cctx, &res)

	// Convert to result pointer or failure point. If fail remains non-nil after
	// this, the command
	var resJSON, failJSON json.RawMessage
	// If error was not a workflow failure, assume bad data format and just
	// assume we can't deserialize
	if execFail, _ := fail.(*temporal.WorkflowExecutionError); execFail != nil {
		// Convert to JSON error representation
		failureProto := temporal.GetDefaultFailureConverter().ErrorToFailure(execFail.Unwrap())
		if failJSON, err = protojson.Marshal(failureProto); err != nil {
			fail = fmt.Errorf("<failed deserializing error: %w>", err)
		}
	} else if fail != nil {
		res = fmt.Sprintf("<failed converting: %v>", err)
		fail = run.Get(cctx, nil)
	}

	// Log and set result or failure JSON if not already set
	if fail != nil {
		cctx.Logger.Error("Failed running workflow",
			"workflow.id", run.GetID(), "workflow.runId", run.GetRunID(), "error", fail)
		if len(failJSON) == 0 {
			failJSON, _ = json.Marshal(err.Error())
		}
	} else {
		// We only log success, we letter outer error failure log the failure
		cctx.Logger.Info("Workflow completed successfully",
			"workflow.id", run.GetID(), "workflow.runId", run.GetRunID(), "result", res)
		// Try to convert to JSON
		if resJSON, err = json.Marshal(res); err != nil {
			resJSON, _ = json.Marshal(fmt.Sprintf("<failed converting: %v>", err))
		}
	}

	// Print completion
	err = cctx.Printer.Print(PrintOptions{}, struct {
		WorkflowId string          `json:"workflowId"`
		RunId      string          `json:"runId"`
		Type       string          `json:"type"`
		Namespace  string          `json:"namespace"`
		TaskQueue  string          `json:"taskQueue"`
		Result     json.RawMessage `json:"result,omitempty"`
		Failure    json.RawMessage `json:"failure,omitempty"`
	}{
		WorkflowId: run.GetID(),
		RunId:      run.GetRunID(),
		Type:       c.WorkflowStartOptions.Type,
		Namespace:  c.Parent.Namespace,
		TaskQueue:  c.WorkflowStartOptions.TaskQueue,
		Result:     resJSON,
		Failure:    failJSON,
	})
	// Fail if workflow run failed (intentionally swallowing `err` if needed)
	if fail != nil {
		return fail
	}
	return err
}

func (c *TemporalWorkflowStartCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	run, err := c.Parent.startWorkflow(cctx, cl, &c.WorkflowStartOptions, &c.PayloadInputOptions)
	if err != nil {
		return err
	}
	return cctx.Printer.Print(PrintOptions{}, struct {
		WorkflowId string `json:"workflowId"`
		RunId      string `json:"runId"`
		Type       string `json:"type"`
		Namespace  string `json:"namespace"`
		TaskQueue  string `json:"taskQueue"`
	}{
		WorkflowId: run.GetID(),
		RunId:      run.GetRunID(),
		Type:       c.WorkflowStartOptions.Type,
		Namespace:  c.Parent.Namespace,
		TaskQueue:  c.WorkflowStartOptions.TaskQueue,
	})
}

func (c *TemporalWorkflowCommand) startWorkflow(
	cctx *CommandContext,
	cl client.Client,
	workflowOpts *WorkflowStartOptions,
	inputOpts *PayloadInputOptions,
) (client.WorkflowRun, error) {
	startOpts, err := workflowOpts.buildStartOptions()
	if err != nil {
		return nil, err
	}
	input, err := inputOpts.buildRawInput()
	if err != nil {
		return nil, err
	}
	run, err := cl.ExecuteWorkflow(cctx, startOpts, workflowOpts.Type, input...)
	if err != nil {
		return nil, fmt.Errorf("failed starting workflow: %w", err)
	}
	cctx.Logger.Info("Started workflow", "workflow.id", run.GetID(), "workflow.runId", run.GetRunID())
	return run, nil
}

func (w *WorkflowStartOptions) buildStartOptions() (client.StartWorkflowOptions, error) {
	o := client.StartWorkflowOptions{
		ID:                       w.WorkflowId,
		TaskQueue:                w.TaskQueue,
		WorkflowRunTimeout:       w.RunTimeout,
		WorkflowExecutionTimeout: w.ExecutionTimeout,
		WorkflowTaskTimeout:      w.TaskTimeout,
		CronSchedule:             w.Cron,
	}
	if w.IdReusePolicy != "" {
		var err error
		o.WorkflowIDReusePolicy, err = stringToProtoEnum[enums.WorkflowIdReusePolicy](
			w.IdReusePolicy, enums.WorkflowIdReusePolicy_shorthandValue, enums.WorkflowIdReusePolicy_value)
		if err != nil {
			return o, fmt.Errorf("invalid workflow ID reuse policy: %w", err)
		}
	}
	if len(w.Memo) > 0 {
		var err error
		if o.Memo, err = stringKeysJSONValues(w.Memo); err != nil {
			return o, fmt.Errorf("invalid memo values: %w", err)
		}
	}
	if len(w.SearchAttribute) > 0 {
		var err error
		if o.SearchAttributes, err = stringKeysJSONValues(w.SearchAttribute); err != nil {
			return o, fmt.Errorf("invalid search attribute values: %w", err)
		}
	}
	return o, nil
}

func (p *PayloadInputOptions) buildRawInput() ([]any, error) {
	// Get input strings
	var inData [][]byte
	for _, in := range p.Input {
		if len(p.InputFile) > 0 {
			return nil, fmt.Errorf("cannot provide input and input file")
		}
		inData = append(inData, []byte(in))
	}
	for _, inFile := range p.InputFile {
		b, err := os.ReadFile(inFile)
		if err != nil {
			return nil, fmt.Errorf("failed reading input file %q: %w", inFile, err)
		}
		inData = append(inData, b)
	}

	// Build metadata
	metadata := map[string][]byte{"encoding": []byte("json/plain")}
	for _, meta := range p.InputMeta {
		metaPieces := strings.SplitN(meta, "=", 2)
		if len(metaPieces) != 2 {
			return nil, fmt.Errorf("metadata %v expected to have '='", meta)
		}
		metadata[metaPieces[0]] = []byte(metaPieces[1])
	}

	// Convert to raw values
	ret := make([]any, len(inData))
	for i, in := range inData {
		// First, if it's JSON, validate that it is accurate
		if strings.HasPrefix(string(metadata["encoding"]), "json/") && !json.Valid(in) {
			return nil, fmt.Errorf("input #%v is not valid JSON", i+1)
		}
		// Decode base64 if base64'd (std encoding only for now)
		if p.InputBase64 {
			var err error
			if in, err = base64.StdEncoding.DecodeString(string(in)); err != nil {
				return nil, fmt.Errorf("input #%v is not valid base64", i+1)
			}
		}
		ret[i] = rawValue{payload: &common.Payload{Data: in, Metadata: metadata}}
	}
	return ret, nil
}

// Rules:
//
//	Failed - red
//	Timeout - yellow
//	Canceled - magenta
//	Completed - green
//	Started - blue
//	Others - default (white/black)
func coloredEventType(cctx *CommandContext, e enums.EventType) string {
	// We don't color anything if JSON is the output
	if cctx.JSONOutput {
		return e.String()
	}

	fn := func(s string, ignore ...any) string { return s }
	switch e {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_ACTIVITY_TASK_FAILED,
		enums.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED:
		fn = color.RedString
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT,
		enums.EVENT_TYPE_WORKFLOW_TASK_TIMED_OUT,
		enums.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT,
		enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
		fn = color.YellowString
	case enums.EVENT_TYPE_TIMER_CANCELED,
		enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED,
		enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
		fn = color.MagentaString
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
		enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
		fn = color.GreenString
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED,
		enums.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
		fn = color.BlueString
	}
	return fn(e.String())
}
