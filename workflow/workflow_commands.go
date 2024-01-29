package workflow

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pborman/uuid"
	"github.com/temporalio/cli/batch"
	"github.com/temporalio/cli/client"
	"github.com/temporalio/cli/common"
	"github.com/temporalio/cli/common/stringify"
	"github.com/temporalio/cli/trace"
	"github.com/temporalio/tctl-kit/pkg/color"
	"github.com/temporalio/tctl-kit/pkg/iterator"
	"github.com/temporalio/tctl-kit/pkg/output"
	"github.com/temporalio/tctl-kit/pkg/pager"
	"github.com/urfave/cli/v2"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	failurepb "go.temporal.io/api/failure/v1"
	historypb "go.temporal.io/api/history/v1"
	querypb "go.temporal.io/api/query/v1"
	"go.temporal.io/api/serviceerror"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"
	clispb "go.temporal.io/server/api/cli/v1"
	scommon "go.temporal.io/server/common"
	"go.temporal.io/server/common/backoff"
	"go.temporal.io/server/common/collection"
	"go.temporal.io/server/common/convert"
	"go.temporal.io/server/common/payload"
	"go.temporal.io/server/common/primitives/timestamp"
	"go.temporal.io/server/common/searchattribute"
)

var (
	tableHeaderBlue = tablewriter.Colors{tablewriter.FgHiBlueColor}
	resetTypesMap   = map[string]interface{}{
		"FirstWorkflowTask":  "",
		"LastWorkflowTask":   "",
		"LastContinuedAsNew": "",
	}
	resetReapplyTypesMap = map[string]interface{}{
		"":       enumspb.RESET_REAPPLY_TYPE_SIGNAL, // default value
		"Signal": enumspb.RESET_REAPPLY_TYPE_SIGNAL,
		"None":   enumspb.RESET_REAPPLY_TYPE_NONE,
	}
)

func StartWorkflowBaseArgs(c *cli.Context) (
	taskQueue string,
	workflowType string,
	et, rt, dt int,
	wid string,
) {
	taskQueue = c.String(common.FlagTaskQueue)
	workflowType = c.String(common.FlagWorkflowType)
	if workflowType == "" {
		// "workflow start" expects common.FlagType rather than full common.FlagWorkflowType
		workflowType = c.String(common.FlagType)
	}
	et = c.Int(common.FlagWorkflowExecutionTimeout)
	rt = c.Int(common.FlagWorkflowRunTimeout)
	dt = c.Int(common.FlagWorkflowTaskTimeout)
	wid = c.String(common.FlagWorkflowID)
	if len(wid) == 0 {
		wid = uuid.New()
	}
	return
}

// StartWorkflow starts a new workflow execution and optionally prints progress
func StartWorkflow(c *cli.Context, printProgress bool) error {
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	taskQueue, workflowType, et, rt, dt, wid := StartWorkflowBaseArgs(c)

	reusePolicy := common.DefaultWorkflowIDReusePolicy
	if c.IsSet(common.FlagWorkflowIDReusePolicy) {
		reusePolicyInt, err := common.StringToEnum(c.String(common.FlagWorkflowIDReusePolicy), enumspb.WorkflowIdReusePolicy_value)
		if err != nil {
			return fmt.Errorf("unable to parse workflow ID reuse policy: %w", err)
		}
		reusePolicy = enumspb.WorkflowIdReusePolicy(reusePolicyInt)
	}

	inputs, err := common.UnmarshalInputsFromCLI(c)
	if err != nil {
		return err
	}

	wo := sdkclient.StartWorkflowOptions{
		ID:                       wid,
		TaskQueue:                taskQueue,
		WorkflowExecutionTimeout: time.Duration(et) * time.Second,
		WorkflowTaskTimeout:      time.Duration(dt) * time.Second,
		WorkflowRunTimeout:       time.Duration(rt) * time.Second,
		WorkflowIDReusePolicy:    reusePolicy,
	}
	if c.IsSet(common.FlagCronSchedule) {
		wo.CronSchedule = c.String(common.FlagCronSchedule)
	}

	wo.Memo, err = UnmarshalMemoFromCLI(c)
	if err != nil {
		return err
	}
	wo.SearchAttributes, err = UnmarshalSearchAttrFromCLI(c)
	if err != nil {
		return err
	}

	tcCtx, cancel := common.NewContextForLongPoll(c)
	defer cancel()
	resp, err := sdkClient.ExecuteWorkflow(tcCtx, wo, workflowType, inputs...)

	if err != nil {
		return fmt.Errorf("unable to run workflow: %w", err)
	}

	executionDetails := struct {
		WorkflowId string
		RunId      string
		Type       string
		Namespace  string
		TaskQueue  string
		Args       string
	}{

		WorkflowId: wid,
		RunId:      resp.GetRunID(),
		Type:       workflowType,
		Namespace:  namespace,
		TaskQueue:  taskQueue,
		Args:       common.Truncate(formatInputsForDisplay(inputs)),
	}
	data := []interface{}{
		executionDetails,
	}
	fmt.Println(color.Magenta(c, "Running execution:"))
	opts := &output.PrintOptions{
		Fields:       []string{"WorkflowId", "RunId", "Type", "Namespace", "TaskQueue", "Args"},
		ForceFields:  true,
		OutputFormat: output.Card,
		Separator:    "",
	}
	err = output.PrintItems(c, data, opts)
	if err != nil {
		return err
	}

	if printProgress {
		return printWorkflowProgress(c, wid, resp.GetRunID(), true)
	}

	return nil
}

func formatInputsForDisplay(inputs []interface{}) string {
	var result []string
	for _, input := range inputs {
		s, _ := json.Marshal(input)
		result = append(result, string(s))
	}
	return fmt.Sprintf("[%s]", strings.Join(result, ","))
}

func UnmarshalSearchAttrFromCLI(c *cli.Context) (map[string]interface{}, error) {
	raw := c.StringSlice(common.FlagSearchAttribute)
	parsed, err := common.SplitKeyValuePairs(raw)
	if err != nil {
		return nil, err
	}

	attributes := make(map[string]interface{}, len(parsed))
	for k, v := range parsed {
		var j interface{}
		if err := json.Unmarshal([]byte(v), &j); err != nil {
			return nil, fmt.Errorf("unable to parse Search Attribute JSON: %w", err)
		}
		attributes[k] = j
	}

	return attributes, nil
}

