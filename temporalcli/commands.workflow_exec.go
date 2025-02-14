package temporalcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/temporalio/cli/temporalcli/internal/printer"
	"go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/temporalproto"
	"go.temporal.io/sdk/client"
)

func (c *TemporalWorkflowStartCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	_, err = c.Parent.startWorkflow(cctx, cl, &c.SharedWorkflowStartOptions, &c.WorkflowStartOptions, &c.PayloadInputOptions, true)
	return err
}

func (c *TemporalWorkflowExecuteCommand) run(cctx *CommandContext, args []string) error {
	cl, err := c.Parent.ClientOptions.dialClient(cctx)
	if err != nil {
		return err
	}
	defer cl.Close()
	startTime := time.Now()
	run, err := c.Parent.startWorkflow(cctx, cl, &c.SharedWorkflowStartOptions, &c.WorkflowStartOptions, &c.PayloadInputOptions, false)
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
			includeDetails: c.Detailed,
			follow:         true,
		}
		if err := iter.print(cctx); err != nil && cctx.Err() == nil {
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
			cctx.Logger.Error("Workflow failed, and printing the output also failed", "error", err)
		}
		err = fmt.Errorf("workflow failed")
	}
	return err
}

type workflowJSONResult struct {
	WorkflowId     string          `json:"workflowId"`
	RunId          string          `json:"runId"`
	Type           string          `json:"type,omitempty"`
	Namespace      string          `json:"namespace,omitempty"`
	TaskQueue      string          `json:"taskQueue,omitempty"`
	DurationMillis int64           `json:"durationMillis,omitempty"`
	Status         string          `json:"status"`
	CloseEvent     json.RawMessage `json:"closeEvent"`
	Result         json.RawMessage `json:"result,omitempty"`
	History        json.RawMessage `json:"history,omitempty"`
}

func (r *workflowJSONResult) setCloseEvent(cctx *CommandContext, closeEvent *history.HistoryEvent) error {
	// Build status, result, and close event
	var err error
	switch closeEvent.EventType {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		r.Status = "COMPLETED"
		if r.Result, err = cctx.MarshalFriendlyJSONPayloads(
			closeEvent.GetWorkflowExecutionCompletedEventAttributes().GetResult()); err != nil {
			return fmt.Errorf("failed marshaling result: %w", err)
		}
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		r.Status = "FAILED"
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		r.Status = "TIMEOUT"
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		r.Status = "CANCELED"
	}
	if r.CloseEvent, err = cctx.MarshalProtoJSON(closeEvent); err != nil {
		return fmt.Errorf("failed marshaling status detail: %w", err)
	}
	return nil
}

