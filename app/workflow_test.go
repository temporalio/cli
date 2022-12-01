package app_test

import (
	"context"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/temporalio/cli/workflow"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/api/serviceerror"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/server/common/payloads"
	"go.temporal.io/server/common/primitives/timestamp"
)

func (s *cliAppSuite) TestShowHistory() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "workflow", "show", "--namespace", cliTestNamespace, "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestShowHistoryWithFollow() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "workflow", "show", "--namespace", cliTestNamespace, "--workflow-id", "wid", "--follow"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err = s.app.Run([]string{"", "workflow", "show", "--namespace", cliTestNamespace, "--workflow-id", "wid", "--fields", "long", "--follow"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestStartWorkflow() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)

	// start with wid
	err := s.app.Run([]string{"", "workflow", "start", "--namespace", cliTestNamespace, "--task-queue", "testTaskQueue", "--type", "testWorkflowType", "--execution-timeout", "60", "--run-timeout", "60", "--workflow-id", "wid", "--id-reuse-policy", "Unspecified"})
	s.Nil(err)
	s.sdkClient.AssertNotCalled(s.T(), "GetWorkflowHistory")
	s.sdkClient.AssertExpectations(s.T())

	hasNoSearchAttributes := mock.MatchedBy(func(options sdkclient.StartWorkflowOptions) bool {
		return len(options.SearchAttributes) == 0
	})
	s.sdkClient.AssertCalled(s.T(), "ExecuteWorkflow", mock.Anything, hasNoSearchAttributes, mock.Anything, mock.Anything)

	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	// start without wid
	err = s.app.Run([]string{"", "workflow", "start", "--namespace", cliTestNamespace, "--task-queue", "testTaskQueue", "--type", "testWorkflowType", "--execution-timeout", "60", "--run-timeout", "60", "--id-reuse-policy", "Unspecified"})
	s.Nil(err)
	s.sdkClient.AssertNotCalled(s.T(), "GetWorkflowHistory")
	s.sdkClient.AssertExpectations(s.T())
	s.sdkClient.AssertCalled(s.T(), "ExecuteWorkflow", mock.Anything, hasNoSearchAttributes, mock.Anything, mock.Anything)
}

func (s *cliAppSuite) TestStartWorkflow_SearchAttributes() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	// start with basic search attributes
	err := s.app.Run([]string{"", "workflow", "start", "--namespace", cliTestNamespace, "--task-queue", "testTaskQueue", "--type", "testWorkflowType",
		"--search-attribute", "k1=\"v1\"", "--search-attribute", "k2=\"v2\""})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	hasCorrectSearchAttributes := mock.MatchedBy(func(options sdkclient.StartWorkflowOptions) bool {
		return len(options.SearchAttributes) == 2 &&
			options.SearchAttributes["k1"] == "v1" &&
			options.SearchAttributes["k2"] == "v2"
	})
	s.sdkClient.AssertCalled(s.T(), "ExecuteWorkflow", mock.Anything, hasCorrectSearchAttributes, mock.Anything, mock.Anything)

	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
}

func (s *cliAppSuite) TestStartWorkflow_Memo() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	err := s.app.Run([]string{"", "workflow", "start", "--namespace", cliTestNamespace, "--task-queue", "testTaskQueue", "--type", "testWorkflowType",
		"--memo", "k1=\"v1\"", "--memo", "k2=\"v2\""})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	hasCorrectMemo := mock.MatchedBy(func(options sdkclient.StartWorkflowOptions) bool {
		return len(options.Memo) == 2 &&
			options.Memo["k1"] == "v1" &&
			options.Memo["k2"] == "v2"
	})
	s.sdkClient.AssertCalled(s.T(), "ExecuteWorkflow", mock.Anything, hasCorrectMemo, mock.Anything, mock.Anything)

	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
}