func UnmarshalMemoFromCLI(c *cli.Context) (map[string]interface{}, error) {
	if !c.IsSet(common.FlagMemo) && !c.IsSet(common.FlagMemoFile) {
		return nil, nil
	}

	raw := c.StringSlice(common.FlagMemo)

	var rawFromFile []string
	if c.IsSet(common.FlagMemoFile) {
		inputFile := c.String(common.FlagMemoFile)
		// The input comes from a trusted user
		// #nosec
		data, err := os.ReadFile(inputFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read memo file %s", inputFile)
		}
		rawFromFile = strings.Split(string(data), "\n")
	}

	raw = append(raw, rawFromFile...)

	parsed, err := common.SplitKeyValuePairs(raw)
	if err != nil {
		return nil, err
	}

	memo := make(map[string]interface{}, len(parsed))
	for k, v := range parsed {
		var j interface{}
		if err := json.Unmarshal([]byte(v), &j); err != nil {
			return nil, fmt.Errorf("unable to parse Search Attribute JSON: %w", err)
		}
		memo[k] = j
	}

	return memo, nil
}

// historyTableIter adapts history iterator for Table output view
type historyTableIter struct {
	iter           iterator.Iterator[*historypb.HistoryEvent]
	maxFieldLength int
	wfResult       *historypb.HistoryEvent
}

func (h *historyTableIter) HasNext() bool {
	return h.iter.HasNext()
}

func (h *historyTableIter) Next() (interface{}, error) {
	event, err := h.iter.Next()
	if err != nil {
		return nil, err
	}

	reflect.ValueOf(h.wfResult).Elem().Set(reflect.ValueOf(event).Elem())

	// adapted structure for Table output view
	return struct {
		ID      string
		Time    string
		Type    string
		Details string
	}{
		ID:      convert.Int64ToString(event.GetEventId()),
		Time:    common.FormatTime(timestamp.TimeValue(event.GetEventTime()), false),
		Type:    common.ColorEvent(event),
		Details: historyEventToString(event, false, h.maxFieldLength),
	}, nil
}

// helper function to print workflow progress with time refresh every second
func printWorkflowProgress(c *cli.Context, wid, rid string, watch bool) error {
	isJSON := false
	if c.IsSet(output.FlagOutput) {
		outputFlag := c.String(output.FlagOutput)
		isJSON = outputFlag == string(output.JSON)
	}

	var maxFieldLength = c.Int(common.FlagMaxFieldLength)
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	tcCtx, cancel := common.NewIndefiniteContext(c)
	defer cancel()

	doneChan := make(chan bool)
	timeElapsed := 1
	timeElapsedExists := false
	ticker := time.NewTicker(time.Second).C
	if !isJSON {
		fmt.Println(color.Magenta(c, "Progress:"))
	}

	var wfResult historypb.HistoryEvent

	errChan := make(chan error)
	go func() {
		iter := sdkClient.GetWorkflowHistory(tcCtx, wid, rid, watch, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		if isJSON {
			printReplayableHistory(c, iter)
		} else {
			hIter := &historyTableIter{iter: iter, maxFieldLength: maxFieldLength, wfResult: &wfResult}
			po := &output.PrintOptions{
				Fields:     []string{"ID", "Time", "Type"},
				FieldsLong: []string{"Details"},
				Pager:      pager.Less,
			}
			err = output.PrintIterator(c, hIter, po)
		}

		if err != nil {
			errChan <- err
			return
		}

		doneChan <- true
	}()

	for {
		select {
		case <-ticker:
			if !watch {
				continue
			}

			if !isJSON {
				if timeElapsedExists {
					removePrevious2LinesFromTerminal()
				}
				fmt.Printf("\nTime elapsed: %ds\n", timeElapsed)
			}

			timeElapsedExists = true
			timeElapsed++
		case <-doneChan: // print result of this run
			if !isJSON {
				fmt.Println(color.Magenta(c, "\nResult:"))
				if watch {
					fmt.Printf("  Run Time: %d seconds\n", timeElapsed)
				}
				printRunStatus(c, &wfResult)
			}
			return nil
		case err = <-errChan:
			return err
		}
	}
}

func printReplayableHistory(c *cli.Context, iter iterator.Iterator[*historypb.HistoryEvent]) error {
	var events []*historypb.HistoryEvent
	for iter.HasNext() {
		event, err := iter.Next()
		if err != nil {
			return err

		}
		events = append(events, event)
	}

	history := &historypb.History{Events: events}

	common.PrettyPrintJSONObject(c, history)

	return nil
}

// TerminateWorkflow terminates workflow executions based on filter parameters
func TerminateWorkflow(c *cli.Context) error {
	if c.String(common.FlagQuery) != "" {
		return batch.BatchTerminate(c)
	}

	return terminateWorkflowByID(c)
}

// terminateWorkflowByID terminates a single workflow execution
func terminateWorkflowByID(c *cli.Context) error {
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	wid, err := common.RequiredFlag(c, common.FlagWorkflowID)
	if err != nil {
		return err
	}
	rid := c.String(common.FlagRunID)
	reason := c.String(common.FlagReason)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	err = sdkClient.TerminateWorkflow(ctx, wid, rid, reason, nil)
	if err != nil {
		return fmt.Errorf("unable to terminate workflow: %w", err)
	}

	fmt.Println("Terminate workflow succeeded")

	return nil
}

// DeleteWorkflow deletes workflow executions based on filter parameters
func DeleteWorkflow(c *cli.Context) error {
	if c.String(common.FlagQuery) != "" {
		return batch.BatchDelete(c)
	}

	return deleteWorkflowByID(c)
}

// deleteWorkflowByID deletes a single workflow execution
func deleteWorkflowByID(c *cli.Context) error {
	nsName, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)

	fclient := client.Factory(c.App).FrontendClient(c)
	ctx, cancel := common.NewContext(c)
	defer cancel()
	_, err = fclient.DeleteWorkflowExecution(ctx, &workflowservice.DeleteWorkflowExecutionRequest{
		Namespace: nsName,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
	})

	if err != nil {
		return fmt.Errorf("unable to delete workflow: %w", err)
	}

	fmt.Println(color.Green(c, "Delete workflow succeeded"))

	return nil
}

// CancelWorkflow cancels workflow executions based on filter parameters
func CancelWorkflow(c *cli.Context) error {
	if c.String(common.FlagQuery) != "" {
		return batch.BatchCancel(c)
	}

	return cancelWorkflowByID(c)
}

// cancelWorkflowByID cancels a single workflow execution
func cancelWorkflowByID(c *cli.Context) error {
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	wid, err := common.RequiredFlag(c, common.FlagWorkflowID)
	if err != nil {
		return err
	}
	rid := c.String(common.FlagRunID)

	ctx, cancel := common.NewContext(c)
	defer cancel()
	err = sdkClient.CancelWorkflow(ctx, wid, rid)
	if err != nil {
		return fmt.Errorf("unable to cancel workflow: %w", err)
	}
	fmt.Println(color.Green(c, "canceled workflow, workflow id: %s, run id: %s", wid, rid))

	return nil
}

