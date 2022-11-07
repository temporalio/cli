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

package app_test

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/temporal-cli/app"
	"github.com/urfave/cli/v2"
	commonpb "go.temporal.io/api/common/v1"
	"go.temporal.io/api/enums/v1"
	enumspb "go.temporal.io/api/enums/v1"
	historypb "go.temporal.io/api/history/v1"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/operatorservicemock/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/api/workflowservicemock/v1"
	"go.temporal.io/sdk/client"
	sdkclient "go.temporal.io/sdk/client"
	sdkmocks "go.temporal.io/sdk/mocks"
	"go.temporal.io/server/common/primitives/timestamp"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type cliAppSuite struct {
	suite.Suite
	app            *cli.App
	mockCtrl       *gomock.Controller
	frontendClient *workflowservicemock.MockWorkflowServiceClient
	operatorClient *operatorservicemock.MockOperatorServiceClient
	sdkClient      *sdkmocks.Client
}

type clientFactoryMock struct {
	frontendClient workflowservice.WorkflowServiceClient
	operatorClient operatorservice.OperatorServiceClient
	sdkClient      *sdkmocks.Client
}

func (m *clientFactoryMock) FrontendClient(c *cli.Context) workflowservice.WorkflowServiceClient {
	return m.frontendClient
}

func (m *clientFactoryMock) OperatorClient(c *cli.Context) operatorservice.OperatorServiceClient {
	return m.operatorClient
}

func (m *clientFactoryMock) SDKClient(c *cli.Context, namespace string) sdkclient.Client {
	return m.sdkClient
}

func (m *clientFactoryMock) HealthClient(_ *cli.Context) healthpb.HealthClient {
	panic("HealthClient mock is not supported")
}

var commands = []string{
	"activity",
	"workflow",
	"task-queue",
	"operator",
	"env",
}

var cliTestNamespace = "cli-test-namespace"

func TestCLIAppSuite(t *testing.T) {
	s := new(cliAppSuite)
	suite.Run(t, s)
}

func (s *cliAppSuite) SetupSuite() {
	s.app = app.BuildApp("")
}

