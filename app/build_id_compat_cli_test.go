package app_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/app"
	sconfig "github.com/temporalio/cli/server/config"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
)

type buildIdCompatSuite struct {
	suite.Suite
	app              *cli.App
	stopServerCancel context.CancelFunc
	client           client.Client
	port             int
}

func TestBuildIdCompatSuite(t *testing.T) {
	suite.Run(t, new(buildIdCompatSuite))
}

func (s *buildIdCompatSuite) SetupSuite() {
	s.app = app.BuildApp()
	// Don't call os.Exit
	s.app.ExitErrHandler = func(_ *cli.Context, _ error) {}
	portProvider := sconfig.NewPortProvider()
	port := portProvider.MustGetFreePort()
	s.port = port
	portProvider.Close()
	ctx, cancel := context.WithCancel(context.Background())
	s.stopServerCancel = cancel

	args, clientOpts := newServerAndClientOpts(port)
	args = append(args,
		"--dynamic-config-value",
		"frontend.workerVersioningDataAPIs=true",
		"--dynamic-config-value",
		"frontend.workerVersioningWorkflowAPIs=true",
	)

	go func() {
		if err := s.app.RunContext(ctx, args); err != nil {
			fmt.Println("Server closed with error:", err)
		}
	}()

	s.client = assertServerHealth(ctx, s.T(), clientOpts)
}

func (s *buildIdCompatSuite) TearDownSuite() {
	s.stopServerCancel()
}

func (s *buildIdCompatSuite) testTqName() string {
	return "build-id-tq-" + s.T().Name()
}

func (s *buildIdCompatSuite) makeArgs(args ...string) []string {
	allArgs := []string{""}
	allArgs = append(allArgs, args...)
	return append(allArgs,
		"--address", fmt.Sprintf("localhost:%d", s.port),
		"--task-queue", s.testTqName(), "--namespace", "default")
}

func (s *buildIdCompatSuite) TestAddNewDefaultBuildIdAndGet() {
	err := s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "add-new-default", "--build-id", "foo"))
	s.Nil(err)
	err = s.app.Run(s.makeArgs("task-queue", "get-build-ids"))
	s.Nil(err)
}

func (s *buildIdCompatSuite) TestAddNewCompatBuildId() {
	err := s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "add-new-default", "--build-id", "foo"))
	s.Nil(err)
	err = s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "add-new-compatible",
		"--build-id", "bar", "--existing-compatible-build-id", "foo"))
	s.Nil(err)
}

func (s *buildIdCompatSuite) TestPromoteBuildIdSet() {
	err := s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "add-new-default", "--build-id", "foo"))
	s.Nil(err)
	err = s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "promote-set",
		"--build-id", "foo"))
	s.Nil(err)
}

func (s *buildIdCompatSuite) TestPromoteBuildIdInSet() {
	err := s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "add-new-default", "--build-id", "foo"))
	s.Nil(err)
	err = s.app.Run(s.makeArgs(
		"task-queue", "update-build-ids", "promote-id-in-set",
		"--build-id", "foo"))
	s.Nil(err)
}
