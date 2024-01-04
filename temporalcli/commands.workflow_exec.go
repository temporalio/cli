package temporalcli

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func (c *TemporalWorkflowStartCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = c.Parent.startWorkflow(cctx, cl, &c.WorkflowStartOptions, &c.PayloadInputOptions, true)
	return err
}

func (c *TemporalWorkflowExecuteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	startTime := time.Now()
	run, err := c.Parent.startWorkflow(cctx, cl, &c.WorkflowStartOptions, &c.PayloadInputOptions, false)
	if err != nil {
		return err
	}
	// Separate newline
	cctx.Printer.Println()

	// Print history only if not JSON
	if !cctx.JSONOutput {
		cctx.Printer.Println(color.MagentaString("Progress:"))
		iter := &structuredHistoryIter{
			ctx:            cctx,
			client:         cl,
			workflowID:     run.GetID(),
			runID:          run.GetRunID(),
			includeDetails: c.EventDetails,
		}
		if err := iter.print(cctx.Printer); err != nil && cctx.Err() == nil {
			return fmt.Errorf("displaying history failed: %w", err)
		}
		// Separate newline
		cctx.Printer.Println()
	}

	// Get the close event, following continue as new
	var closeEvent *history.HistoryEvent
	for runID := run.GetRunID(); closeEvent == nil; {
		iter := cl.GetWorkflowHistory(cctx, run.GetID(), runID, true, enums.HISTORY_EVENT_FILTER_TYPE_CLOSE_EVENT)
		if !iter.HasNext() {
			return fmt.Errorf("missing close event")
		} else if closeEvent, err = iter.Next(); err != nil {
			return fmt.Errorf("failed getting close event: %w", err)
		} else if canAttr := closeEvent.GetWorkflowExecutionContinuedAsNewEventAttributes(); canAttr != nil {
			closeEvent, runID = nil, canAttr.NewExecutionRunId
		}
	}
	duration := time.Since(startTime)

	// Print result
	if cctx.JSONOutput {
		err = c.printJSONResult(cctx, cl, run, closeEvent, duration)
	} else {
		err = printTextResult(cctx, closeEvent, duration)
	}
	// Log print failure and return workflow failure if workflow failed
	if closeEvent.EventType != enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED {
		if err != nil {
			cctx.Logger.Error("Workflow failed, but also failed printing output", "error", err)
		}
		err = fmt.Errorf("workflow failed")
	}
	return err
}

func (c *TemporalWorkflowExecuteCommand) printJSONResult(
	cctx *CommandContext,
	client client.Client,
	run client.WorkflowRun,
	closeEvent *history.HistoryEvent,
	duration time.Duration,
) error {
	result := struct {
		WorkflowId     string          `json:"workflowId"`
		RunId          string          `json:"runId"`
		Type           string          `json:"type"`
		Namespace      string          `json:"namespace"`
		TaskQueue      string          `json:"taskQueue"`
		DurationMillis int64           `json:"durationMillis"`
		Status         string          `json:"status"`
		CloseEvent     json.RawMessage `json:"closeEvent"`
		Result         json.RawMessage `json:"result,omitempty"`
		History        json.RawMessage `json:"history,omitempty"`
	}{
		WorkflowId:     run.GetID(),
		RunId:          run.GetRunID(),
		Type:           c.WorkflowStartOptions.Type,
		Namespace:      c.Parent.Namespace,
		TaskQueue:      c.WorkflowStartOptions.TaskQueue,
		DurationMillis: int64(duration / time.Millisecond),
		Status:         "<unknown>",
		CloseEvent:     json.RawMessage("null"),
	}
	// Build status, result, and close event
	var err error
	switch closeEvent.EventType {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		result.Status = "COMPLETED"
		if result.Result, err = cctx.MarshalFriendlyJSONPayloads(
			closeEvent.GetWorkflowExecutionCompletedEventAttributes().GetResult()); err != nil {
			return fmt.Errorf("failed marshaling result: %w", err)
		}
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		result.Status = "FAILED"
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		result.Status = "TIMEOUT"
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		result.Status = "CANCELED"
	}
	if result.CloseEvent, err = cctx.MarshalProtoJSON(closeEvent); err != nil {
		return fmt.Errorf("failed marshaling status detail: %w", err)
	}

	// Build history if requested
	if c.EventDetails {
		var histProto history.History
		iter := client.GetWorkflowHistory(cctx, run.GetID(), run.GetRunID(), false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				return fmt.Errorf("failed reading history after completion: %w", err)
			}
			histProto.Events = append(histProto.Events, event)
		}
		// Do proto serialization here that would never do shorthand payloads
		if result.History, err = MarshalProtoJSONWithOptions(&histProto, false); err != nil {
			return fmt.Errorf("failed marshaling history: %w", err)
		}
	}

	// Print
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func printTextResult(
	cctx *CommandContext,
	closeEvent *history.HistoryEvent,
	duration time.Duration,
) error {
	cctx.Printer.Println(color.MagentaString("Results:"))
	result := struct {
		RunTime string `cli:",cardOmitEmpty"`
		Status  string
		Result  json.RawMessage `cli:",cardOmitEmpty"`
		Failure string          `cli:",cardOmitEmpty"`
	}{}
	if duration > 0 {
		result.RunTime = duration.Truncate(10 * time.Millisecond).String()
	}
	switch closeEvent.EventType {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		result.Status = color.GreenString("COMPLETED")
		var err error
		result.Result, err = cctx.MarshalFriendlyJSONPayloads(
			closeEvent.GetWorkflowExecutionCompletedEventAttributes().GetResult())
		if err != nil {
			return fmt.Errorf("failed marshaling result: %w", err)
		}
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		result.Status = color.RedString("FAILED")
		result.Failure = cctx.MarshalFriendlyFailureBodyText(
			closeEvent.GetWorkflowExecutionFailedEventAttributes().Failure, "    ")
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		result.Status = color.RedString("TIMEOUT")
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		result.Status = color.RedString("CANCELED")
	}
	return cctx.Printer.PrintStructured(result, printer.StructuredOptions{})
}

