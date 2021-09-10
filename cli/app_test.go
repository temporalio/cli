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
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/urfave/cli/v2"
	commonpb "go.temporal.io/api/common/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	namespacepb "go.temporal.io/api/namespace/v1"
	replicationpb "go.temporal.io/api/replication/v1"
	"go.temporal.io/api/serviceerror"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/api/workflowservicemock/v1"
	sdkclient "go.temporal.io/sdk/client"
	sdkmocks "go.temporal.io/sdk/mocks"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"go.temporal.io/server/common/payloads"
	"go.temporal.io/server/common/primitives/timestamp"
)

type cliAppSuite struct {
	suite.Suite
	app            *cli.App
	mockCtrl       *gomock.Controller
	frontendClient *workflowservicemock.MockWorkflowServiceClient
	sdkClient      *sdkmocks.Client
}

type clientFactoryMock struct {
	frontendClient workflowservice.WorkflowServiceClient
	sdkClient      *sdkmocks.Client
}

func (m *clientFactoryMock) FrontendClient(c *cli.Context) workflowservice.WorkflowServiceClient {
	return m.frontendClient
}

func (m *clientFactoryMock) SDKClient(c *cli.Context, namespace string) sdkclient.Client {
	return m.sdkClient
}

func (m *clientFactoryMock) HealthClient(_ *cli.Context) healthpb.HealthClient {
	panic("HealthClient mock is not supported.")
}

var commands = []string{
	"namespace", "n",
	"workflow", "w",
	"taskqueue", "tq",
}

var cliTestNamespace = "cli-test-namespace"

func TestCLIAppSuite(t *testing.T) {
	s := new(cliAppSuite)
	suite.Run(t, s)
}

func (s *cliAppSuite) SetupSuite() {
	s.app = NewCliApp()
}

func (s *cliAppSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())

	s.frontendClient = workflowservicemock.NewMockWorkflowServiceClient(s.mockCtrl)
	s.sdkClient = &sdkmocks.Client{}
	SetFactory(&clientFactoryMock{
		frontendClient: s.frontendClient,
		sdkClient:      s.sdkClient,
	})
}

func (s *cliAppSuite) TearDownTest() {
	s.mockCtrl.Finish() // assert mock’s expectations
}

func (s *cliAppSuite) TestAppCommands() {
	for _, test := range commands {
		cmd := s.app.Command(test)
		s.NotNil(cmd)
	}
}

func (s *cliAppSuite) TestNamespaceRegister_LocalNamespace() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "register", "--global_namespace", "false"})
	s.Equal(0, errorCode)
}

func (s *cliAppSuite) TestNamespaceRegister_GlobalNamespace() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, nil)
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "register", "--global_namespace", "true"})
	s.Equal(0, errorCode)
}

func (s *cliAppSuite) TestNamespaceRegister_NamespaceExist() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNamespaceAlreadyExists(""))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "register", "--global_namespace", "true"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceRegister_Failed() {
	s.frontendClient.EXPECT().RegisterNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "register", "--global_namespace", "true"})
	s.Equal(1, errorCode)
}

var describeNamespaceResponseServer = &workflowservice.DescribeNamespaceResponse{
	NamespaceInfo: &namespacepb.NamespaceInfo{
		Name:        "test-namespace",
		Description: "a test namespace",
		OwnerEmail:  "test@uber.com",
	},
	Config: &namespacepb.NamespaceConfig{
		WorkflowExecutionRetentionTtl: timestamp.DurationPtr(3 * time.Hour * 24),
	},
	ReplicationConfig: &replicationpb.NamespaceReplicationConfig{
		ActiveClusterName: "active",
		Clusters: []*replicationpb.ClusterReplicationConfig{
			{
				ClusterName: "active",
			},
			{
				ClusterName: "standby",
			},
		},
	},
}

func (s *cliAppSuite) TestNamespaceUpdate() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil).Times(2)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, nil).Times(2)
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "update"})
	s.Nil(err)
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "update", "--desc", "another desc", "--oe", "another@uber.com", "--rd", "1"})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceUpdate_NamespaceNotExist() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNotFound(""))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "update"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceUpdate_ActiveClusterFlagNotSet_NamespaceNotExist() {
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewNotFound(""))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "update"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceUpdate_Failed() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, nil)
	s.frontendClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "update"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceDescribe() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), &workflowservice.DescribeNamespaceRequest{Namespace: cliTestNamespace, Id: ""}).Return(resp, nil)
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "describe"})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceDescribe_ById() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), &workflowservice.DescribeNamespaceRequest{Namespace: "", Id: "nid"}).Return(resp, nil)
	err := s.app.Run([]string{"", "namespace", "describe", "--namespace-id", "nid"})
	s.Nil(err)
}

