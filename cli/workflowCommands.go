// The MIT License
//
// Copyright (c) 2020 Temporal Technologies Inc.  All rights reserved.
//
// Copyright (c) 2020 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cli

import (
	"bufio"
	"context"
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/pborman/uuid"
	"github.com/urfave/cli/v2"
	"github.com/valyala/fastjson"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	failurepb "go.temporal.io/api/failure/v1"
	filterpb "go.temporal.io/api/filter/v1"
	historypb "go.temporal.io/api/history/v1"
	namespacepb "go.temporal.io/api/namespace/v1"
	querypb "go.temporal.io/api/query/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"

	"github.com/temporalio/shared-go/collection"
	"github.com/temporalio/shared-go/timestamp"
	"github.com/temporalio/tctl/common/payload"
	"github.com/temporalio/tctl/common/payloads"
	"github.com/temporalio/tctl/terminal"
	clispb "go.temporal.io/server/api/cli/v1"
	"go.temporal.io/server/common/clock"
	"go.temporal.io/server/common/convert"
	"go.temporal.io/server/common/searchattribute"
	"go.temporal.io/server/service/history"
)

// ShowHistory shows the history of given workflow execution based on workflowID and runID.
func ShowHistory(c *cli.Context) {
	wid, rid := getWorkflowParams(c)

	namespace := getRequiredGlobalOption(c, FlagNamespace)
	var maxFieldLength int
	if c.IsSet(FlagMaxFieldLength) || true {
		maxFieldLength = c.Int(FlagMaxFieldLength)
	}
	client := cFactory.FrontendClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		ctx, cancel := newContext(c)
		defer cancel()
		var err error

		req := &workflowservice.GetWorkflowExecutionHistoryRequest{
			Namespace: namespace,
			Execution: &commonpb.WorkflowExecution{
				WorkflowId: wid,
				RunId:      rid,
			},
			NextPageToken:          npt,
			HistoryEventFilterType: enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT,
		}
		res, err := client.GetWorkflowExecutionHistory(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		var items []interface{}
		for _, e := range res.History.Events {
			item := eventRow{
				ID:      convert.Int64ToString(e.GetEventId()),
				Type:    ColorEvent(e),
				Details: HistoryEventToString(e, false, maxFieldLength),
			}
			items = append(items, item)
		}
		if err != nil {
			return nil, nil, err
		}

		return items, res.NextPageToken, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &terminal.PaginateOptions{Fields: []string{"ID", "Type", "Event"}}
	terminal.Paginate(c, iter, opts)
}

// RunWorkflow starts a new workflow execution and print workflow progress and result
func RunWorkflow(c *cli.Context) {
	serviceClient := cFactory.FrontendClient(c)

	namespace := getRequiredGlobalOption(c, FlagNamespace)
	taskQueue := getRequiredOption(c, FlagTaskQueue)
	workflowType := getRequiredOption(c, FlagWorkflowType)
	et := c.Int(FlagExecutionTimeout)
	dt := c.Int(FlagWorkflowTaskTimeout)
	wid := c.String(FlagWorkflowID)
	if len(wid) == 0 {
		wid = uuid.New()
	}
	reusePolicy := defaultWorkflowIDReusePolicy
	if c.IsSet(FlagWorkflowIDReusePolicy) {
		reusePolicyInt, err := stringToEnum(c.String(FlagWorkflowIDReusePolicy), enumspb.WorkflowIdReusePolicy_value)
		if err != nil {
			ErrorAndExit("Failed to parse Reuse Policy", err)
		}
		reusePolicy = enumspb.WorkflowIdReusePolicy(reusePolicyInt)
	}

	input := processJSONInput(c)
	startRequest := &workflowservice.StartWorkflowExecutionRequest{
		RequestId:  uuid.New(),
		Namespace:  namespace,
		WorkflowId: wid,
		WorkflowType: &commonpb.WorkflowType{
			Name: workflowType,
		},
		TaskQueue: &taskqueuepb.TaskQueue{
			Name: taskQueue,
			Kind: enumspb.TASK_QUEUE_KIND_NORMAL,
		},
		Input:                    input,
		WorkflowExecutionTimeout: timestamp.DurationPtr(time.Duration(et) * time.Second),
		WorkflowTaskTimeout:      timestamp.DurationPtr(time.Duration(dt) * time.Second),
		Identity:                 getCliIdentity(),
		WorkflowIdReusePolicy:    reusePolicy,
	}
	if c.IsSet(FlagCronSchedule) {
		startRequest.CronSchedule = c.String(FlagCronSchedule)
	}

	memoFields := processMemo(c)
	if len(memoFields) != 0 {
		startRequest.Memo = &commonpb.Memo{Fields: memoFields}
	}

	startRequest.SearchAttributes = processSearchAttr(c)

	tcCtx, cancel := newContextForLongPoll(c)
	defer cancel()
	resp, err := serviceClient.StartWorkflowExecution(tcCtx, startRequest)

	if err != nil {
		ErrorAndExit("Failed to run workflow.", err)
	}

	type row struct {
		Field string
		Value string
	}

	executionData := []interface{}{
		&row{Field: "Workflow Id", Value: wid},
		&row{Field: "Run Id", Value: resp.GetRunId()},
		&row{Field: "Type", Value: workflowType},
		&row{Field: "Namespace", Value: namespace},
		&row{Field: "Task Queue", Value: taskQueue},
		&row{Field: "Args", Value: truncate(payloads.ToString(input))},
	}
	fmt.Println(colorMagenta("Running execution:"))
	opts := &terminal.PrintOptions{Fields: []string{"Field", "Value"}, Header: false}
	terminal.PrintItems(c, executionData, opts)

	printWorkflowProgress(c, wid, resp.GetRunId())
}

func processSearchAttr(c *cli.Context) *commonpb.SearchAttributes {
	sanitize := func(val string) []string {
		trimmedVal := strings.TrimSpace(val)
		if len(trimmedVal) == 0 {
			return nil
		}
		splitVal := strings.Split(trimmedVal, searchAttrInputSeparator)
		result := make([]string, len(splitVal))
		for i, v := range splitVal {
			result[i] = strings.TrimSpace(v)
		}
		return result
	}

	searchAttrKeys := sanitize(c.String(FlagSearchAttributesKey))
	if len(searchAttrKeys) == 0 {
		return nil
	}
	searchAttrVals := sanitize(c.String(FlagSearchAttributesVal))
	if len(searchAttrVals) == 0 {
		return nil
	}

	if len(searchAttrKeys) != len(searchAttrVals) {
		ErrorAndExit(fmt.Sprintf("Uneven number of search attributes keys (%d): %v and values(%d): %v.", len(searchAttrKeys), searchAttrKeys, len(searchAttrVals), searchAttrVals), nil)
	}

	searchAttributesStr := make(map[string]string, len(searchAttrKeys))
	for i, searchAttrVal := range searchAttrVals {
		searchAttributesStr[searchAttrKeys[i]] = searchAttrVal
	}

	searchAttributes, err := searchattribute.Parse(searchAttributesStr, nil)
	if err != nil {
		ErrorAndExit("Unable to parse search attributes.", err)
	}

	return searchAttributes
}

func processMemo(c *cli.Context) map[string]*commonpb.Payload {
	rawMemoKey := c.String(FlagMemoKey)
	var memoKeys []string
	if strings.TrimSpace(rawMemoKey) != "" {
		memoKeys = strings.Split(rawMemoKey, " ")
	}

	jsonsRaw := readJSONInputs(c, jsonTypeMemo)
	if len(jsonsRaw) == 0 {
		return nil
	}
	rawMemoValue := string(jsonsRaw[0]) // StringFlag may contain up to one json

	if err := validateJSONs(rawMemoValue); err != nil {
		ErrorAndExit("Input is not valid JSON, or JSONs concatenated with spaces/newlines.", err)
	}

	var memoValues []string

	var sc fastjson.Scanner
	sc.Init(rawMemoValue)
	for sc.Next() {
		memoValues = append(memoValues, sc.Value().String())
	}
	if err := sc.Error(); err != nil {
		ErrorAndExit("Parse json error.", err)
	}
	if len(memoKeys) != len(memoValues) {
		ErrorAndExit("Number of memo keys and values are not equal.", nil)
	}

	fields := map[string]*commonpb.Payload{}
	for i, key := range memoKeys {
		fields[key] = payload.EncodeString(memoValues[i])
	}
	return fields
}

type historyIterator struct {
	iter interface {
		HasNext() bool
		Next() (*historypb.HistoryEvent, error)
	}
	maxFieldLength int
	showDetails    bool
	lastEvent      *historypb.HistoryEvent
}

func (h *historyIterator) HasNext() bool {
	return h.iter.HasNext()
}

func (h *historyIterator) Next() (interface{}, error) {
	event, err := h.iter.Next()
	if err != nil {
		return nil, err
	}

	reflect.ValueOf(h.lastEvent).Elem().Set(reflect.ValueOf(event).Elem())

	return eventRow{
		ID:      convert.Int64ToString(event.GetEventId()),
		Time:    formatTime(timestamp.TimeValue(event.GetEventTime()), false),
		Type:    ColorEvent(event),
		Details: HistoryEventToString(event, false, h.maxFieldLength),
	}, nil
}

// helper function to print workflow progress with time refresh every second
func printWorkflowProgress(c *cli.Context, wid, rid string) {
	showDetails := c.Bool(FlagShowDetail)
	var maxFieldLength int
	if c.IsSet(FlagMaxFieldLength) {
		maxFieldLength = c.Int(FlagMaxFieldLength)
	}
	sdkClient := getSDKClient(c)

	tcCtx, cancel := newIndefiniteContext(c)
	defer cancel()

	doneChan := make(chan bool)
	timeElapse := 1
	isTimeElapseExist := false
	ticker := time.NewTicker(time.Second).C
	opts := &terminal.PaginateOptions{
		Fields: []string{"ID", "Time", "Type"},
		All:    true,
	}
	if showDetails {
		opts.Fields = append(opts.Fields, "Details")
	}
	fmt.Println(colorMagenta("Progress:"))
	var lastEvent historypb.HistoryEvent // used for print result of this run

	go func() {
		hIter := sdkClient.GetWorkflowHistory(tcCtx, wid, rid, true, enumspb.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		iter := &historyIterator{iter: hIter, maxFieldLength: maxFieldLength, showDetails: showDetails, lastEvent: &lastEvent}
		terminal.Paginate(c, iter, opts)

		doneChan <- true
	}()

	for {
		select {
		case <-ticker:
			if isTimeElapseExist {
				removePrevious2LinesFromTerminal()
			}
			fmt.Printf("\nTime elapse: %ds\n", timeElapse)
			isTimeElapseExist = true
			timeElapse++
		case <-doneChan: // print result of this run
			fmt.Println(colorMagenta("\nResult:"))
			fmt.Printf("  Run Time: %d seconds\n", timeElapse)
			printRunStatus(&lastEvent)
			return
		}
	}
}

// TerminateWorkflow terminates a workflow execution
func TerminateWorkflow(c *cli.Context) {
	sdkClient := getSDKClient(c)

	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	reason := c.String(FlagReason)

	ctx, cancel := newContext(c)
	defer cancel()
	err := sdkClient.TerminateWorkflow(ctx, wid, rid, reason, nil)

	if err != nil {
		ErrorAndExit("Terminate workflow failed.", err)
	} else {
		fmt.Println("Terminate workflow succeeded.")
	}
}

// CancelWorkflow cancels a workflow execution
func CancelWorkflow(c *cli.Context) {
	sdkClient := getSDKClient(c)

	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)

	ctx, cancel := newContext(c)
	defer cancel()
	err := sdkClient.CancelWorkflow(ctx, wid, rid)

	if err != nil {
		ErrorAndExit("Cancel workflow failed.", err)
	} else {
		fmt.Println("Cancel workflow succeeded.")
	}
}

// SignalWorkflow signals a workflow execution
func SignalWorkflow(c *cli.Context) {
	serviceClient := cFactory.FrontendClient(c)

	namespace := getRequiredGlobalOption(c, FlagNamespace)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	name := getRequiredOption(c, FlagName)
	input := processJSONInput(c)

	tcCtx, cancel := newContext(c)
	defer cancel()
	_, err := serviceClient.SignalWorkflowExecution(tcCtx, &workflowservice.SignalWorkflowExecutionRequest{
		Namespace: namespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
		SignalName: name,
		Input:      input,
		Identity:   getCliIdentity(),
	})

	if err != nil {
		ErrorAndExit("Signal workflow failed.", err)
	} else {
		fmt.Println("Signal workflow succeeded.")
	}
}

// QueryWorkflow query workflow execution
func QueryWorkflow(c *cli.Context) {
	getRequiredGlobalOption(c, FlagNamespace) // for pre-check and alert if not provided
	getRequiredOption(c, FlagWorkflowID)
	queryType := getRequiredOption(c, FlagQueryType)

	queryWorkflowHelper(c, queryType)
}

// QueryWorkflowUsingStackTrace query workflow execution using __stack_trace as query type
func QueryWorkflowUsingStackTrace(c *cli.Context) {
	queryWorkflowHelper(c, "__stack_trace")
}

func queryWorkflowHelper(c *cli.Context, queryType string) {
	serviceClient := cFactory.FrontendClient(c)

	namespace := getRequiredGlobalOption(c, FlagNamespace)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := c.String(FlagRunID)
	input := processJSONInput(c)

	tcCtx, cancel := newContext(c)
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
	if c.IsSet(FlagQueryRejectCondition) {
		var rejectCondition enumspb.QueryRejectCondition
		switch c.String(FlagQueryRejectCondition) {
		case "not_open":
			rejectCondition = enumspb.QUERY_REJECT_CONDITION_NOT_OPEN
		case "not_completed_cleanly":
			rejectCondition = enumspb.QUERY_REJECT_CONDITION_NOT_COMPLETED_CLEANLY
		default:
			ErrorAndExit(fmt.Sprintf("invalid reject condition %v, valid values are \"not_open\" and \"not_completed_cleanly\"", c.String(FlagQueryRejectCondition)), nil)
		}
		queryRequest.QueryRejectCondition = rejectCondition
	}
	queryResponse, err := serviceClient.QueryWorkflow(tcCtx, queryRequest)
	if err != nil {
		ErrorAndExit("Query workflow failed.", err)
		return
	}

	if queryResponse.QueryRejected != nil {
		fmt.Printf("Query was rejected, workflow has status: %v\n", queryResponse.QueryRejected.GetStatus())
	} else {
		queryResult := payloads.ToString(queryResponse.QueryResult)
		fmt.Printf("Query result:\n%v\n", queryResult)
	}
}

// ListWorkflow list workflow executions based on filters
func ListWorkflow(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	queryOpen := c.Bool(FlagOpen)
	workflowID := c.String(FlagWorkflowID)
	workflowType := c.String(FlagWorkflowType)
	earliestTime := parseTime(c.String(FlagEarliestTime), time.Time{}, time.Now().UTC())
	latestTime := parseTime(c.String(FlagLatestTime), time.Now().UTC(), time.Now().UTC())
	wfStatusInt, err := stringToEnum(c.String(FlagWorkflowStatus), enumspb.WorkflowExecutionStatus_value)
	if err != nil {
		ErrorAndExit("Failed to parse Workflow Status", err)
	}
	wfStatus := enumspb.WorkflowExecutionStatus(wfStatusInt)
	client := cFactory.FrontendClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		ctx, cancel := newContextForLongPoll(c)
		defer cancel()
		var items []interface{}
		var err error
		if c.IsSet(FlagListQuery) {
			query := c.String(FlagListQuery)
			items, npt, err = listWorkflows(ctx, client, npt, namespace, query)
		} else if queryOpen {
			items, npt, err = listOpenWorkflows(ctx, client, npt, namespace, earliestTime, latestTime, workflowID, workflowType)
		} else {
			items, npt, err = listClosedWorkflows(ctx, client, npt, namespace, earliestTime, latestTime, workflowID, workflowType, wfStatus)
		}
		if err != nil {
			return nil, nil, err
		}

		return items, npt, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &terminal.PaginateOptions{Fields: []string{"Type.Name", "Execution.WorkflowId", "Execution.RunId", "TaskQueue", "StartTime", "ExecutionTime", "CloseTime"}}
	terminal.Paginate(c, iter, opts)
}

// ScanAllWorkflow list all workflow executions using Scan API.
func ScanAllWorkflow(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	listQuery := c.String(FlagListQuery)
	client := cFactory.FrontendClient(c)

	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		ctx, cancel := newContextForLongPoll(c)
		defer cancel()
		var err error

		req := &workflowservice.ScanWorkflowExecutionsRequest{
			Namespace:     namespace,
			NextPageToken: npt,
			Query:         listQuery,
		}

		resp, err := client.ScanWorkflowExecutions(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		var items []interface{}
		for _, e := range resp.Executions {
			items = append(items, e)
		}
		if err != nil {
			return nil, nil, err
		}

		return items, resp.NextPageToken, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &terminal.PaginateOptions{Fields: []string{"Type.Name", "Execution.WorkflowId", "Execution.RunId", "TaskQueue", "StartTime", "ExecutionTime", "CloseTime"}}
	terminal.Paginate(c, iter, opts)
}

// CountWorkflow count number of workflows
func CountWorkflow(c *cli.Context) {
	client := cFactory.FrontendClient(c)

	request := &workflowservice.CountWorkflowExecutionsRequest{
		Namespace: getRequiredGlobalOption(c, FlagNamespace),
		Query:     c.String(FlagListQuery),
	}

	ctx, cancel := newContextForVisibility(c)
	defer cancel()
	response, err := client.CountWorkflowExecutions(ctx, request)
	if err != nil {
		ErrorAndExit("Unable to count workflow.", err)
	}

	fmt.Println(response.GetCount())
}

// ListArchivedWorkflow lists archived workflow executions based on filters
func ListArchivedWorkflow(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	query := getRequiredOption(c, FlagListQuery)
	contextTimeout := defaultContextTimeoutForListArchivedWorkflow
	if c.IsSet(FlagContextTimeout) {
		contextTimeout = time.Duration(c.Int(FlagContextTimeout)) * time.Second
	}

	client := cFactory.FrontendClient(c)
	req := &workflowservice.ListArchivedWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     query,
	}
	var resp *workflowservice.ListArchivedWorkflowExecutionsResponse
	var err error
	paginationFunc := func(npt []byte) ([]interface{}, []byte, error) {
		// the executions will be empty if the query is still running before timeout
		// so keep calling the API until some results are returned (query completed)
		req.NextPageToken = npt
		for resp == nil || (len(resp.Executions) == 0 && resp.NextPageToken != nil) {
			ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
			resp, err = client.ListArchivedWorkflowExecutions(ctx, req)
			if err != nil {
				cancel()
				ErrorAndExit("Failed to list archived workflow.", err)
			}
			cancel()
		}

		var items []interface{}
		for _, e := range resp.Executions {
			items = append(items, e)
		}

		return items, resp.NextPageToken, nil
	}

	iter := collection.NewPagingIterator(paginationFunc)
	opts := &terminal.PaginateOptions{Fields: []string{"Type.Name", "Execution.WorkflowId", "Execution.RunId", "TaskQueue", "StartTime", "ExecutionTime", "CloseTime"}}
	terminal.Paginate(c, iter, opts)
}

// DescribeWorkflow show information about the specified workflow execution
func DescribeWorkflow(c *cli.Context) {
	wid, rid := getWorkflowParams(c)

	frontendClient := cFactory.FrontendClient(c)
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	printRaw := c.Bool(FlagPrintRaw) // printRaw is false by default,
	// and will show datetime and decoded search attributes instead of raw timestamp and byte arrays
	printResetPointsOnly := c.Bool(FlagResetPointsOnly)

	ctx, cancel := newContext(c)
	defer cancel()

	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &workflowservice.DescribeWorkflowExecutionRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
	})
	if err != nil {
		ErrorAndExit("Describe workflow execution failed", err)
	}

	if printResetPointsOnly {
		printAutoResetPoints(resp)
		return
	}

	if printRaw {
		prettyPrintJSONObject(resp)
	} else {
		prettyPrintJSONObject(convertDescribeWorkflowExecutionResponse(resp))
	}
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

