package tests

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/app"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type (
	e2eSuite struct {
		suite.Suite
		app                  *cli.App
		ts                   *testsuite.DevServer
		workers              []worker.Worker
		defaultWorkerOptions worker.Options
		writer               *MemWriter
	}
)

func TestClientIntegrationSuite(t *testing.T) {
	suite.Run(t, new(e2eSuite))
}

func (s *e2eSuite) SetupSuite() {
	s.app = app.BuildApp()
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ExtraArgs: []string{
			// server logs are too noisy, limit server logs
			"--log-level", "error",
			//TODO: remove this flag when update workflow is enabled in the server by default
			"--dynamic-config-value", "frontend.enableUpdateWorkflowExecution=true",
		},
	})
	s.NoError(err)
	s.ts = server
}

func (s *e2eSuite) TearDownSuite() {
	err := s.ts.Stop()
	s.NoError(err)
}

func (s *e2eSuite) SetupTest() {
	app.SetFactory(&clientFactory{
		frontendClient: nil,
		operatorClient: nil,
		sdkClient:      s.ts.Client(),
	})
	s.writer = &MemWriter{}
	s.app.Writer = s.writer

	// noop exiter to prevent the app from exiting mid test
	cli.OsExiter = func(code int) { return }
}

func (s *e2eSuite) TearDownTest() {
}

func (s *e2eSuite) NewWorker(taskQueue string, registerFunc func(registry worker.Registry)) worker.Worker {
	w := worker.New(s.ts.Client(), taskQueue, s.defaultWorkerOptions)
	registerFunc(w)
	s.workers = append(s.workers, w)

	err := w.Start()
	s.NoError(err)

	return w
}

type clientFactory struct {
	frontendClient workflowservice.WorkflowServiceClient
	operatorClient operatorservice.OperatorServiceClient
	sdkClient      client.Client
}

func (m *clientFactory) FrontendClient(c *cli.Context) workflowservice.WorkflowServiceClient {
	return m.frontendClient
}

func (m *clientFactory) OperatorClient(c *cli.Context) operatorservice.OperatorServiceClient {
	return m.operatorClient
}

func (m *clientFactory) SDKClient(c *cli.Context, namespace string) client.Client {
	return m.sdkClient
}

func (m *clientFactory) HealthClient(_ *cli.Context) healthpb.HealthClient {
	panic("HealthClient mock is not supported")
}

// MemWriter is an io.Writer implementation that stores the written content.
type MemWriter struct {
	content bytes.Buffer
}

func (tlw *MemWriter) Write(p []byte) (n int, err error) {
	return tlw.content.Write(p)
}

func (tlw *MemWriter) GetContent() string {
	return tlw.content.String()
}

func TestMemWriter(t *testing.T) {
	tlw := &MemWriter{}
	fmt.Fprintln(tlw, "This message is written to the TestLogWriter.")

	content := tlw.GetContent()
	expected := "This message is written to the TestLogWriter."

	if !strings.Contains(content, expected) {
		t.Errorf("Expected log content to contain '%s', but it doesn't. Content: '%s'", expected, content)
	}
}