// SignalWorkflow signals workflow executions based on filter parameters
func SignalWorkflow(c *cli.Context) error {
	if c.String(common.FlagQuery) != "" {
		return batch.BatchSignal(c)
	}

	return signalWorkflowByID(c)
}

// signalWorkflowByID signals a single workflow execution
func signalWorkflowByID(c *cli.Context) error {
	serviceClient := client.Factory(c.App).FrontendClient(c)

	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}

	wid, err := common.RequiredFlag(c, common.FlagWorkflowID)
	if err != nil {
		return err
	}
	rid := c.String(common.FlagRunID)
	name := c.String(common.FlagName)
	input, err := common.ProcessJSONInput(c)
	if err != nil {
		return err
	}

	tcCtx, cancel := common.NewContext(c)
	defer cancel()
	_, err = serviceClient.SignalWorkflowExecution(tcCtx, &workflowservice.SignalWorkflowExecutionRequest{
		Namespace: namespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		SignalName: name,
		Input:      input,
		Identity:   common.GetCliIdentity(),
	})

	if err != nil {
		return fmt.Errorf("signal workflow failed: %w", err)
	}

	fmt.Println("Signal workflow succeeded")

	return nil
}

// QueryWorkflow query workflow execution
func QueryWorkflow(c *cli.Context) error {
	queryType := c.String(common.FlagType)

	if err := queryWorkflowHelper(c, queryType); err != nil {
		return err
	}

	return nil
}

// QueryWorkflowUsingStackTrace query workflow execution using __stack_trace as query type
func QueryWorkflowUsingStackTrace(c *cli.Context) error {
	return queryWorkflowHelper(c, "__stack_trace")
}

func queryWorkflowHelper(c *cli.Context, queryType string) error {
	fclient := client.Factory(c.App).FrontendClient(c)

	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)
	input, err := common.ProcessJSONInput(c)
	if err != nil {
		return err
	}

	tcCtx, cancel := common.NewContext(c)
	defer cancel()
	queryRequest := &workflowservice.QueryWorkflowRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		Query: &querypb.WorkflowQuery{
			QueryType: queryType,
		},
	}
	if input != nil {
		queryRequest.Query.QueryArgs = input
	}
	if c.IsSet(common.FlagQueryRejectCondition) {
		var rejectCondition enumspb.QueryRejectCondition
		switch c.String(common.FlagQueryRejectCondition) {
		case "not_open":
			rejectCondition = enumspb.QUERY_REJECT_CONDITION_NOT_OPEN
		case "not_completed_cleanly":
			rejectCondition = enumspb.QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
		default:
			return fmt.Errorf("invalid reject condition %v, valid values are \"not_open\" and \"not_completed_cleanly\"", c.String(common.FlagQueryRejectCondition))
		}
		queryRequest.QueryRejectCondition = rejectCondition
	}
	queryResponse, err := fclient.QueryWorkflow(tcCtx, queryRequest)
	if err != nil {
		return fmt.Errorf("query workflow failed: %w", err)
	}

	if queryResponse.QueryRejected != nil {
		fmt.Printf("Query was rejected, workflow has status: %v\n", queryResponse.QueryRejected.GetStatus())
	} else {
		queryResult := stringify.AnyToString(queryResponse.QueryResult, true, 0)
		fmt.Printf("Query result:\n%v\n", queryResult)
	}

	return nil
}

// ListWorkflow list workflow executions based on filters
func ListWorkflow(c *cli.Context) error {
	archived := c.Bool(common.FlagArchive)

	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		var items []interface{}
		var err error
		query := c.String(common.FlagQuery)

		if archived {
			items, npt, err = listArchivedWorkflows(c, sdkClient, npt, query)
		} else {
			items, npt, err = listWorkflows(c, sdkClient, npt, query)
		}

		if err != nil {
			return nil, nil, err
		}

		return items, npt, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &output.PrintOptions{
		Fields:     []string{"Status", "Execution.WorkflowId", "Type.Name", "StartTime"},
		FieldsLong: []string{"CloseTime", "Execution.RunId", "TaskQueue"},
		Pager:      pager.Less,
	}
	return output.PrintIterator(c, iter, opts)
}

// CountWorkflow count number of workflows
func CountWorkflow(c *cli.Context) error {
	sdkClient, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	query := c.String(common.FlagQuery)
	request := &workflowservice.CountWorkflowExecutionsRequest{
		Query: query,
	}

	var response *workflowservice.CountWorkflowExecutionsResponse
	op := func() error {
		ctx, cancel := common.NewContext(c)
		defer cancel()
		var err error
		response, err = sdkClient.CountWorkflow(ctx, request)
		if err != nil {
			return err
		}
		return nil
	}
	err = backoff.ThrottleRetry(op, scommon.CreateFrontendClientRetryPolicy(), scommon.IsContextDeadlineExceededErr)
	if err != nil {
		return fmt.Errorf("unable to count workflows: %w", err)
	}
	fmt.Printf("Total: %d\n", response.GetCount())
	groups := response.GetGroups()
	for _, g := range groups {
		values := make([]any, len(g.GroupValues))
		for i, v := range g.GroupValues {
			if err := payload.Decode(v, &values[i]); err != nil {
				return err
			}
		}
		fmt.Printf("Group: %v,  Count: %d\n", values, g.Count)
	}
	return nil
}

// DescribeWorkflow show information about the specified workflow execution
func DescribeWorkflow(c *cli.Context) error {
	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)

	frontendClient := client.Factory(c.App).FrontendClient(c)
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	printRaw := c.Bool(common.FlagPrintRaw) // printRaw is false by default,
	// and will show datetime and decoded search attributes instead of raw timestamp and byte arrays
	printResetPointsOnly := c.Bool(common.FlagResetPointsOnly)

	ctx, cancel := common.NewContext(c)
	defer cancel()

	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &workflowservice.DescribeWorkflowExecutionRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
	})
	if err != nil {
		return fmt.Errorf("workflow describe failed: %w", err)
	}

	if printResetPointsOnly {
		printAutoResetPoints(resp)
		return nil
	}

	if printRaw {
		common.PrettyPrintJSONObject(c, resp)
	} else {
		common.PrettyPrintJSONObject(c, convertDescribeWorkflowExecutionResponse(c, resp))
	}

	return nil
}