func (c *TemporalWorkflowExecuteCommand) printJSONResult(
	cctx *CommandContext,
	client client.Client,
	run client.WorkflowRun,
	closeEvent *history.HistoryEvent,
	duration time.Duration,
) error {
	result := &workflowJSONResult{
		WorkflowId:     run.GetID(),
		RunId:          run.GetRunID(),
		Type:           c.SharedWorkflowStartOptions.Type,
		Namespace:      c.Parent.Namespace,
		TaskQueue:      c.SharedWorkflowStartOptions.TaskQueue,
		DurationMillis: int64(duration / time.Millisecond),
		Status:         "<unknown>",
		CloseEvent:     json.RawMessage("null"),
	}
	if err := result.setCloseEvent(cctx, closeEvent); err != nil {
		return err
	}

	// Build history if requested
	if c.Detailed {
		var histProto history.History
		iter := client.GetWorkflowHistory(cctx, run.GetID(), run.GetRunID(), false, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iter.HasNext() {
			event, err := iter.Next()
			if err != nil {
				return fmt.Errorf("failed reading history after completion: %w", err)
			}
			histProto.Events = append(histProto.Events, event)
		}
		// Do proto serialization here that would never do shorthand (i.e.
		// auto-lift JSON) payloads
		var err error
		if result.History, err = cctx.MarshalProtoJSONWithOptions(&histProto, false); err != nil {
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
	if closeEvent == nil {
		return nil
	}
	cctx.Printer.Println(color.MagentaString("Results:"))
	result := struct {
		RunTime        string `cli:",cardOmitEmpty"`
		Status         string
		Result         json.RawMessage `cli:",cardOmitEmpty"`
		ResultEncoding string          `cli:",cardOmitEmpty"`
		Failure        string          `cli:",cardOmitEmpty"`
	}{}
	if duration > 0 {
		result.RunTime = duration.Truncate(10 * time.Millisecond).String()
	}
	switch closeEvent.EventType {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		result.Status = color.GreenString("COMPLETED")
		var err error
		resultPayloads := closeEvent.GetWorkflowExecutionCompletedEventAttributes().GetResult()
		result.Result, err = cctx.MarshalFriendlyJSONPayloads(resultPayloads)
		if err != nil {
			return fmt.Errorf("failed marshaling result: %w", err)
		}
		if resultPayloads != nil && len(resultPayloads.Payloads) > 0 {
			metadata := resultPayloads.Payloads[0].GetMetadata()
			if metadata != nil {
				if enc, ok := metadata["encoding"]; ok {
					result.ResultEncoding = string(enc)
				}
			}
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
	sharedWorkflowOpts *SharedWorkflowStartOptions,
	workflowOpts *WorkflowStartOptions,
	inputOpts *PayloadInputOptions,
	printRunningExecutionEvenWithJSON bool,
) (client.WorkflowRun, error) {
	startOpts, err := buildStartOptions(sharedWorkflowOpts, workflowOpts)
	if err != nil {
		return nil, err
	}
	input, err := inputOpts.buildRawInput()
	if err != nil {
		return nil, err
	}
	run, err := cl.ExecuteWorkflow(cctx, startOpts, sharedWorkflowOpts.Type, input...)
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
			Type:       sharedWorkflowOpts.Type,
			Namespace:  c.Namespace,
			TaskQueue:  sharedWorkflowOpts.TaskQueue,
		}, printer.StructuredOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed printing: %w", err)
		}
	}
	return run, nil
}

func buildStartOptions(sw *SharedWorkflowStartOptions, w *WorkflowStartOptions) (client.StartWorkflowOptions, error) {
	o := client.StartWorkflowOptions{
		ID:                                       sw.WorkflowId,
		TaskQueue:                                sw.TaskQueue,
		WorkflowRunTimeout:                       sw.RunTimeout.Duration(),
		WorkflowExecutionTimeout:                 sw.ExecutionTimeout.Duration(),
		WorkflowTaskTimeout:                      sw.TaskTimeout.Duration(),
		CronSchedule:                             w.Cron,
		WorkflowExecutionErrorWhenAlreadyStarted: w.FailExisting,
		StartDelay:                               w.StartDelay.Duration(),
		StaticSummary:                            sw.StaticSummary,
		StaticDetails:                            sw.StaticDetails,
	}
	if w.IdReusePolicy.Value != "" {
		var err error
		o.WorkflowIDReusePolicy, err = stringToProtoEnum[enums.WorkflowIdReusePolicy](
			w.IdReusePolicy.Value, enums.WorkflowIdReusePolicy_shorthandValue, enums.WorkflowIdReusePolicy_value)
		if err != nil {
			return o, fmt.Errorf("invalid workflow ID reuse policy: %w", err)
		}
	}
	if w.IdConflictPolicy.Value != "" {
		var err error
		o.WorkflowIDConflictPolicy, err = stringToProtoEnum[enums.WorkflowIdConflictPolicy](
			w.IdConflictPolicy.Value, enums.WorkflowIdConflictPolicy_shorthandValue, enums.WorkflowIdConflictPolicy_value)
		if err != nil {
			return o, fmt.Errorf("invalid workflow ID conflict policy: %w", err)
		}
	}
	if len(sw.Memo) > 0 {
		var err error
		if o.Memo, err = stringKeysJSONValues(sw.Memo, false); err != nil {
			return o, fmt.Errorf("invalid memo values: %w", err)
		}
	}
	if len(sw.SearchAttribute) > 0 {
		var err error
		if o.SearchAttributes, err = stringKeysJSONValues(sw.SearchAttribute, false); err != nil {
			return o, fmt.Errorf("invalid search attribute values: %w", err)
		}
	}
	return o, nil
}

func (p *PayloadInputOptions) buildRawInput() ([]any, error) {
	payloads, err := p.buildRawInputPayloads()
	if err != nil {
		return nil, err
	}
	// Convert to raw values that our special data converter understands
	ret := make([]any, len(payloads.Payloads))
	for i, payload := range payloads.Payloads {
		ret[i] = RawValue{payload}
	}
	return ret, nil
}

func (p *PayloadInputOptions) buildRawInputPayloads() (*common.Payloads, error) {
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
	return CreatePayloads(inData, metadata, p.InputBase64)
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
	// If set true, long poll the history for updates
	follow bool
	// If and when the iterator encounters a workflow-terminating event, it will store it here
	wfResult *history.HistoryEvent

	// Internal
	iter client.HistoryEventIterator
}

func (s *structuredHistoryIter) print(cctx *CommandContext) error {
	// If we're not including details, just print the streaming table
	if !s.includeDetails {
		return cctx.Printer.PrintStructuredTableIter(
			structuredHistoryEventType,
			s,
			printer.StructuredOptions{Table: &printer.TableOptions{}},
		)
	}

	// Since details are wanted, we are going to do each event as a section
	first := true
	for {
		event, err := s.NextRawEvent()
		if event == nil || err != nil {
			return err
		}
		// Add blank line if not first
		if !first {
			cctx.Printer.Println()
		}
		first = false

		// Print section heading
		cctx.Printer.Printlnf("--------------- [%v] %v ---------------", event.EventId, event.EventType)
		// Convert the event to dot-delimited-field/value and print one per line
		fields, err := s.flattenFields(cctx, event)
		if err != nil {
			return fmt.Errorf("failed flattening event fields: %w", err)
		}
		for _, field := range fields {
			cctx.Printer.Printlnf("%v: %v", field.field, field.value)
		}
	}
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
	Type string `cli:",width=26"`
}

var structuredHistoryEventType = reflect.TypeOf(structuredHistoryEvent{})

func (s *structuredHistoryIter) Next() (any, error) {
	event, err := s.NextRawEvent()
	if err != nil {
		return nil, err
	}
	if event == nil {
		return nil, nil
	}
	// Build data
	data := structuredHistoryEvent{
		ID:   event.EventId,
		Time: event.EventTime.AsTime().Format(time.RFC3339),
		Type: coloredEventType(event.EventType),
	}

	// Follow continue as new
	if attr := event.GetWorkflowExecutionContinuedAsNewEventAttributes(); attr != nil {
		s.runID = attr.NewExecutionRunId
		s.iter = nil
	}
	return data, nil
}

func (s *structuredHistoryIter) NextRawEvent() (*history.HistoryEvent, error) {
	// Load iter
	if s.iter == nil {
		s.iter = s.client.GetWorkflowHistory(
			s.ctx, s.workflowID, s.runID, s.follow, enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
	}
	if !s.iter.HasNext() {
		return nil, nil
	}
	event, err := s.iter.Next()
	if err != nil {
		return nil, err
	}
	if isWorkflowTerminatingEvent(event.EventType) {
		s.wfResult = event
	}
	return event, nil
}

type eventFieldValue struct {
	field string
	value string
}

func (s *structuredHistoryIter) flattenFields(
	cctx *CommandContext,
	event *history.HistoryEvent,
) ([]eventFieldValue, error) {
	// We want all event fields and all attribute fields converted to the same
	// top-level JSON object. First do the proto conversion.
	opts := temporalproto.CustomJSONMarshalOptions{}
	if cctx.JSONShorthandPayloads {
		opts.Metadata = map[string]any{common.EnablePayloadShorthandMetadataKey: true}
	}
	protoJSON, err := opts.Marshal(event)
	if err != nil {
		return nil, fmt.Errorf("failed marshaling event: %w", err)
	}
	// Convert from string back to JSON
	dec := json.NewDecoder(bytes.NewReader(protoJSON))
	// We want json.Number
	dec.UseNumber()
	fieldsMap := map[string]any{}
	if err := dec.Decode(&fieldsMap); err != nil {
		return nil, fmt.Errorf("failed unmarshaling event proto: %w", err)
	}
	// Exclude eventId and eventType
	delete(fieldsMap, "eventId")
	delete(fieldsMap, "eventType")
	// Lift any "Attributes"-suffixed fields up to the top level
	for k, v := range fieldsMap {
		if strings.HasSuffix(k, "Attributes") {
			subMap, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("unexpectedly invalid attribute map")
			}
			for subK, subV := range subMap {
				fieldsMap[subK] = subV
			}
			delete(fieldsMap, k)
		}
	}
	// Flatten JSON map and sort
	fields, err := s.flattenJSONValue(nil, "", fieldsMap)
	if err != nil {
		return nil, err
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].field < fields[j].field })
	return fields, nil
}

func (s *structuredHistoryIter) flattenJSONValue(
	to []eventFieldValue,
	field string,
	value any,
) ([]eventFieldValue, error) {
	var err error
	switch value := value.(type) {
	case bool, string, json.Number, nil:
		// Note, empty values should not occur
		to = append(to, eventFieldValue{field, fmt.Sprintf("%v", value)})
	case []any:
		for i, subValue := range value {
			if to, err = s.flattenJSONValue(to, fmt.Sprintf("%v[%v]", field, i), subValue); err != nil {
				return nil, err
			}
		}
	case map[string]any:
		// Only add a dot if existing field not empty (i.e. not first)
		prefix := field
		if prefix != "" {
			prefix += "."
		}
		for subField, subValue := range value {
			if to, err = s.flattenJSONValue(to, prefix+subField, subValue); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("failed converting field %v, unknown type %T", field, value)
	}
	return to, nil
}

func isWorkflowTerminatingEvent(t enums.EventType) bool {
	switch t {
	case enums.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED,
		enums.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED,
		enums.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT,
		enums.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		return true
	}
	return false
}