func (s *cliAppSuite) TestNamespaceDescribe_NamespaceNotExist() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, serviceerror.NewNotFound(""))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "describe"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestNamespaceDescribe_Failed() {
	resp := describeNamespaceResponseServer
	s.frontendClient.EXPECT().DescribeNamespace(gomock.Any(), gomock.Any()).Return(resp, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "namespace", "describe"})
	s.Equal(1, errorCode)
}

var (
	eventType = enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED
)

func (s *cliAppSuite) TestShowHistory() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "show", "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestShowHistoryWithID() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "showid", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestShowHistory_PrintRawTime() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "show", "--workflow-id", "wid", "-prt"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestShowHistory_PrintDateTime() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "show", "--workflow-id", "wid", "-pdt"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestStartWorkflow() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	// start with wid
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "start", "--taskqueue", "testTaskQueue", "--workflow-type", "testWorkflowType", "-et", "60", "--workflow-id", "wid", "-wrp", "AllowDuplicateFailedOnly"})
	s.Nil(err)
	// start without wid
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "start", "--taskqueue", "testTaskQueue", "--workflow-type", "testWorkflowType", "-et", "60", "-wrp", "AllowDuplicateFailedOnly"})
	s.Nil(err)
}

func (s *cliAppSuite) TestStartWorkflow_Failed() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), serviceerror.NewInvalidArgument("faked error"))
	// start with wid
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "start", "--taskqueue", "testTaskQueue", "--workflow-type", "testWorkflowType", "-et", "60", "--workflow-id", "wid"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestRunWorkflow() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", mock.Anything, mock.Anything, mock.Anything).Return(historyEventIterator()).Once()

	// start with wid
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "run", "--taskqueue", "testTaskQueue", "--workflow-type", "testWorkflowType", "-et", "60", "--workflow-id", "wid", "wrp", "2"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), nil)
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	// start without wid
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "run", "--taskqueue", "testTaskQueue", "--workflow-type", "testWorkflowType", "-et", "60", "wrp", "2"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestRunWorkflow_Failed() {
	s.sdkClient.On("ExecuteWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(workflowRun(), serviceerror.NewInvalidArgument("fake error"))
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", mock.Anything, mock.Anything, mock.Anything).Return(historyEventIterator()).Once()

	// start with wid
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "run", "--taskqueue", "testTaskQueue", "--workflow-type", "testWorkflowType", "-et", "60", "--workflow-id", "wid"})
	s.Equal(1, errorCode)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestTerminateWorkflow() {
	s.sdkClient.On("TerminateWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "terminate", "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestTerminateWorkflow_Failed() {
	s.sdkClient.On("TerminateWorkflow", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(serviceerror.NewInvalidArgument("faked error")).Once()
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "terminate", "--workflow-id", "wid"})
	s.Equal(1, errorCode)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCancelWorkflow() {
	s.sdkClient.On("CancelWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "cancel", "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCancelWorkflow_Failed() {
	s.sdkClient.On("CancelWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(serviceerror.NewInvalidArgument("faked error")).Once()
	s.frontendClient.EXPECT().RequestCancelWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "cancel", "--workflow-id", "wid"})
	s.Equal(1, errorCode)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestSignalWorkflow() {
	s.frontendClient.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, nil)
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "signal", "--workflow-id", "wid", "-n", "signal-name"})
	s.Nil(err)
}

func (s *cliAppSuite) TestSignalWorkflow_Failed() {
	s.frontendClient.EXPECT().SignalWorkflowExecution(gomock.Any(), gomock.Any()).Return(nil, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "signal", "--workflow-id", "wid", "-n", "signal-name"})
	s.Equal(1, errorCode)
}

func (s *cliAppSuite) TestQueryWorkflow() {
	resp := &workflowservice.QueryWorkflowResponse{
		QueryResult: payloads.EncodeString("query-result"),
	}
	s.frontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, nil)
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "query", "--workflow-id", "wid", "-qt", "query-type-test"})
	s.Nil(err)
}

func (s *cliAppSuite) TestQueryWorkflowUsingStackTrace() {
	resp := &workflowservice.QueryWorkflowResponse{
		QueryResult: payloads.EncodeString("query-result"),
	}
	s.frontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, nil)
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "stack", "--workflow-id", "wid"})
	s.Nil(err)
}

func (s *cliAppSuite) TestQueryWorkflow_Failed() {
	resp := &workflowservice.QueryWorkflowResponse{
		QueryResult: payloads.EncodeString("query-result"),
	}
	s.frontendClient.EXPECT().QueryWorkflow(gomock.Any(), gomock.Any()).Return(resp, serviceerror.NewInvalidArgument("faked error"))
	errorCode := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "query", "--workflow-id", "wid", "-qt", "query-type-test"})
	s.Equal(1, errorCode)
}

