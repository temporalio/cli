package tests

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/temporalio/cli/common"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/app"
	"github.com/temporalio/cli/client"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	sdkclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"

	"go.temporal.io/sdk/worker"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type (
	e2eSuite struct {
		suite.Suite
		defaultWorkerOptions worker.Options
		exePath              string
		mutex                sync.Mutex
		servers              []*testsuite.DevServer
	}
)

func TestClientE2ESuite(t *testing.T) {
	suite.Run(t, new(e2eSuite))
}

func (s *e2eSuite) SetupSuite() {
	// build the temporal cli binary
	temp := os.TempDir()
	s.exePath = filepath.Join(temp, "temporal-cli-e2e")

	if runtime.GOOS == "windows" {
		s.exePath += ".exe"
	}
	build := exec.Command("go", "build", "-o", s.exePath, "../cmd/temporal")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	err := build.Run()
	require.NoError(s.T(), err)

	// noop exiter to prevent the app from exiting mid test
	cli.OsExiter = func(code int) {}
}

func (s *e2eSuite) TearDownSuite() {
	// ensure all servers are stopped in case not explicitly stopped from tests
	for _, server := range s.servers {
		_ = server.Stop()
	}
}

func (s *e2eSuite) SetupTest() {
}

func (s *e2eSuite) TearDownTest() {
}

func (s *e2eSuite) setUpTestEnvironment(extraArgs ...string) (*testsuite.DevServer, *cli.App, *common.MemWriter) {
	server, err := s.createServer(extraArgs...)
	s.Require().NoError(err)

	writer := &common.MemWriter{}
	tcli := s.createApp(server, writer)

	return server, tcli, writer
}

func (s *e2eSuite) createServer(extraArgs ...string) (*testsuite.DevServer, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ExistingPath: s.exePath,
		ExtraArgs: append([]string{
			// server logs are too noisy, limit server logs
			"--log-level", "error",
			//TODO: remove this flag when update workflow is enabled in the server by default
			"--dynamic-config-value", "frontend.enableUpdateWorkflowExecution=true",
		}, extraArgs...),
	})
	if err != nil {
		return nil, err
	}

	s.servers = append(s.servers, server)
	return server, err
}

func (s *e2eSuite) createApp(server *testsuite.DevServer, writer *common.MemWriter) *cli.App {
	tcli := app.BuildApp()
	tcli.Writer = writer

	client.SetFactory(tcli, &clientFactory{
		frontendClient: nil,
		operatorClient: nil,
		sdkClient:      server.Client(),
	})

	return tcli
}

func (s *e2eSuite) newWorker(server *testsuite.DevServer, taskQueue string, registerFunc func(registry worker.Registry)) worker.Worker {
	w := worker.New(server.Client(), taskQueue, s.defaultWorkerOptions)
	registerFunc(w)

	err := w.Start()
	s.NoError(err)

	return w
}

type clientFactory struct {
	frontendClient workflowservice.WorkflowServiceClient
	operatorClient operatorservice.OperatorServiceClient
	sdkClient      sdkclient.Client
}

func (m *clientFactory) FrontendClient(c *cli.Context) workflowservice.WorkflowServiceClient {
	return m.frontendClient
}

func (m *clientFactory) OperatorClient(c *cli.Context) operatorservice.OperatorServiceClient {
	return m.operatorClient
}

func (m *clientFactory) SDKClient(c *cli.Context, namespace string) sdkclient.Client {
	return m.sdkClient
}

func (m *clientFactory) HealthClient(_ *cli.Context) healthpb.HealthClient {
	panic("HealthClient mock is not supported")
}
