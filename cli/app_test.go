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
	"encoding/json"
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
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/operatorservicemock/v1"
	taskqueuepb "go.temporal.io/api/taskqueue/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/api/workflowservicemock/v1"
	sdkclient "go.temporal.io/sdk/client"
	sdkmocks "go.temporal.io/sdk/mocks"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"go.temporal.io/server/common/primitives/timestamp"
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
	"namespace",
	"workflow",
	"task-queue",
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
	s.operatorClient = operatorservicemock.NewMockOperatorServiceClient(s.mockCtrl)
	s.sdkClient = &sdkmocks.Client{}
	SetFactory(&clientFactoryMock{
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

// TestAcceptStringSliceArgsWithCommas tests that the cli accepts string slice args with commas
// If the test fails consider downgrading urfave/cli/v2 to v2.4.0
// See https://github.com/urfave/cli/pull/1241
func (s *cliAppSuite) TestAcceptStringSliceArgsWithCommas() {
	app := cli.NewApp()
	app.Name = "testapp"
	app.Commands = []*cli.Command{
		{
			Name: "dostuff",
			Action: func(c *cli.Context) error {
				s.Equal(2, len(c.StringSlice("input")))
				for _, inp := range c.StringSlice("input") {
					var thing any
					s.NoError(json.Unmarshal([]byte(inp), &thing))
				}
				return nil
			},
			Flags: []cli.Flag{
				&cli.StringSliceFlag{
					Name: "input",
				},
			},
		},
	}
	app.Run([]string{"testapp", "dostuff",
		"--input", `{"field1": 34, "field2": false}`,
		"--input", `{"numbers": [4,5,6]}`})
}

func (s *cliAppSuite) TestDescribeTaskQueue() {
	s.sdkClient.On("DescribeTaskQueue", mock.Anything, mock.Anything, mock.Anything).Return(describeTaskQueueResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "task-queue", "describe", "--task-queue", "test-taskQueue"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

func (s *cliAppSuite) TestDescribeTaskQueue_Activity() {
	s.sdkClient.On("DescribeTaskQueue", mock.Anything, mock.Anything, mock.Anything).Return(describeTaskQueueResponse, nil).Once()
	err := s.app.Run([]string{"", "--namespace", cliTestNamespace, "task-queue", "describe", "--task-queue", "test-taskQueue", "--task-queue-type", "activity"})
	s.Nil(err)
	s.sdkClient.AssertExpectations(s.T())
}

// TestParseTime tests the parsing of date argument in UTC and UnixNano formats
func (s *cliAppSuite) TestParseTime() {
	t, err := parseTime("", time.Date(1978, 8, 22, 0, 0, 0, 0, time.UTC), time.Now().UTC())
	s.NoError(err)
	s.Equal("1978-08-22 00:00:00 +0000 UTC", t.String())

	t, err = parseTime("2018-06-07T15:04:05+07:00", time.Time{}, time.Now())
	s.NoError(err)
	s.Equal("2018-06-07T15:04:05+07:00", t.Format(time.RFC3339))

	expected, err := time.Parse(defaultDateTimeFormat, "2018-06-07T15:04:05+07:00")
	s.NoError(err)

	t, err = parseTime("1528358645000000000", time.Time{}, time.Now().UTC())
	s.NoError(err)
	s.Equal(expected.UTC(), t)
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
		parsedTime, err := parseTime(te.timeStr, te.defVal, now)
		s.NoError(err)

		s.True(te.expected.Before(parsedTime) || te.expected == parsedTime, "Case: %s. %d must be less or equal than parsed %d", te.timeStr, te.expected, parsedTime)
		s.True(te.expected.Add(delta).After(parsedTime) || te.expected.Add(delta) == parsedTime, "Case: %s. %d must be greater or equal than parsed %d", te.timeStr, te.expected, parsedTime)
	}
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