var (
	status = enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED

	listClosedWorkflowExecutionsResponse = &workflowservice.ListClosedWorkflowExecutionsResponse{
		Executions: []*workflowpb.WorkflowExecutionInfo{
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
				Status:        status,
				HistoryLength: 12,
			},
		},
	}

	listOpenWorkflowExecutionsResponse = &workflowservice.ListOpenWorkflowExecutionsResponse{
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
		},
	}
)

func (s *cliAppSuite) TestListWorkflow() {
	s.sdkClient.On("ListClosedWorkflow", mock.Anything, mock.Anything).Return(listClosedWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_DeadlineExceeded() {
	s.sdkClient.On("ListClosedWorkflow", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded).Once()
	s.sdkClient.On("ListClosedWorkflow", mock.Anything, mock.Anything).Return(listClosedWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_WithWorkflowID() {
	s.sdkClient.On("ListClosedWorkflow", mock.Anything, mock.Anything).Return(listClosedWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list", "--workflow-id", "nothing"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_WithWorkflowType() {
	s.sdkClient.On("ListClosedWorkflow", mock.Anything, mock.Anything).Return(listClosedWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list", "--workflow-type", "no-type"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_Open() {
	s.sdkClient.On("ListOpenWorkflow", mock.Anything, mock.Anything).Return(listOpenWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list", "--open"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_Open_DeadlineExceeded() {
	s.sdkClient.On("ListOpenWorkflow", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded).Once()
	s.sdkClient.On("ListOpenWorkflow", mock.Anything, mock.Anything).Return(listOpenWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list", "--open"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_Open_WithWorkflowID() {
	s.sdkClient.On("ListOpenWorkflow", mock.Anything, mock.Anything).Return(listOpenWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list", "--open", "--workflow-id", "nothing"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListWorkflow_Open_WithWorkflowType() {
	s.sdkClient.On("ListOpenWorkflow", mock.Anything, mock.Anything).Return(listOpenWorkflowExecutionsResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "list", "--open", "--workflow-type", "no-type"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestListArchivedWorkflow() {
	s.sdkClient.On("ListArchivedWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.ListArchivedWorkflowExecutionsResponse{}, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "listarchived", "--query", "some query string"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCountWorkflow() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{}, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "count"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{}, nil).Once()
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "count", "-q", "'CloseTime = missing'"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestCountWorkflowDeadlineExceeded() {
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(nil, context.DeadlineExceeded).Once()
	s.sdkClient.On("CountWorkflow", mock.Anything, mock.Anything).Return(&workflowservice.CountWorkflowExecutionsResponse{}, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "count", "--query", "'CloseTime = missing'"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

var describeTaskQueueResponse = &workflowservice.DescribeTaskQueueResponse{
	Pollers: []*taskqueuepb.PollerInfo{
		{
			LastAccessTime: timestamp.TimePtr(time.Now().UTC()),
			Identity:       "tester",
		},
	},
}

func (s *cliAppSuite) TestDescribeTaskQueue() {
	s.sdkClient.On("DescribeTaskQueue", mock.Anything, mock.Anything, mock.Anything).Return(describeTaskQueueResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "taskqueue", "describe", "--taskqueue", "test-taskQueue"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestDescribeTaskQueue_Activity() {
	s.sdkClient.On("DescribeTaskQueue", mock.Anything, mock.Anything, mock.Anything).Return(describeTaskQueueResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "taskqueue", "describe", "--taskqueue", "test-taskQueue", "--taskqueuetype", "activity"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestObserveWorkflow() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "observe", "--workflow-id", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "observe", "--workflow-id", "wid", "--show-detail"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestObserveWorkflowWithID() {
	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "observe", "wid"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("GetWorkflowHistory", mock.Anything, "wid", "", mock.Anything, mock.Anything).Return(historyEventIterator()).Once()
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "workflow", "observe", "wid", "--show-detail"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

// TestParseTime tests the parsing of date argument in UTC and UnixNano formats
func (s *cliAppSuite) TestParseTime() {
	s.Equal("1978-08-22 00:00:00 +0000 UTC", parseTime("", time.Date(1978, 8, 22, 0, 0, 0, 0, time.UTC), time.Now().UTC()).String())
	s.Equal("2018-06-07T15:04:05+07:00", parseTime("2018-06-07T15:04:05+07:00", time.Time{}, time.Now()).Format(time.RFC3339))
	expected, err := time.Parse(defaultDateTimeFormat, "2018-06-07T15:04:05+07:00")
	s.NoError(err)
	s.Equal(expected.UTC(), parseTime("1528358645000000000", time.Time{}, time.Now().UTC()))
}

// TestParseTimeDateRange tests the parsing of date argument in time range format, N<duration>
// where N is the integral multiplier, and duration can be second/minute/hour/day/week/month/year
func (s *cliAppSuite) TestParseTimeDateRange() {
	now := time.Now().UTC()
	tests := []struct {
		timeStr  string    // input
		defVal   time.Time // input
		expected time.Time // expected unix nano (approx)
	}{
		{
			timeStr:  "1s",
			defVal:   time.Time{},
			expected: now.Add(-time.Second),
		},
		{
			timeStr:  "100second",
			defVal:   time.Time{},
			expected: now.Add(-100 * time.Second),
		},
		{
			timeStr:  "2m",
			defVal:   time.Time{},
			expected: now.Add(-2 * time.Minute),
		},
		{
			timeStr:  "200minute",
			defVal:   time.Time{},
			expected: now.Add(-200 * time.Minute),
		},
		{
			timeStr:  "3h",
			defVal:   time.Time{},
			expected: now.Add(-3 * time.Hour),
		},
		{
			timeStr:  "1000hour",
			defVal:   time.Time{},
			expected: now.Add(-1000 * time.Hour),
		},
		{
			timeStr:  "5d",
			defVal:   time.Time{},
			expected: now.Add(-5 * day),
		},
		{
			timeStr:  "25day",
			defVal:   time.Time{},
			expected: now.Add(-25 * day),
		},
		{
			timeStr:  "5w",
			defVal:   time.Time{},
			expected: now.Add(-5 * week),
		},
		{
			timeStr:  "52week",
			defVal:   time.Time{},
			expected: now.Add(-52 * week),
		},
		{
			timeStr:  "3M",
			defVal:   time.Time{},
			expected: now.Add(-3 * month),
		},
		{
			timeStr:  "6month",
			defVal:   time.Time{},
			expected: now.Add(-6 * month),
		},
		{
			timeStr:  "1y",
			defVal:   time.Time{},
			expected: now.Add(-year),
		},
		{
			timeStr:  "7year",
			defVal:   time.Time{},
			expected: now.Add(-7 * year),
		},
		{
			timeStr:  "100y", // epoch time will be returned as that's the minimum unix timestamp possible
			defVal:   time.Time{},
			expected: time.Unix(0, 0).UTC(),
		},
	}
	const delta = 5 * time.Millisecond
	for _, te := range tests {
		parsedTime := parseTime(te.timeStr, te.defVal, now)
		s.True(te.expected.Before(parsedTime) || te.expected == parsedTime, "Case: %s. %d must be less or equal than parsed %d", te.timeStr, te.expected, parsedTime)
		s.True(te.expected.Add(delta).After(parsedTime) || te.expected.Add(delta) == parsedTime, "Case: %s. %d must be greater or equal than parsed %d", te.timeStr, te.expected, parsedTime)
	}
}

func (s *cliAppSuite) TestGetSearchAttributes() {
	s.sdkClient.On("GetSearchAttributes", mock.Anything).Return(&workflowservice.GetSearchAttributesResponse{}, nil).Once()
	err := s.app.Run([]string{"", "cluster", "list-search-attributes"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())

	s.sdkClient.On("GetSearchAttributes", mock.Anything).Return(&workflowservice.GetSearchAttributesResponse{}, nil).Once()
	err = s.app.Run([]string{"", "--namespace", cliTestNamespace, "cluster", "list-search-attributes"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func historyEventIterator() sdkclient.HistoryEventIterator {
	iteratorMock := &sdkmocks.HistoryEventIterator{}

	counter := 0
	hasNextFn := func() bool {
		if counter == 0 {
			return true
		} else {
			return false
		}
	}

	nextFn := func() *historypb.HistoryEvent {
		if counter == 0 {
			event := &historypb.HistoryEvent{
				EventType: eventType,
				Attributes: &historypb.HistoryEvent_WorkflowExecutionStartedEventAttributes{WorkflowExecutionStartedEventAttributes: &historypb.WorkflowExecutionStartedEventAttributes{
					WorkflowType:        &commonpb.WorkflowType{Name: "TestWorkflow"},
					TaskQueue:           &taskqueuepb.TaskQueue{Name: "taskQueue"},
					WorkflowRunTimeout:  timestamp.DurationPtr(60 * time.Second),
					WorkflowTaskTimeout: timestamp.DurationPtr(10 * time.Second),
					Identity:            "tester",
				}},
			}
			counter++
			return event
		} else {
			return nil
		}
	}

	iteratorMock.On("HasNext").Return(hasNextFn).Twice()
	iteratorMock.On("Next").Return(nextFn, nil).Once()

	return iteratorMock
}

func workflowRun() sdkclient.WorkflowRun {
	workflowRunMock := &sdkmocks.WorkflowRun{}

	workflowRunMock.On("GetRunID").Return(uuid.New()).Maybe()
	workflowRunMock.On("GetID").Return(uuid.New()).Maybe()

	return workflowRunMock
}