func printAutoResetPoints(resp *workflowservice.DescribeWorkflowExecutionResponse) {
	fmt.Println("Auto Reset Points:")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetBorder(true)
	table.SetColumnSeparator("|")
	header := []string{"Binary Checksum", "Create Time", "RunId", "EventId"}
	headerColor := []tablewriter.Colors{tableHeaderBlue, tableHeaderBlue, tableHeaderBlue, tableHeaderBlue}
	table.SetHeader(header)
	table.SetHeaderColor(headerColor...)
	if resp.WorkflowExecutionInfo.AutoResetPoints != nil && len(resp.WorkflowExecutionInfo.AutoResetPoints.Points) > 0 {
		for _, pt := range resp.WorkflowExecutionInfo.AutoResetPoints.Points {
			var row []string
			row = append(row, pt.GetBinaryChecksum())
			row = append(row, timestamp.TimeValue(pt.GetCreateTime()).String())
			row = append(row, pt.GetRunId())
			row = append(row, convert.Int64ToString(pt.GetFirstWorkflowTaskCompletedId()))
			table.Append(row)
		}
	}
	table.Render()
}

func convertDescribeWorkflowExecutionResponse(c *cli.Context, resp *workflowservice.DescribeWorkflowExecutionResponse) *clispb.DescribeWorkflowExecutionResponse {

	info := resp.GetWorkflowExecutionInfo()
	executionInfo := &clispb.WorkflowExecutionInfo{
		Execution:                    info.GetExecution(),
		Type:                         info.GetType(),
		CloseTime:                    info.GetCloseTime(),
		StartTime:                    info.GetStartTime(),
		Status:                       info.GetStatus(),
		HistoryLength:                info.GetHistoryLength(),
		ParentNamespaceId:            info.GetParentNamespaceId(),
		ParentExecution:              info.GetParentExecution(),
		Memo:                         info.GetMemo(),
		SearchAttributes:             convertSearchAttributes(c, info.GetSearchAttributes()),
		AutoResetPoints:              info.GetAutoResetPoints(),
		StateTransitionCount:         info.GetStateTransitionCount(),
		ExecutionTime:                info.GetExecutionTime(),
		HistorySizeBytes:             info.GetHistorySizeBytes(),
		MostRecentWorkerVersionStamp: info.GetMostRecentWorkerVersionStamp(),
	}

	var pendingActivitiesStr []*clispb.PendingActivityInfo
	for _, pendingActivity := range resp.GetPendingActivities() {
		pendingActivityStr := &clispb.PendingActivityInfo{
			ActivityId:         pendingActivity.GetActivityId(),
			ActivityType:       pendingActivity.GetActivityType(),
			State:              pendingActivity.GetState(),
			ScheduledTime:      pendingActivity.GetScheduledTime(),
			LastStartedTime:    pendingActivity.GetLastStartedTime(),
			LastHeartbeatTime:  pendingActivity.GetLastHeartbeatTime(),
			Attempt:            pendingActivity.GetAttempt(),
			MaximumAttempts:    pendingActivity.GetMaximumAttempts(),
			ExpirationTime:     pendingActivity.GetExpirationTime(),
			LastFailure:        convertFailure(pendingActivity.GetLastFailure()),
			LastWorkerIdentity: pendingActivity.GetLastWorkerIdentity(),
		}

		if pendingActivity.GetHeartbeatDetails() != nil {
			pendingActivityStr.HeartbeatDetails = stringify.AnyToString(pendingActivity.GetHeartbeatDetails(), true, 0)
		}
		pendingActivitiesStr = append(pendingActivitiesStr, pendingActivityStr)
	}

	return &clispb.DescribeWorkflowExecutionResponse{
		ExecutionConfig:       resp.ExecutionConfig,
		WorkflowExecutionInfo: executionInfo,
		PendingActivities:     pendingActivitiesStr,
		PendingChildren:       resp.PendingChildren,
		PendingWorkflowTask:   resp.PendingWorkflowTask,
	}
}

func convertSearchAttributes(c *cli.Context, searchAttributes *commonpb.SearchAttributes) *clispb.SearchAttributes {
	if len(searchAttributes.GetIndexedFields()) == 0 {
		return nil
	}

	fields, err := searchattribute.Stringify(searchAttributes, nil)
	if err != nil {
		fmt.Printf("%s: unable to stringify search attribute: %v\n",
			color.Magenta(c, "Warning"),
			err)
	}

	return &clispb.SearchAttributes{IndexedFields: fields}
}

func convertFailure(failure *failurepb.Failure) *clispb.Failure {
	if failure == nil {
		return nil
	}

	fType := reflect.TypeOf(failure.GetFailureInfo()).Elem().Name()
	if failure.GetTimeoutFailureInfo() != nil {
		fType = fmt.Sprintf("%s: %s", fType, failure.GetTimeoutFailureInfo().GetTimeoutType().String())
	}

	f := &clispb.Failure{
		Message:     failure.GetMessage(),
		Source:      failure.GetSource(),
		StackTrace:  failure.GetStackTrace(),
		Cause:       convertFailure(failure.GetCause()),
		FailureType: fType,
	}

	return f
}

func printRunStatus(c *cli.Context, event *historypb.HistoryEvent) {
	switch event.GetEventType() {
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		fmt.Printf("  Status: %s\n", color.Green(c, "COMPLETED"))
		result := stringify.AnyToString(event.GetWorkflowExecutionCompletedEventAttributes().GetResult(), true, 0)
		fmt.Printf("  Output: %s\n", result)
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		fmt.Printf("  Status: %s\n", color.Red(c, "FAILED"))
		fmt.Printf("  Failure: %s\n", convertFailure(event.GetWorkflowExecutionFailedEventAttributes().GetFailure()).String())
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		fmt.Printf("  Status: %s\n", color.Red(c, "TIMEOUT"))
		fmt.Printf("  Retry status: %s\n", event.GetWorkflowExecutionTimedOutEventAttributes().GetRetryState())
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		fmt.Printf("  Status: %s\n", color.Red(c, "CANCELED"))
		details := stringify.AnyToString(event.GetWorkflowExecutionCanceledEventAttributes().GetDetails(), true, 0)
		fmt.Printf("  Detail: %s\n", details)
	}
}

// ShowHistory shows the history of given workflow execution based on workflowID and runID.
func ShowHistory(c *cli.Context) error {
	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)
	follow := c.Bool(output.FlagFollow)

	return printWorkflowProgress(c, wid, rid, follow)
}