func convertDescribeWorkflowExecutionResponse(resp *workflowservice.DescribeWorkflowExecutionResponse) *clispb.DescribeWorkflowExecutionResponse {

	info := resp.GetWorkflowExecutionInfo()
	executionInfo := &clispb.WorkflowExecutionInfo{
		Execution:            info.GetExecution(),
		Type:                 info.GetType(),
		CloseTime:            info.GetCloseTime(),
		StartTime:            info.GetStartTime(),
		Status:               info.GetStatus(),
		HistoryLength:        info.GetHistoryLength(),
		ParentNamespaceId:    info.GetParentNamespaceId(),
		ParentExecution:      info.GetParentExecution(),
		Memo:                 info.GetMemo(),
		SearchAttributes:     convertSearchAttributes(info.GetSearchAttributes()),
		AutoResetPoints:      info.GetAutoResetPoints(),
		StateTransitionCount: info.GetStateTransitionCount(),
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
			pendingActivityStr.HeartbeatDetails = payloads.ToString(pendingActivity.GetHeartbeatDetails())
		}
		pendingActivitiesStr = append(pendingActivitiesStr, pendingActivityStr)
	}

	return &clispb.DescribeWorkflowExecutionResponse{
		ExecutionConfig:       resp.ExecutionConfig,
		WorkflowExecutionInfo: executionInfo,
		PendingActivities:     pendingActivitiesStr,
		PendingChildren:       resp.PendingChildren,
	}
}