func (c *TemporalWorkflowCommand) startWorkflow(
	cctx *CommandContext,
	cl client.Client,
	workflowOpts *WorkflowStartOptions,
	inputOpts *PayloadInputOptions,
	printRunningExecutionEvenWithJSON bool,
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

	// Print running execution
	if !cctx.JSONOutput || printRunningExecutionEvenWithJSON {
		cctx.Printer.Println(color.MagentaString("Running execution:"))
		err := cctx.Printer.PrintStructured(struct {
			WorkflowId string `json:"workflowId"`
			RunId      string `json:"runId"`
			Type       string `json:"type"`
			Namespace  string `json:"namespace"`
			TaskQueue  string `json:"taskQueue"`
		}{
			WorkflowId: run.GetID(),
			RunId:      run.GetRunID(),
			Type:       workflowOpts.Type,
			Namespace:  c.Namespace,
			TaskQueue:  workflowOpts.TaskQueue,
		}, printer.StructuredOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed printing: %w", err)
		}
	}
	return run, nil
}

func (w *WorkflowStartOptions) buildStartOptions() (client.StartWorkflowOptions, error) {
	o := client.StartWorkflowOptions{
		ID:                                       w.WorkflowId,
		TaskQueue:                                w.TaskQueue,
		WorkflowRunTimeout:                       w.RunTimeout,
		WorkflowExecutionTimeout:                 w.ExecutionTimeout,
		WorkflowTaskTimeout:                      w.TaskTimeout,
		CronSchedule:                             w.Cron,
		WorkflowExecutionErrorWhenAlreadyStarted: !w.AllowExisting,
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
func coloredEventType(e enums.EventType) string {
	fn := func(s string, ignore ...any) string { return s }
	switch e {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_WORKFLOW_TASK_FAILED,
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

type structuredHistoryIter struct {
	ctx            context.Context
	client         client.Client
	workflowID     string
	runID          string
	includeDetails bool

	// Internal
	iter client.HistoryEventIterator
}

func (s *structuredHistoryIter) print(p *printer.Printer) error {
	options := printer.StructuredOptions{Table: &printer.TableOptions{}}
	if !s.includeDetails {
		options.ExcludeFields = []string{"Details"}
	}
	return p.PrintStructuredIter(structuredHistoryEventType, s, options)
}

type structuredHistoryEvent struct {
	// Once it gets into the thousands width will bleed over, that's fine
	ID int64 `cli:",width=3"`
	// We pre-format time here because the default --time-format makes no
	// sense if it's "relative"
	// TODO(cretz): Allow user to override?
	Time string `cli:",width=20"`
	// We're going to set width to a semi-reasonable number for good header
	// placement, but we expect it to extend past for larger
	Type    string `cli:",width=26"`
	Details string `cli:",width=20"`
}

var structuredHistoryEventType = reflect.TypeOf(structuredHistoryEvent{})

func (s *structuredHistoryIter) Next() (any, error) {
	// Load iter
	if s.iter == nil {
		s.iter = s.client.GetWorkflowHistory(s.ctx, s.workflowID, s.runID, true, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	}
	if !s.iter.HasNext() {
		return nil, nil
	}
	event, err := s.iter.Next()
	if err != nil {
		return nil, err
	}

	// Build data
	data := structuredHistoryEvent{
		ID:   event.EventId,
		Time: event.EventTime.AsTime().Format(time.RFC3339),
		Type: coloredEventType(event.EventType),
	}
	if s.includeDetails {
		// First field in the attributes
		attrs := reflect.ValueOf(event.Attributes).Elem().Field(0).Interface().(proto.Message)
		if b, err := protojson.Marshal(attrs); err != nil {
			data.Details = "<failed serializing details>"
		} else {
			data.Details = string(b)
		}
	}

	// Follow continue as new
	if attr := event.GetWorkflowExecutionContinuedAsNewEventAttributes(); attr != nil {
		s.runID = attr.NewExecutionRunId
		s.iter = nil
	}
	return data, nil
}