func (s *cliAppSuite) TestStartWorkflow_Failed() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), serviceerror.NewInvalidArgument("fake error"))

	// start with wid
	errorCode := s.RunWithExitCode([]string{"", "workflow", "start", "--namespace", cliTestNamespace, "--task-queue", "testTaskQueue", "--type", "testWorkflowType", "--execution-timeout", "60", "--run-timeout", "60", "--workflow-id", "wid"})
	s.Equal(1, errorCode)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestExecuteWorkflow() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(historyEventIterator()).Once()

	err := s.app.Run([]string{"", "workflow", "execute", "--namespace", cliTestNamespace, "--task-queue", "testTaskQueue", "--type", "testWorkflowType"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestTerminateWorkflow() {
	s.sdkClient.On("TerminateWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	err := s.app.Run([]string{"", "workflow", "terminate", "--namespace", cliTestNamespace, "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestTerminateWorkflow_Failed() {
	s.sdkClient.On("TerminateWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(serviceerror.NewInvalidArgument("faked error")).Once()
	errorCode := s.RunWithExitCode([]string{"", "workflow", "terminate", "--namespace", cliTestNamespace, "--workflow-id", "wid"})
	s.Equal(1, errorCode)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCancelWorkflow() {
	s.sdkClient.On("CancelWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err := s.app.Run([]string{"", "workflow", "cancel", "--namespace", cliTestNamespace, "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCancelWorkflow_Failed() {
	s.sdkClient.On("CancelWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(serviceerror.NewInvalidArgument("faked error")).Once()
	errorCode := s.RunWithExitCode([]string{"", "workflow", "cancel", "--namespace", cliTestNamespace, "--workflow-id", "wid"})
	s.Equal(1, errorCode)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestSignalWorkflow() {
	s.frontendClient.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "workflow", "signal", "--namespace", cliTestNamespace, "--name", "signal-name", "--workflow-id", "wid"})
	s.Nil(err)
}

func (s *cliAppSuite) TestSignalWorkflow_Failed() {
	s.frontendClient.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.RunWithExitCode([]string{"", "workflow", "signal", "--namespace", cliTestNamespace, "--name", "signal-name", "--workflow-id", "wid"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestQueryWorkflow() {
	resp := &workflowservice.QueryWorkflowResponse{
		QueryResult: payloads.EncodeString("query-result"),
	}
	s.frontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, nil)
	err := s.app.Run([]string{"", "workflow", "query", "--namespace", cliTestNamespace, "--workflow-id", "wid", "--type", "query-type-test"})
	s.Nil(err)
}

func (s *cliAppSuite) TestQueryWorkflowUsingStackTrace() {
	resp := &workflowservice.QueryWorkflowResponse{
		QueryResult: payloads.EncodeString("query-result"),
	}
	s.frontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, nil)
	err := s.app.Run([]string{"", "workflow", "stack", "--namespace", cliTestNamespace, "--workflow-id", "wid"})
	s.Nil(err)
}

func (s *cliAppSuite) TestQueryWorkflow_Failed() {
	resp := &workflowservice.QueryWorkflowResponse{
		QueryResult: payloads.EncodeString("query-result"),
	}
	s.frontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.RunWithExitCode([]string{"", "workflow", "query", "--namespace", cliTestNamespace, "--workflow-id", "wid", "--type", "query-type-test"})
	s.Equal(1, errorCode)
}

var (
	listWorkflowExecutionsResponse = &workflowservice.ListWorkflowExecutionsResponse{
		Executions: []*workflowpb.WorkflowExecutionInfo{
			{
				Execution: &commonpb.WorkflowExecution{
					WorkflowId: "test-list-open-workflow-id",
					RunId:      uuid.New(),
				},
				Type: &commonpb.WorkflowType{
					Name: "test-list-open-workflow-type",
				},
				StartTime:     timestamp.TimePtr(time.Now().UTC()),
				CloseTime:     timestamp.TimePtr(time.Now().UTC().Add(time.Hour)),
				HistoryLength: 12,
			},
			{
				Execution: &commonpb.WorkflowExecution{
					WorkflowId: "test-list-workflow-id",
					RunId:      uuid.New(),
				},
				Type: &commonpb.WorkflowType{
					Name: "test-list-workflow-type",
				},
				StartTime:     timestamp.TimePtr(time.Now().UTC()),
				CloseTime:     timestamp.TimePtr(time.Now().UTC().Add(time.Hour)),
				Status:        enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				HistoryLength: 12,
			},
		},
	}
)

func (s *cliAppSuite) TestListWorkflow() {
	s.sdkClient.On("ListWorkflow", mock.Anything, mock.Anything).Return(listWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "workflow", "list", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_DeadlineExceeded() {
	s.sdkClient.On("ListWorkflow", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded).Once()
	s.sdkClient.On("ListWorkflow", mock.Anything, mock.Anything).Return(listWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "workflow", "list", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_Open_WithQuery() {
	s.sdkClient.On("ListWorkflow", mock.Anything, mock.Anything).Return(listWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "workflow", "list", "--query", "ExecutionStatus='Running'", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListArchivedWorkflow() {
	s.sdkClient.On("ListArchivedWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.ListArchivedWorkflowExecutionsResponse{}, nil).Once()
	err := s.app.Run([]string{"", "workflow", "list", "--archived", "--query", "some query string", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCountWorkflow() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{}, nil).Once()
	err := s.app.Run([]string{"", "workflow", "count", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{}, nil).Once()
	err = s.app.Run([]string{"", "workflow", "count", "--query", "'CloseTime = missing'", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCountWorkflowDeadlineExceeded() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded).Once()
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{}, nil).Once()
	err := s.app.Run([]string{"", "workflow", "count", "--query", "'CloseTime = missing'", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestParseFoldStatusList() {
	tests := map[string]struct {
		value   string
		want    []enumspb.WorkflowExecutionStatus
		wantErr bool
	}{
		"default values": {
			value: "completed,canceled,terminated",
			want: []enumspb.WorkflowExecutionStatus{
				enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
				enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
			},
		},
		"no values": {
			value: "",
			want:  nil,
		},
		"invalid": {
			value:   "Foobar",
			wantErr: true,
		},
		"title case": {
			value: "Running,Completed,Failed,Canceled,Terminated,ContinuedAsNew,TimedOut",
			want: []enumspb.WorkflowExecutionStatus{
				enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
				enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				enumspb.WORKFLOW_EXECUTION_STATUS_FAILED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
				enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
				enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
			},
		},
		"upper case": {
			value: "RUNNING,COMPLETED,FAILED,CANCELED,TERMINATED,CONTINUEDASNEW,TIMEDOUT",
			want: []enumspb.WorkflowExecutionStatus{
				enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING,
				enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED,
				enumspb.WORKFLOW_EXECUTION_STATUS_FAILED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CANCELED,
				enumspb.WORKFLOW_EXECUTION_STATUS_TERMINATED,
				enumspb.WORKFLOW_EXECUTION_STATUS_CONTINUED_AS_NEW,
				enumspb.WORKFLOW_EXECUTION_STATUS_TIMED_OUT,
			},
		},
	}
	for name, tt := range tests {
		s.Run(name, func() {
			got, err := workflow.ParseFoldStatusList(tt.value)
			if tt.wantErr {
				s.Error(err)
			} else {
				s.Equal(tt.want, got)
			}
		})
	}
}