// ResetWorkflow reset workflow
func ResetWorkflow(c *cli.Context) error {
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	wid := c.String(common.FlagWorkflowID)
	reason := c.String(common.FlagReason)
	if len(reason) == 0 {
		return fmt.Errorf("reason flag cannot be empty")
	}
	rid := c.String(common.FlagRunID)
	eventID := c.Int64(common.FlagEventID)
	resetType := c.String(common.FlagType)
	extraForResetType, ok := resetTypesMap[resetType]
	if !ok && eventID <= 0 {
		return fmt.Errorf("specify either valid event id or reset type (one of %s)", strings.Join(mapKeysToArray(resetTypesMap), ", "))
	}
	if ok && len(extraForResetType.(string)) > 0 {
		value := c.String(extraForResetType.(string))
		if len(value) == 0 {
			return fmt.Errorf("option %s is required", extraForResetType.(string))
		}
	}
	resetReapplyType := c.String(common.FlagResetReapplyType)
	if _, ok := resetReapplyTypesMap[resetReapplyType]; !ok {
		return fmt.Errorf("must specify valid reset reapply type: %v", strings.Join(mapKeysToArray(resetReapplyTypesMap), ", "))
	}

	ctx, cancel := common.NewContext(c)
	defer cancel()

	frontendClient := client.Factory(c.App).FrontendClient(c)

	resetBaseRunID := rid
	workflowTaskFinishID := eventID
	if resetType != "" {
		resetBaseRunID, workflowTaskFinishID, err = getResetEventIDByType(ctx, c, resetType, namespace, wid, rid, frontendClient)
		if err != nil {
			return fmt.Errorf("getting reset event ID by type failed: %w", err)
		}
	}
	resp, err := frontendClient.ResetWorkflowExecution(ctx, &workflowservice.ResetWorkflowExecutionRequest{
		Namespace: namespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      resetBaseRunID,
		},
		Reason:                    fmt.Sprintf("%v:%v", common.GetCurrentUserFromEnv(), reason),
		WorkflowTaskFinishEventId: workflowTaskFinishID,
		RequestId:                 uuid.New(),
		ResetReapplyType:          resetReapplyTypesMap[resetReapplyType].(enumspb.ResetReapplyType),
	})
	if err != nil {
		return fmt.Errorf("reset failed: %w", err)
	}
	common.PrettyPrintJSONObject(c, resp)
	return nil
}

func processResets(c *cli.Context, namespace string, wes chan commonpb.WorkflowExecution, done chan bool, wg *sync.WaitGroup, params batchResetParamsType) {
	for {
		select {
		case we := <-wes:
			fmt.Println("received: ", we.GetWorkflowId(), we.GetRunId())
			wid := we.GetWorkflowId()
			rid := we.GetRunId()
			var err error
			for i := 0; i < 3; i++ {
				err = doReset(c, namespace, wid, rid, params)
				if err == nil {
					break
				}
				if _, ok := err.(*serviceerror.InvalidArgument); ok {
					break
				}
				fmt.Println("failed and retry...: ", wid, rid, err)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(2000)))
			}
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			if err != nil {
				fmt.Println("[ERROR] failed processing: ", wid, rid, err.Error())
			}
		case <-done:
			wg.Done()
			return
		}
	}
}

type batchResetParamsType struct {
	reason               string
	skipOpen             bool
	nonDeterministicOnly bool
	skipBaseNotCurrent   bool
	dryRun               bool
	resetType            string
}

// ResetInBatch resets workflow in batch
func ResetInBatch(c *cli.Context) error {
	namespace, err := common.RequiredFlag(c, common.FlagNamespace)
	if err != nil {
		return err
	}
	resetType := c.String(common.FlagType)

	inFileName := c.String(common.FlagInputFile)
	query := c.String(common.FlagQuery)
	excFileName := c.String(common.FlagExcludeFile)
	separator := c.String(common.FlagInputSeparator)
	parallel := c.Int(common.FlagParallelism)

	extraForResetType, ok := resetTypesMap[resetType]
	if !ok {
		return fmt.Errorf("reset type is not supported: %v", extraForResetType)
	} else if len(extraForResetType.(string)) > 0 {
		value := c.String(extraForResetType.(string))
		if len(value) == 0 {
			return fmt.Errorf("option %s is required", extraForResetType.(string))
		}
	}

	batchResetParams := batchResetParamsType{
		reason:               c.String(common.FlagReason),
		skipOpen:             c.Bool(common.FlagSkipCurrentOpen),
		nonDeterministicOnly: c.Bool(common.FlagNonDeterministic),
		skipBaseNotCurrent:   c.Bool(common.FlagSkipBaseIsNotCurrent),
		dryRun:               c.Bool(common.FlagDryRun),
		resetType:            resetType,
	}

	if inFileName == "" && query == "" {
		return fmt.Errorf("must provide input file or list query to get target workflows to reset")
	}

	wg := &sync.WaitGroup{}

	wes := make(chan commonpb.WorkflowExecution)
	done := make(chan bool)
	for i := 0; i < parallel; i++ {
		wg.Add(1)
		go processResets(c, namespace, wes, done, wg, batchResetParams)
	}

	// read exclude
	excludes := map[string]string{}
	if len(excFileName) > 0 {
		// This code is only used in the CLI. The input provided is from a trusted user.
		// #nosec
		excFile, err := os.Open(excFileName)
		if err != nil {
			return fmt.Errorf("unable to read exclude rules: %w", err)
		}
		defer excFile.Close()
		scanner := bufio.NewScanner(excFile)
		idx := 0
		for scanner.Scan() {
			idx++
			line := strings.TrimSpace(scanner.Text())
			if len(line) == 0 {
				fmt.Printf("line %v is empty, skipped\n", idx)
				continue
			}
			cols := strings.Split(line, separator)
			if len(cols) < 1 {
				return fmt.Errorf("exclude file: unable to split, line %v has less than 1 cols separated by comma, only %v", idx, len(cols))
			}
			wid := strings.TrimSpace(cols[0])
			rid := "not-needed"
			excludes[wid] = rid
		}
	}
	fmt.Println("num of excludes:", len(excludes))

	if len(inFileName) > 0 {
		inFile, err := os.Open(inFileName)
		if err != nil {
			return fmt.Errorf("unable to open input file: %w", err)
		}
		defer inFile.Close()
		scanner := bufio.NewScanner(inFile)
		idx := 0
		for scanner.Scan() {
			idx++
			line := strings.TrimSpace(scanner.Text())
			if len(line) == 0 {
				fmt.Printf("line %v is empty, skipped\n", idx)
				continue
			}
			cols := strings.Split(line, separator)
			if len(cols) < 1 {
				return fmt.Errorf("include file: unable to split, line %v has less than 1 cols separated by comma, only %v", idx, len(cols))
			}
			fmt.Printf("Start processing line %v ...\n", idx)
			wid := strings.TrimSpace(cols[0])
			rid := ""
			if len(cols) > 1 {
				rid = strings.TrimSpace(cols[1])
			}

			_, ok := excludes[wid]
			if ok {
				fmt.Println("skip by exclude file: ", wid, rid)
				continue
			}

			wes <- commonpb.WorkflowExecution{
				WorkflowId: wid,
				RunId:      rid,
			}
		}
	} else {
		sdkClient, err := client.GetSDKClient(c)
		if err != nil {
			return err
		}

		var nextPageToken []byte
		var result []any
		for {
			result, nextPageToken, err = listWorkflows(c, sdkClient, nextPageToken, query)
			if err != nil {
				return err
			}
			for _, resultItem := range result {
				we, ok := resultItem.(*workflowpb.WorkflowExecutionInfo)
				if !ok {
					fmt.Printf("skip by wrong type:%T instead of:%T\n", resultItem, &workflowpb.WorkflowExecutionInfo{})
					continue
				}

				wid := we.Execution.GetWorkflowId()
				rid := we.Execution.GetRunId()
				_, ok = excludes[wid]
				if ok {
					fmt.Println("skip by exclude file: ", wid, rid)
					continue
				}

				wes <- commonpb.WorkflowExecution{
					WorkflowId: wid,
					RunId:      rid,
				}
			}

			if nextPageToken == nil {
				break
			}
		}
	}

	close(done)
	fmt.Println("wait for all goroutines...")
	wg.Wait()

	return nil
}

