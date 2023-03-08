package app_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/app"
	sconfig "github.com/temporalio/cli/server/config"
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
	"batch",
	"completion",
	"env",
	"operator",
	"schedule",
	"server",
	"task-queue",
	"workflow",
}

var cliTestNamespace = "cli-test-namespace"

func TestCLIAppSuite(t *testing.T) {
	s := new(cliAppSuite)
	suite.Run(t, s)
}

func (s *cliAppSuite) SetupSuite() {
	s.app = app.BuildApp()
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

func (s *cliAppSuite) TestTopLevelCommands() {
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

// TestAcceptStringSliceArgsWithCommas tests that the cli accepts string slice flags with commas (ex. JSON)
func (s *cliAppSuite) TestAcceptStringSliceArgsWithCommas() {
	// verify that SliceFlagSeparator is disabled by default
	s.True(s.app.DisableSliceFlagSeparator)

	// verify that disabling works
	app := cli.NewApp()
	app.Name = "testapp"
	app.DisableSliceFlagSeparator = true
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

// TestFlagCategory_IsSet verifies that command flags have Category set
// As urfave/cli only prints flags in --help with Category set
func (s *cliAppSuite) TestFlagCategory_IsSet() {
	for _, cmd := range s.app.Commands {
		if cmd.Name == "server" {
			continue
		}

		verifyCategory(s, cmd)
	}
}

func verifyCategory(s *cliAppSuite, cmd *cli.Command) {
	msgT := "flag %s should have a category, command: %s %s"

	for _, flag := range cmd.Flags {
		if flag.Names()[0] == "help" {
			continue
		}

		msg := fmt.Sprintf(msgT, "--"+flag.Names()[0], cmd.Name, cmd.Usage)

		switch f := flag.(type) {
		case *cli.BoolFlag:
			s.NotEmpty(f.Category, msg)
		case *cli.IntSliceFlag:
			s.NotEmpty(f.Category, msg)
		case *cli.StringFlag:
			s.NotEmpty(f.Category, msg)
		case *cli.StringSliceFlag:
			s.NotEmpty(f.Category, msg)
		}
	}

	for _, subCmd := range cmd.Subcommands {
		verifyCategory(s, subCmd)
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

func TestCreateDataDirectory_MissingDirectory(t *testing.T) {
	temporalCLI := app.BuildApp()
	// Don't call os.Exit
	temporalCLI.ExitErrHandler = func(_ *cli.Context, _ error) {}

	portProvider := sconfig.NewPortProvider()
	port := portProvider.MustGetFreePort()
	portProvider.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	testUserHome := setupConfigOptions(t)

	customDBPath := filepath.Join(testUserHome, "foo", "bar", "baz.db")
	args, _ := newServerAndClientOpts(
		port, "-f", customDBPath,
	)
	err := temporalCLI.RunContext(ctx, args)
	if errCoder, ok := err.(cli.ExitCoder); !ok || errCoder.ExitCode() != 0 {
		t.Errorf("expected no error (error code = 0).got %q", err)
	}
}

func TestCreateDataDirectory_ExistingDirectory(t *testing.T) {
	temporalCLI := app.BuildApp()
	// Don't call os.Exit
	temporalCLI.ExitErrHandler = func(_ *cli.Context, _ error) {}

	portProvider := sconfig.NewPortProvider()
	port := portProvider.MustGetFreePort()
	portProvider.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testUserHome := setupConfigOptions(t)

	args, clientOpts := newServerAndClientOpts(
		port, "-f", filepath.Join(testUserHome, "foo.db"),
	)

	go func() {
		if err := temporalCLI.RunContext(ctx, args); err != nil {
			fmt.Println("Server closed with error:", err)
		}
	}()

	assertServerHealth(t, ctx, clientOpts)
}

func setupConfigOptions(t *testing.T) string {
	testUserHome := filepath.Join(os.TempDir(), "temporal_test", t.Name())
	t.Cleanup(func() {
		if err := os.RemoveAll(testUserHome); err != nil {
			fmt.Println("error cleaning up temp dir:", err)
		}
	})
	// Set user home for all supported operating systems
	t.Setenv("AppData", testUserHome)         // Windows
	t.Setenv("HOME", testUserHome)            // macOS
	t.Setenv("XDG_CONFIG_HOME", testUserHome) // linux
	// Verify that worked
	configDir, _ := os.UserConfigDir()
	if !strings.HasPrefix(configDir, testUserHome) {
		t.Fatalf("expected config dir %q to be inside user home directory %q", configDir, testUserHome)
	}

	return testUserHome
}
