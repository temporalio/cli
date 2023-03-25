package tests

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/app"
	"github.com/urfave/cli/v2"
	"go.temporal.io/api/operatorservice/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/server/temporaltest"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type (
	integrationSuite struct {
		suite.Suite
		app    *cli.App
		ts     *temporaltest.TestServer
		writer *MemWriter
	}
)

func TestClientIntegrationSuite(t *testing.T) {
	suite.Run(t, new(integrationSuite))
}

func (s *integrationSuite) SetupSuite() {
	s.app = app.BuildApp()
	s.ts = temporaltest.NewServer(temporaltest.WithT(s.T()))
}

func (s *integrationSuite) TearDownSuite() {
}

func (s *integrationSuite) SetupTest() {
	app.SetFactory(&clientFactory{
		frontendClient: nil,
		operatorClient: nil,
		sdkClient:      s.ts.GetDefaultClient(),
	})
	s.writer = &MemWriter{}
	s.app.Writer = s.writer
}

func (s *integrationSuite) TearDownTest() {
	s.ts.Stop()
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