func printErrorAndReturn(msg string, err error) error {
	fmt.Println(msg)
	return err
}

func doReset(c *cli.Context, namespace, wid, rid string, params batchResetParamsType) error {
	ctx, cancel := common.NewContext(c)
	defer cancel()

	frontendClient := client.Factory(c.App).FrontendClient(c)
	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &workflowservice.DescribeWorkflowExecutionRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
		},
	})
	if err != nil {
		return printErrorAndReturn("DescribeWorkflowExecution failed", err)
	}

	currentRunID := resp.WorkflowExecutionInfo.Execution.GetRunId()
	if currentRunID != rid && params.skipBaseNotCurrent {
		fmt.Println("skip because base run is different from current run: ", wid, rid, currentRunID)
		return nil
	}
	if rid == "" {
		rid = currentRunID
	}

	if resp.WorkflowExecutionInfo.GetStatus() == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING || resp.WorkflowExecutionInfo.CloseTime == nil {
		if params.skipOpen {
			fmt.Println("skip because current run is open: ", wid, rid, currentRunID)
			// skip and not terminate current if open
			return nil
		}
	}

	if params.nonDeterministicOnly {
		isLDN, err := isLastEventWorkflowTaskFailedWithNonDeterminism(ctx, namespace, wid, rid, frontendClient)
		if err != nil {
			return printErrorAndReturn("check isLastEventWorkflowTaskFailedWithNonDeterminism failed", err)
		}
		if !isLDN {
			fmt.Println("skip because last event is not WorkflowTaskFailedWithNonDeterminism")
			return nil
		}
	}

	resetBaseRunID, workflowTaskFinishID, err := getResetEventIDByType(ctx, c, params.resetType, namespace, wid, rid, frontendClient)
	if err != nil {
		return printErrorAndReturn("getResetEventIDByType failed", err)
	}
	fmt.Println("WorkflowTaskFinishEventId for reset:", wid, rid, resetBaseRunID, workflowTaskFinishID)

	if params.dryRun {
		fmt.Printf("dry run to reset wid: %v, rid:%v to baseRunId:%v, eventId:%v \n", wid, rid, resetBaseRunID, workflowTaskFinishID)
	} else {
		resp2, err := frontendClient.ResetWorkflowExecution(ctx, &workflowservice.ResetWorkflowExecutionRequest{
			Namespace: namespace,
			WorkflowExecution: &commonpb.WorkflowExecution{
				WorkflowId: wid,
				RunId:      resetBaseRunID,
			},
			WorkflowTaskFinishEventId: workflowTaskFinishID,
			RequestId:                 uuid.New(),
			Reason:                    fmt.Sprintf("%v:%v", common.GetCurrentUserFromEnv(), params.reason),
		})

		if err != nil {
			return printErrorAndReturn("ResetWorkflowExecution failed", err)
		}
		fmt.Println("new runId for wid/rid is ,", wid, rid, resp2.GetRunId())
	}

	return nil
}

func isLastEventWorkflowTaskFailedWithNonDeterminism(ctx context.Context, namespace, wid, rid string, frontendClient workflowservice.WorkflowServiceClient) (bool, error) {
	req := &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}

	var firstEvent, workflowTaskFailedEvent *historypb.HistoryEvent
	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return false, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if firstEvent == nil {
				firstEvent = e
			}
			if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_FAILED {
				workflowTaskFailedEvent = e
			} else if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskFailedEvent = nil
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}

	if workflowTaskFailedEvent != nil {
		attr := workflowTaskFailedEvent.GetWorkflowTaskFailedEventAttributes()

		if (attr.GetCause() == enumspb.WORKFLOW_TASK_FAILED_CAUSE_NON_DETERMINISTIC_ERROR) ||
			(attr.GetCause() == enumspb.WORKFLOW_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE ||
				strings.Contains(attr.GetFailure().GetMessage(), "nondeterministic")) {
			fmt.Printf("found non determnistic workflow wid:%v, rid:%v, orignalStartTime:%v \n", wid, rid, timestamp.TimeValue(firstEvent.GetEventTime()))
			return true, nil
		}
	}

	return false, nil
}

func getResetEventIDByType(ctx context.Context, c *cli.Context, resetType, namespace, wid, rid string, frontendClient workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskFinishID int64, err error) {
	fmt.Println("resetType:", resetType)
	switch resetType {
	case "LastWorkflowTask":
		resetBaseRunID, workflowTaskFinishID, err = getLastWorkflowTaskEventID(ctx, namespace, wid, rid, frontendClient)
		if err != nil {
			return
		}
	case "LastContinuedAsNew":
		resetBaseRunID, workflowTaskFinishID, err = getLastContinueAsNewID(ctx, namespace, wid, rid, frontendClient)
		if err != nil {
			return
		}
	case "FirstWorkflowTask":
		resetBaseRunID, workflowTaskFinishID, err = getFirstWorkflowTaskEventID(ctx, namespace, wid, rid, frontendClient)
		if err != nil {
			return
		}
	default:
		panic("not supported resetType")
	}
	return
}