func (s *cliAppSuite) SetupTest() {
	s.mockCtrl = gomock.NewController(s.T())

	s.frontendClient = workflowservicemock.NewMockWorkflowServiceClient(s.mockCtrl)
	s.operatorClient = operatorservicemock.NewMockOperatorServiceClient(s.mockCtrl)
	s.sdkClient = &sdkmocks.Client{}
	app.SetFactory(&clientFactoryMock{
		frontendClient: s.frontendClient,
		operatorClient: s.operatorClient,
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

var (
	eventType = enumspb.EVENT_TYPE_WORKFLOW_EXECUTION_STARTED
)

var describeTaskQueueResponse = &workflowservice.DescribeTaskQueueResponse{
	Pollers: []*taskqueuepb.PollerInfo{
		{
			LastAccessTime: timestamp.TimePtr(time.Now().UTC()),
			Identity:       "tester",
		},
	},
}

// // TestAcceptStringSliceArgsWithCommas tests that the cli accepts string slice args with commas
// // If the test fails consider downgrading urfave/cli/v2 to v2.4.0
// // See https://github.com/urfave/cli/pull/1241
// func (s *cliAppSuite) TestAcceptStringSliceArgsWithCommas() {
// 	app := cli.NewApp()
// 	app.Name = "testapp"
// 	app.Commands = []*cli.Command{
// 		{
// 			Name: "dostuff",
// 			Action: func(c *cli.Context) error {
// 				s.Equal(2, len(c.StringSlice("input")))
// 				for _, inp := range c.StringSlice("input") {
// 					var thing any
// 					s.NoError(json.Unmarshal([]byte(inp), &thing))
// 				}
// 				return nil
// 			},
// 			Flags: []cli.Flag{
// 				&cli.StringSliceFlag{
// 					Name: "input",
// 				},
// 			},
// 		},
// 	}
// 	app.Run([]string{"testapp", "dostuff",
// 		"--input", `{"field1": 34, "field2": false}`,
// 		"--input", `{"numbers": [4,5,6]}`})
// }

func (s *cliAppSuite) TestDescribeTaskQueue() {
	s.sdkClient.On("DescribeTaskQueue", mock.Anything, mock.Anything, mock.Anything).Return(describeTaskQueueResponse, nil).Once()
	err := s.app.Run([]string{"", "task-queue", "describe", "--task-queue", "test-taskQueue", "--namespace", cliTestNamespace})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestDescribeTaskQueue_Activity() {
	s.sdkClient.On("DescribeTaskQueue", mock.Anything, mock.Anything, mock.Anything).Return(describeTaskQueueResponse, nil).Once()
	err := s.app.Run([]string{"", "task-queue", "describe", "--namespace", cliTestNamespace, "--task-queue", "test-taskQueue", "--task-queue-type", "activity"})
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

	iteratorMock.On("HasNext").Return(hasNextFn)
	iteratorMock.On("Next").Return(nextFn, nil).Once()

	return iteratorMock
}

func workflowRun() sdkclient.WorkflowRun {
	workflowRunMock := &sdkmocks.WorkflowRun{}

	workflowRunMock.On("GetRunID").Return(uuid.New()).Maybe()
	workflowRunMock.On("GetID").Return(uuid.New()).Maybe()

	return workflowRunMock
}

func (s *cliAppSuite) RunWithExitCode(arguments []string) int {
	origExiter := cli.OsExiter
	defer func() { cli.OsExiter = origExiter }()

	var exitCode int
	cli.OsExiter = func(code int) {
		exitCode = code
	}

	s.app.Run(arguments)
	return exitCode
}

func newServerAndClientOpts(port int, customArgs ...string) ([]string, client.Options) {
	args := []string{
		"temporal",
		"server",
		"start-dev",
		"--namespace", "default",
		// Use noop logger to avoid fatal logs failing tests on shutdown signal.
		"--log-format", "noop",
		"--headless",
		"--port", strconv.Itoa(port),
	}

	return append(args, customArgs...), client.Options{
		HostPort:  fmt.Sprintf("localhost:%d", port),
		Namespace: "temporal-system",
	}
}

func assertServerHealth(t *testing.T, ctx context.Context, opts client.Options) {
	var (
		c         client.Client
		clientErr error
	)
	for i := 0; i < 50; i++ {
		if c, clientErr = client.Dial(opts); clientErr == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if clientErr != nil {
		t.Error(clientErr)
	}

	if _, err := c.CheckHealth(ctx, nil); err != nil {
		t.Error(err)
	}

	// Check for pollers on a system task queue to ensure that the worker service is running.
	for {
		if ctx.Err() != nil {
			t.Error(ctx.Err())
			break
		}
		resp, err := c.DescribeTaskQueue(ctx, "temporal-sys-tq-scanner-taskqueue-0", enums.TASK_QUEUE_TYPE_WORKFLOW)
		if err != nil {
			t.Error(err)
		}
		if len(resp.GetPollers()) > 0 {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
}

// func TestCreateDataDirectory(t *testing.T) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// 	defer cancel()

// 	testUserHome := filepath.Join(os.TempDir(), "temporal_test", t.Name())
// 	t.Cleanup(func() {
// 		if err := os.RemoveAll(testUserHome); err != nil {
// 			fmt.Println("error cleaning up temp dir:", err)
// 		}
// 	})
// 	// Set user home for all supported operating systems
// 	t.Setenv("AppData", testUserHome)         // Windows
// 	t.Setenv("HOME", testUserHome)            // macOS
// 	t.Setenv("XDG_CONFIG_HOME", testUserHome) // linux
// 	// Verify that worked
// 	configDir, _ := os.UserConfigDir()
// 	if !strings.HasPrefix(configDir, testUserHome) {
// 		t.Fatalf("expected config dir %q to be inside user home directory %q", configDir, testUserHome)
// 	}

// 	temporalCLI := app.BuildApp("")
// 	// Don't call os.Exit
// 	temporalCLI.ExitErrHandler = func(_ *cli.Context, _ error) {}

// 	portProvider := sconfig.NewPortProvider()
// 	var (
// 		port1 = portProvider.MustGetFreePort()
// 		port2 = portProvider.MustGetFreePort()
// 		port3 = portProvider.MustGetFreePort()
// 	)
// 	portProvider.Close()

// 	t.Run("default db path", func(t *testing.T) {
// 		ctx, cancel := context.WithCancel(ctx)
// 		defer cancel()

// 		args, clientOpts := newServerAndClientOpts(port1)

// 		go func() {
// 			if err := temporalCLI.RunContext(ctx, args); err != nil {
// 				fmt.Println("Server closed with error:", err)
// 			}
// 		}()

// 		assertServerHealth(t, ctx, clientOpts)

// 		// If the rest of this test case passes but this assertion fails,
// 		// there may have been a breaking change in the liteconfig package
// 		// related to how the default db file path is calculated.
// 		if _, err := os.Stat(filepath.Join(configDir, "temporal", "db", "default.db")); err != nil {
// 			t.Errorf("error checking for default db file: %s", err)
// 		}
// 	})

// 	t.Run("custom db path -- missing directory", func(t *testing.T) {
// 		customDBPath := filepath.Join(testUserHome, "foo", "bar", "baz.db")
// 		args, _ := newServerAndClientOpts(
// 			port2, "-f", customDBPath,
// 		)
// 		if err := temporalCLI.RunContext(ctx, args); err != nil {
// 			if !errors.Is(err, os.ErrNotExist) {
// 				t.Errorf("expected error %q, got %q", os.ErrNotExist, err)
// 			}
// 			if !strings.Contains(err.Error(), filepath.Dir(customDBPath)) {
// 				t.Errorf("expected error %q to contain string %q", err, filepath.Dir(customDBPath))
// 			}
// 		} else {
// 			t.Error("no error when directory missing")
// 		}
// 	})

// 	t.Run("custom db path -- existing directory", func(t *testing.T) {
// 		ctx, cancel := context.WithCancel(ctx)
// 		defer cancel()

// 		args, clientOpts := newServerAndClientOpts(
// 			port3, "-f", filepath.Join(testUserHome, "foo.db"),
// 		)

// 		go func() {
// 			if err := temporalCLI.RunContext(ctx, args); err != nil {
// 				fmt.Println("Server closed with error:", err)
// 			}
// 		}()

// 		assertServerHealth(t, ctx, clientOpts)
// 	})
// }