func convertSearchAttributes(searchAttributes *commonpb.SearchAttributes) *clispb.SearchAttributes {
	if len(searchAttributes.GetIndexedFields()) == 0 {
		return nil
	}

	fields, err := searchattribute.Stringify(searchAttributes, nil)
	if err != nil {
		fmt.Printf("%s: unable to stringify search attribute: %v\n",
			colorMagenta("Warning"),
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

func printRunStatus(event *historypb.HistoryEvent) {
	switch event.GetEventType() {
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_COMPLETED:
		fmt.Printf("  Status: %s\n", colorGreen("COMPLETED"))
		result := payloads.ToString(event.GetWorkflowExecutionCompletedEventAttributes().GetResult())
		fmt.Printf("  Output: %s\n", result)
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_FAILED:
		fmt.Printf("  Status: %s\n", colorRed("FAILED"))
		fmt.Printf("  Failure: %s\n", convertFailure(event.GetWorkflowExecutionFailedEventAttributes().GetFailure()).String())
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_TIMED_OUT:
		fmt.Printf("  Status: %s\n", colorRed("TIMEOUT"))
		fmt.Printf("  Retry status: %s\n", event.GetWorkflowExecutionTimedOutEventAttributes().GetRetryState())
	case enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_CANCELED:
		fmt.Printf("  Status: %s\n", colorRed("CANCELED"))
		details := payloads.ToString(event.GetWorkflowExecutionCanceledEventAttributes().GetDetails())
		fmt.Printf("  Detail: %s\n", details)
	}
}

func scanWorkflowExecutions(sdkClient sdkclient.Client, pageSize int, nextPageToken []byte, query string, c *cli.Context) ([]*workflowpb.WorkflowExecutionInfo, []byte) {

	request := &workflowservice.ScanWorkflowExecutionsRequest{
		PageSize:      int32(pageSize),
		NextPageToken: nextPageToken,
		Query:         query,
	}

	ctx, cancel := newContextForVisibility(c)
	defer cancel()
	response, err := sdkClient.ScanWorkflow(ctx, request)
	if err != nil {
		ErrorAndExit("Failed to list workflow.", err)
	}
	return response.Executions, response.NextPageToken
}

// ObserveHistory show the process of running workflow
func ObserveHistory(c *cli.Context) {
	wid, rid := getWorkflowParams(c)

	printWorkflowProgress(c, wid, rid)
}

// ResetWorkflow reset workflow
func ResetWorkflow(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	wid := getRequiredOption(c, FlagWorkflowID)
	reason := getRequiredOption(c, FlagReason)
	if len(reason) == 0 {
		ErrorAndExit("wrong reason", fmt.Errorf("reason cannot be empty"))
	}
	rid := c.String(FlagRunID)
	eventID := c.Int64(FlagEventID)
	resetType := c.String(FlagResetType)
	extraForResetType, ok := resetTypesMap[resetType]
	if !ok && eventID <= 0 {
		ErrorAndExit(fmt.Sprintf("must specify either valid event_id or reset_type (one of %s)", strings.Join(mapKeysToArray(resetTypesMap), ", ")), nil)
	}
	if ok && len(extraForResetType) > 0 {
		getRequiredOption(c, extraForResetType)
	}

	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.FrontendClient(c)

	resetBaseRunID := rid
	workflowTaskFinishID := eventID
	var err error
	if resetType != "" {
		resetBaseRunID, workflowTaskFinishID, err = getResetEventIDByType(ctx, c, resetType, namespace, wid, rid, frontendClient)
		if err != nil {
			ErrorAndExit("getResetEventIDByType failed", err)
		}
	}
	resp, err := frontendClient.ResetWorkflowExecution(ctx, &workflowservice.ResetWorkflowExecutionRequest{
		Namespace: namespace,
		WorkflowExecution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      resetBaseRunID,
		},
		Reason:                    fmt.Sprintf("%v:%v", getCurrentUserFromEnv(), reason),
		WorkflowTaskFinishEventId: workflowTaskFinishID,
		RequestId:                 uuid.New(),
	})
	if err != nil {
		ErrorAndExit("reset failed", err)
	}
	prettyPrintJSONObject(resp)
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
func ResetInBatch(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	resetType := getRequiredOption(c, FlagResetType)

	inFileName := c.String(FlagInputFile)
	query := c.String(FlagListQuery)
	excFileName := c.String(FlagExcludeFile)
	separator := c.String(FlagInputSeparator)
	parallel := c.Int(FlagParallism)

	extraForResetType, ok := resetTypesMap[resetType]
	if !ok {
		ErrorAndExit("Not supported reset type", nil)
	} else if len(extraForResetType) > 0 {
		getRequiredOption(c, extraForResetType)
	}

	batchResetParams := batchResetParamsType{
		reason:               getRequiredOption(c, FlagReason),
		skipOpen:             c.Bool(FlagSkipCurrentOpen),
		nonDeterministicOnly: c.Bool(FlagNonDeterministicOnly),
		skipBaseNotCurrent:   c.Bool(FlagSkipBaseIsNotCurrent),
		dryRun:               c.Bool(FlagDryRun),
		resetType:            resetType,
	}

	if inFileName == "" && query == "" {
		ErrorAndExit("Must provide input file or list query to get target workflows to reset", nil)
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
			ErrorAndExit("Open failed2", err)
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
				ErrorAndExit("Split failed", fmt.Errorf("line %v has less than 1 cols separated by comma, only %v ", idx, len(cols)))
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
			ErrorAndExit("Open failed", err)
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
				ErrorAndExit("Split failed", fmt.Errorf("line %v has less than 1 cols separated by comma, only %v ", idx, len(cols)))
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
		sdkClient := getSDKClient(c)
		pageSize := 1000
		var nextPageToken []byte
		var result []*workflowpb.WorkflowExecutionInfo
		for {
			result, nextPageToken = scanWorkflowExecutions(sdkClient, pageSize, nextPageToken, query, c)
			for _, we := range result {
				wid := we.Execution.GetWorkflowId()
				rid := we.Execution.GetRunId()
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

			if nextPageToken == nil {
				break
			}
		}
	}

	close(done)
	fmt.Println("wait for all goroutines...")
	wg.Wait()
}

func printErrorAndReturn(msg string, err error) error {
	fmt.Println(msg)
	return err
}

func doReset(c *cli.Context, namespace, wid, rid string, params batchResetParamsType) error {
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.FrontendClient(c)
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
			Reason:                    fmt.Sprintf("%v:%v", getCurrentUserFromEnv(), params.reason),
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

		if attr.GetCause() == enumspb.WORKFLOW_TASK_FAILED_CAUSE_WORKFLOW_WORKER_UNHANDLED_FAILURE ||
			strings.Contains(attr.GetFailure().GetMessage(), "nondeterministic") {
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
	case "BadBinary":
		binCheckSum := c.String(FlagResetBadBinaryChecksum)
		resetBaseRunID, workflowTaskFinishID, err = getBadWorkflowTaskCompletedID(ctx, namespace, wid, rid, binCheckSum, frontendClient)
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
			} else if e.GetEventType() == enumspb.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED {
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

func getBadWorkflowTaskCompletedID(ctx context.Context, namespace, wid, rid, binChecksum string, frontendClient workflowservice.WorkflowServiceClient) (resetBaseRunID string, workflowTaskCompletedID int64, err error) {
	resetBaseRunID = rid
	resp, err := frontendClient.DescribeWorkflowExecution(ctx, &workflowservice.DescribeWorkflowExecutionRequest{
		Namespace: namespace,
		Execution: &commonpb.WorkflowExecution{
			WorkflowId: wid,
			RunId:      rid,
		},
	})
	if err != nil {
		return "", 0, printErrorAndReturn("DescribeWorkflowExecution failed", err)
	}

	_, p := history.FindAutoResetPoint(clock.NewRealTimeSource(), &namespacepb.BadBinaries{
		Binaries: map[string]*namespacepb.BadBinaryInfo{
			binChecksum: {},
		},
	}, resp.WorkflowExecutionInfo.AutoResetPoints)
	if p != nil {
		workflowTaskCompletedID = p.GetFirstWorkflowTaskCompletedId()
	}

	if workflowTaskCompletedID == 0 {
		return "", 0, printErrorAndReturn("Get BadWorkflowTaskCompletedID failed", serviceerror.NewInvalidArgument("no WorkflowTaskCompletedID"))
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

// CompleteActivity completes an activity
func CompleteActivity(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := getRequiredOption(c, FlagRunID)
	activityID := getRequiredOption(c, FlagActivityID)
	if len(activityID) == 0 {
		ErrorAndExit("Invalid activityId", fmt.Errorf("activityId cannot be empty"))
	}
	result := getRequiredOption(c, FlagResult)
	identity := getRequiredOption(c, FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.FrontendClient(c)
	_, err := frontendClient.RespondActivityTaskCompletedById(ctx, &workflowservice.RespondActivityTaskCompletedByIdRequest{
		Namespace:  namespace,
		WorkflowId: wid,
		RunId:      rid,
		ActivityId: activityID,
		Result:     payloads.EncodeString(result),
		Identity:   identity,
	})
	if err != nil {
		ErrorAndExit("Completing activity failed", err)
	} else {
		fmt.Println("Complete activity successfully.")
	}
}

// FailActivity fails an activity
func FailActivity(c *cli.Context) {
	namespace := getRequiredGlobalOption(c, FlagNamespace)
	wid := getRequiredOption(c, FlagWorkflowID)
	rid := getRequiredOption(c, FlagRunID)
	activityID := getRequiredOption(c, FlagActivityID)
	if len(activityID) == 0 {
		ErrorAndExit("Invalid activityId", fmt.Errorf("activityId cannot be empty"))
	}
	reason := getRequiredOption(c, FlagReason)
	detail := getRequiredOption(c, FlagDetail)
	identity := getRequiredOption(c, FlagIdentity)
	ctx, cancel := newContext(c)
	defer cancel()

	frontendClient := cFactory.FrontendClient(c)
	_, err := frontendClient.RespondActivityTaskFailedById(ctx, &workflowservice.RespondActivityTaskFailedByIdRequest{
		Namespace:  namespace,
		WorkflowId: wid,
		RunId:      rid,
		ActivityId: activityID,
		Failure: &failurepb.Failure{
			Message: reason,
			Source:  "CLI",
			FailureInfo: &failurepb.Failure_ApplicationFailureInfo{ApplicationFailureInfo: &failurepb.ApplicationFailureInfo{
				NonRetryable: true,
				Details:      payloads.EncodeString(detail),
			}},
		},
		Identity: identity,
	})
	if err != nil {
		ErrorAndExit("Failing activity failed", err)
	} else {
		fmt.Println("Fail activity successfully.")
	}
}

func getWorkflowParams(c *cli.Context) (string, string) {
	var wid, rid string

	if c.NArg() >= 1 {
		wid = c.Args().First()
		if c.NArg() >= 2 {
			rid = c.Args().Get(1)
		}
	} else {
		wid = getRequiredOption(c, FlagWorkflowID)
		rid = c.String(FlagRunID)
	}

	return wid, rid
}

func listWorkflows(ctx context.Context, client workflowservice.WorkflowServiceClient, npt []byte, namespace string, query string) ([]interface{}, []byte, error) {
	req := &workflowservice.ListWorkflowExecutionsRequest{
		Namespace: namespace,
		Query:     query,
	}
	resp, err := client.ListWorkflowExecutions(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	var items []interface{}
	for _, e := range resp.Executions {
		items = append(items, e)
	}

	return items, resp.NextPageToken, nil
}

func listOpenWorkflows(ctx context.Context, client workflowservice.WorkflowServiceClient, npt []byte, namespace string, earliestTime, latestTime time.Time, wfID, wfType string) ([]interface{}, []byte, error) {
	req := &workflowservice.ListOpenWorkflowExecutionsRequest{
		Namespace: namespace,
		StartTimeFilter: &filterpb.StartTimeFilter{
			EarliestTime: &earliestTime,
			LatestTime:   &latestTime,
		},
	}
	if len(wfID) > 0 {
		req.Filters = &workflowservice.ListOpenWorkflowExecutionsRequest_ExecutionFilter{ExecutionFilter: &filterpb.WorkflowExecutionFilter{WorkflowId: wfID}}
	}
	if len(wfType) > 0 {
		req.Filters = &workflowservice.ListOpenWorkflowExecutionsRequest_TypeFilter{TypeFilter: &filterpb.WorkflowTypeFilter{Name: wfType}}
	}
	resp, err := client.ListOpenWorkflowExecutions(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	var items []interface{}
	for _, e := range resp.Executions {
		items = append(items, e)
	}

	return items, resp.NextPageToken, nil
}

func listClosedWorkflows(ctx context.Context, client workflowservice.WorkflowServiceClient, npt []byte, namespace string, earliestTime, latestTime time.Time, wfID, wfType string,
	wfStatus enumspb.WorkflowExecutionStatus) ([]interface{}, []byte, error) {
	req := &workflowservice.ListClosedWorkflowExecutionsRequest{
		Namespace: namespace,
		StartTimeFilter: &filterpb.StartTimeFilter{
			EarliestTime: &earliestTime,
			LatestTime:   &latestTime,
		},
	}
	if len(wfID) > 0 {
		req.Filters = &workflowservice.ListClosedWorkflowExecutionsRequest_ExecutionFilter{ExecutionFilter: &filterpb.WorkflowExecutionFilter{WorkflowId: wfID}}
	}
	if len(wfType) > 0 {
		req.Filters = &workflowservice.ListClosedWorkflowExecutionsRequest_TypeFilter{TypeFilter: &filterpb.WorkflowTypeFilter{Name: wfType}}
	}
	if wfStatus != enumspb.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED {
		req.Filters = &workflowservice.ListClosedWorkflowExecutionsRequest_StatusFilter{StatusFilter: &filterpb.StatusFilter{Status: wfStatus}}
	}
	resp, err := client.ListClosedWorkflowExecutions(ctx, req)
	if err != nil {
		return nil, nil, err
	}
	var items []interface{}
	for _, e := range resp.Executions {
		items = append(items, e)
	}

	return items, resp.NextPageToken, nil
}

type eventRow struct {
	ID      string
	Time    string
	Type    string
	Details string
}