// Returns event id of the last completed task or id of the next event after scheduled task.
func getLastWorkflowTaskEventID(ctx context.Context, namespace, wid, rid string, frontendClient workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskEventID int64, err error) {
	resetBaseRunID = rid
	req := &workflowservice.GetWorkflowExecutionHistoryReverseRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}

	for {
		resp, err := frontendClient.GetWorkflowExecutionHistoryReverse(ctx, req)
		if err != nil {
			return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskEventID = e.GetEventId()
				break
			} else if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED {
				// if there is no task completed event, set it to first scheduled event + 1
				workflowTaskEventID = e.GetEventId() + 1
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if workflowTaskEventID == 0 {
		return "", 0, printErrorAndReturn("Get LastWorkflowTaskID failed", fmt.Errorf("unable to find any scheduled or completed task"))
	}
	return
}

// Returns id of the first workflow task completed event or if it doesn't exist then id of the event after task scheduled event.
func getFirstWorkflowTaskEventID(ctx context.Context, namespace, wid, rid string, frontendClient workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskEventID int64, err error) {
	resetBaseRunID = rid
	req := &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}
	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskEventID = e.GetEventId()
				return resetBaseRunID, workflowTaskEventID, nil
			}
			if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED {
				if workflowTaskEventID == 0 {
					workflowTaskEventID = e.GetEventId() + 1
				}
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if workflowTaskEventID == 0 {
		return "", 0, printErrorAndReturn("Get FirstWorkflowTaskID failed", fmt.Errorf("unable to find any scheduled or completed task"))
	}
	return
}

func getLastContinueAsNewID(ctx context.Context, namespace, wid, rid string, frontendClient workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskCompletedID int64, err error) {
	// get first event
	req := &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		MaximumPageSize: 1,
		NextPageToken:   nil,
	}
	resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
	if err != nil {
		return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
	}
	firstEvent := resp.History.Events[0]
	resetBaseRunID = firstEvent.GetWorkflowExecutionStartedEventAttributes().GetContinuedExecutionRunId()
	if resetBaseRunID == "" {
		return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", fmt.Errorf("cannot get resetBaseRunId"))
	}

	req = &workflowservice.GetWorkflowExecutionHistoryRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      resetBaseRunID,
		},
		MaximumPageSize: 1000,
		NextPageToken:   nil,
	}
	for {
		resp, err := frontendClient.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return "", 0, printErrorAndReturn("GetWorkflowExecutionHistory failed", err)
		}
		for _, e := range resp.GetHistory().GetEvents() {
			if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED {
				workflowTaskCompletedID = e.GetEventId()
			}
		}
		if len(resp.NextPageToken) != 0 {
			req.NextPageToken = resp.NextPageToken
		} else {
			break
		}
	}
	if workflowTaskCompletedID == 0 {
		return "", 0, printErrorAndReturn("Get LastContinueAsNewID failed", fmt.Errorf("no WorkflowTaskCompletedID"))
	}
	return
}

func listWorkflows(c *cli.Context, sdkClient sdkclient.Client, npt []byte, query string) ([]interface{}, []byte, error) {
	req := &workflowservice.ListWorkflowExecutionsRequest{
		NextPageToken: npt,
		Query:         query,
	}

	var workflows *workflowservice.ListWorkflowExecutionsResponse
	op := func() error {
		ctx, cancel := common.NewContext(c)
		defer cancel()
		resp, err := sdkClient.ListWorkflow(ctx, req)
		if err != nil {
			return err
		}
		workflows = resp
		return nil
	}
	err := backoff.ThrottleRetry(op, scommon.CreateFrontendClientRetryPolicy(), scommon.IsContextDeadlineExceededErr)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list workflow executions: %w", err)
	}

	var items []interface{}
	for _, e := range workflows.Executions {
		items = append(items, e)
	}

	return items, workflows.NextPageToken, nil
}

func listArchivedWorkflows(c *cli.Context, sdkClient sdkclient.Client, npt []byte, query string) ([]interface{}, []byte, error) {
	req := &workflowservice.ListArchivedWorkflowExecutionsRequest{
		NextPageToken: npt,
		Query:         query,
	}

	contextTimeout := common.DefaultContextTimeoutForListArchivedWorkflow
	if c.IsSet(common.FlagContextTimeout) {
		contextTimeout = time.Duration(c.Int(common.FlagContextTimeout)) * time.Second
	}

	var workflows *workflowservice.ListArchivedWorkflowExecutionsResponse
	op := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)

		defer cancel()
		resp, err := sdkClient.ListArchivedWorkflow(ctx, req)
		if err != nil {
			return err
		}
		workflows = resp
		return nil
	}
	err := backoff.ThrottleRetry(op, scommon.CreateFrontendClientRetryPolicy(), scommon.IsContextDeadlineExceededErr)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list archived workflow executions: %w", err)
	}

	var items []interface{}
	for _, e := range workflows.Executions {
		items = append(items, e)
	}

	return items, workflows.NextPageToken, nil
}

func TraceWorkflow(c *cli.Context) error {
	foldStatus, err := ParseFoldStatusList(c.String(common.FlagFold))
	if err != nil {
		return err
	}
	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)
	_, err = trace.PrintWorkflowTrace(c, wid, rid, foldStatus)
	return err
}

func UpdateWorkflow(c *cli.Context) error {
	wid := c.String(common.FlagWorkflowID)
	rid := c.String(common.FlagRunID)
	name := c.String(common.FlagName)
	firstExecutionRunID := c.String(common.FlagUpdateFirstExecutionRunID)
	args, err := common.UnmarshalInputsFromCLI(c)
	if err != nil {
		return err
	}
	request := sdkclient.UpdateWorkflowWithOptionsRequest{
		WorkflowID:          wid,
		RunID:               rid,
		UpdateName:          name,
		Args:                args,
		FirstExecutionRunID: firstExecutionRunID,
	}
	return updateWorkflowHelper(c, &request)

}

func updateWorkflowHelper(c *cli.Context, request *sdkclient.UpdateWorkflowWithOptionsRequest) error {
	ctx, cancel := common.NewContext(c)
	defer cancel()
	sdk, err := client.GetSDKClient(c)
	if err != nil {
		return err
	}

	workflowUpdateHandle, err := sdk.UpdateWorkflowWithOptions(ctx, request)

	if err != nil {
		return fmt.Errorf("unable to update workflow: %w", err)
	}

	var valuePtr interface{}
	err = workflowUpdateHandle.Get(ctx, &valuePtr)
	if err != nil {
		return fmt.Errorf("unable to update workflow: %w", err)
	}
	result := map[string]interface{}{
		"Name":     request.UpdateName,
		"UpdateID": workflowUpdateHandle.UpdateID(),
		"Result":   valuePtr,
	}

	common.PrettyPrintJSONObject(c, result)
	return nil
}

