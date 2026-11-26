package temporalcli

import (
	"context"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/api/history/v1"
	"go.temporal.io/api/workflowservice/v1"
	"google.golang.org/grpc"
)

// reverseHistoryClient serves a fixed list of history pages, newest first.
type reverseHistoryClient struct {
	workflowservice.WorkflowServiceClient
	pages [][]*history.HistoryEvent
}

func (c *reverseHistoryClient) GetWorkflowExecutionHistoryReverse(
	_ context.Context,
	req *workflowservice.GetWorkflowExecutionHistoryReverseRequest,
	_ ...grpc.CallOption,
) (*workflowservice.GetWorkflowExecutionHistoryReverseResponse, error) {
	index := 0
	if len(req.NextPageToken) > 0 {
		var err error
		if index, err = strconv.Atoi(string(req.NextPageToken)); err != nil {
			return nil, err
		}
	}
	resp := &workflowservice.GetWorkflowExecutionHistoryReverseResponse{
		History: &history.History{Events: c.pages[index]},
	}
	if index+1 < len(c.pages) {
		resp.NextPageToken = []byte(strconv.Itoa(index + 1))
	}
	return resp, nil
}

func TestGetLastWorkflowTaskEventID_MultiplePages(t *testing.T) {
	event := func(id int64, eventType enums.EventType) *history.HistoryEvent {
		return &history.HistoryEvent{EventId: id, EventType: eventType}
	}
	// Newest events first: the last completed workflow task is event 9, and
	// an older one (event 3) sits on the next page.
	wfsvc := &reverseHistoryClient{pages: [][]*history.HistoryEvent{
		{
			event(10, enums.EVENT_TYPE_ACTIVITY_TASK_COMPLETED),
			event(9, enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED),
			event(8, enums.EVENT_TYPE_WORKFLOW_TASK_STARTED),
		},
		{
			event(7, enums.EVENT_TYPE_WORKFLOW_TASK_SCHEDULED),
			event(4, enums.EVENT_TYPE_ACTIVITY_TASK_SCHEDULED),
			event(3, enums.EVENT_TYPE_WORKFLOW_TASK_COMPLETED),
		},
	}}

	_, eventID, err := getLastWorkflowTaskEventID(context.Background(), "ns", "wid", "rid", wfsvc)
	require.NoError(t, err)
	require.Equal(t, int64(9), eventID)
}