// this only works for ANSI terminal, which means remove existing lines won't work if users redirect to file
// ref: https://en.wikipedia.org/wiki/ANSI_escape_code
func removePrevious2LinesFromTerminal() {
	fmt.Printf("\033[1A")
	fmt.Printf("\033[2K")
	fmt.Printf("\033[1A")
	fmt.Printf("\033[2K")
}

func mapKeysToArray(m map[string]interface{}) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	return out
}

func ParseFoldStatusList(flagValue string) ([]enumspb.WorkflowExecutionStatus, error) {
	var statusList []enumspb.WorkflowExecutionStatus
	for _, value := range strings.Split(flagValue, ",") {
		if status, ok := findWorkflowStatusValue(value); ok {
			statusList = append(statusList, status)
		} else {
			return nil,
				fmt.Errorf("invalid status \"%s\" for fold flag. Valid values: %v", value, listWorkflowExecutionStatusNames())
		}
	}
	return statusList, nil
}

func listWorkflowExecutionStatusNames() string {
	var names []string
	for _, name := range enumspb.WorkflowExecutionStatus_name {
		names = append(names, strings.ToLower(name))
	}
	return strings.Join(names, ", ")
}

// findWorkflowStatusValue finds a WorkflowExecutionStatus by its name. This search is case-insensitive.
func findWorkflowStatusValue(name string) (enumspb.WorkflowExecutionStatus, bool) {
	lowerName := strings.ToLower(name)
	for key, value := range enumspb.WorkflowExecutionStatus_value {
		if lowerName == strings.ToLower(key) {
			return enumspb.WorkflowExecutionStatus(value), true
		}
	}

	return 0, false
}

// historyEventToString convert HistoryEvent to string
//
//revive:disable:flag-parameter
func historyEventToString(e *historypb.HistoryEvent, printFully bool, maxFieldLength int) string {
	data := getEventAttributes(e)
	return stringify.AnyToString(data, printFully, maxFieldLength)
}

func getEventAttributes(e *historypb.HistoryEvent) interface{} {
	var data interface{}
	switch e.GetEventType() {
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED:
		data = e.GetWorkflowExecutionStartedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		data = e.GetWorkflowExecutionCompletedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		data = e.GetWorkflowExecutionFailedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_FAILED:
		data = e.GetWorkflowTaskFailedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		data = e.GetWorkflowExecutionTimedOutEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED:
		data = e.GetWorkflowTaskScheduledEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_STARTED:
		data = e.GetWorkflowTaskStartedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_COMPLETED:
		data = e.GetWorkflowTaskCompletedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_TASK_TIMED_OUT:
		data = e.GetWorkflowTaskTimedOutEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED:
		data = e.GetActivityTaskScheduledEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_STARTED:
		data = e.GetActivityTaskStartedEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_COMPLETED:
		data = e.GetActivityTaskCompletedEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_FAILED:
		data = e.GetActivityTaskFailedEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_TIMED_OUT:
		data = e.GetActivityTaskTimedOutEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_CANCEL_REQUESTED:
		data = e.GetActivityTaskCancelRequestedEventAttributes()

	case enumspb.EVENT_TYPE_ACTIVITY_TASK_CANCELED:
		data = e.GetActivityTaskCanceledEventAttributes()

	case enumspb.EVENT_TYPE_TIMER_STARTED:
		data = e.GetTimerStartedEventAttributes()

	case enumspb.EVENT_TYPE_TIMER_FIRED:
		data = e.GetTimerFiredEventAttributes()

	case enumspb.EVENT_TYPE_TIMER_CANCELED:
		data = e.GetTimerCanceledEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCEL_REQUESTED:
		data = e.GetWorkflowExecutionCancelRequestedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		data = e.GetWorkflowExecutionCanceledEventAttributes()

	case enumspb.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_INITIATED:
		data = e.GetRequestCancelExternalWorkflowExecutionInitiatedEventAttributes()

	case enumspb.EVENT_TYPE_REQUEST_CANCEL_EXTERNAL_WORKFLOW_EXECUTION_FAILED:
		data = e.GetRequestCancelExternalWorkflowExecutionFailedEventAttributes()

	case enumspb.EVENT_TYPE_EXTERNAL_WORKFLOW_EXECUTION_CANCEL_REQUESTED:
		data = e.GetExternalWorkflowExecutionCancelRequestedEventAttributes()

	case enumspb.EVENT_TYPE_MARKER_RECORDED:
		data = e.GetMarkerRecordedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_SIGNALED:
		data = e.GetWorkflowExecutionSignaledEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TERMINATED:
		data = e.GetWorkflowExecutionTerminatedEventAttributes()

	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CONTINUED_AS_NEW:
		data = e.GetWorkflowExecutionContinuedAsNewEventAttributes()

	case enumspb.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_INITIATED:
		data = e.GetStartChildWorkflowExecutionInitiatedEventAttributes()

	case enumspb.EVENT_TYPE_START_CHILD_WORKFLOW_EXECUTION_FAILED:
		data = e.GetStartChildWorkflowExecutionFailedEventAttributes()

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_STARTED:
		data = e.GetChildWorkflowExecutionStartedEventAttributes()

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_COMPLETED:
		data = e.GetChildWorkflowExecutionCompletedEventAttributes()

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_FAILED:
		data = e.GetChildWorkflowExecutionFailedEventAttributes()

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_CANCELED:
		data = e.GetChildWorkflowExecutionCanceledEventAttributes()

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TIMED_OUT:
		data = e.GetChildWorkflowExecutionTimedOutEventAttributes()

	case enumspb.EVENT_TYPE_CHILD_WORKFLOW_EXECUTION_TERMINATED:
		data = e.GetChildWorkflowExecutionTerminatedEventAttributes()

	case enumspb.EVENT_TYPE_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_INITIATED:
		data = e.GetSignalExternalWorkflowExecutionInitiatedEventAttributes()

	case enumspb.EVENT_TYPE_SIGNAL_EXTERNAL_WORKFLOW_EXECUTION_FAILED:
		data = e.GetSignalExternalWorkflowExecutionFailedEventAttributes()

	case enumspb.EVENT_TYPE_EXTERNAL_WORKFLOW_EXECUTION_SIGNALED:
		data = e.GetExternalWorkflowExecutionSignaledEventAttributes()

	case enumspb.EVENT_TYPE_UPSERT_WORKFLOW_SEARCH_ATTRIBUTES:
		data = e.GetUpsertWorkflowSearchAttributesEventAttributes()

	default:
		data = e
	}
	return data
}
