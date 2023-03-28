package app_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"github.com/temporalio/cli/app"
	sconfig "github.com/temporalio/cli/server/config"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
	"testing"
)

type buildIdCompatSuite struct {
	suite.Suite
	app              *cli.App
	stopServerCancel context.CancelFunc
	client           client.Client
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
	portProvider.Close()
	ctx, cancel := context.WithCancel(context.Background())
	s.stopServerCancel = cancel

	args, clientOpts := newServerAndClientOpts(port)

	go func() {
		if err := s.app.RunContext(ctx, args); err != nil {
			fmt.Println("Server closed with error:", err)
		}
	}()

	s.client = assertServerHealth(s.T(), ctx, clientOpts)
}

func (s *buildIdCompatSuite) TearDownSuite() {
	s.stopServerCancel()
}

func (s *buildIdCompatSuite) testTqName() string {
	return "build-id-tq-" + s.T().Name()
}

func (s *buildIdCompatSuite) TestAddNewDefaultBuildIdAndGet() {
	err := s.app.Run([]string{"", "task-queue", "update-build-ids", "add-new-default",
		"--task-queue", s.testTqName(), "--namespace", "default",
		"--build-id", "foo"})
	s.Nil(err)
	err = s.app.Run([]string{"", "task-queue", "get-build-ids",
		"--task-queue", s.testTqName(), "--namespace", "default"})
	s.Nil(err)
}

func (s *buildIdCompatSuite) TestAddNewCompatBuildId() {
	err := s.app.Run([]string{"", "task-queue", "update-build-ids", "add-new-default",
		"--task-queue", s.testTqName(), "--namespace", "default",
		"--build-id", "foo"})
	s.Nil(err)
	err = s.app.Run([]string{"", "task-queue", "update-build-ids", "add-new-compatible",
		"--task-queue", s.testTqName(), "--namespace", "default",
		"--build-id", "bar", "--existing-compatible-build-id", "foo"})
	s.Nil(err)
}

func (s *buildIdCompatSuite) TestPromoteBuildIdSet() {
	err := s.app.Run([]string{"", "task-queue", "update-build-ids", "promote-set",
		"--task-queue", s.testTqName(), "--namespace", "default",
		"--build-id", "foo"})
	s.Nil(err)
}

func (s *buildIdCompatSuite) TestPromoteBuildIdInSet() {
	err := s.app.Run([]string{"", "task-queue", "update-build-ids", "promote-id-in-set",
		"--task-queue", s.testTqName(), "--namespace", "default",
		"--build-id", "foo"})
	s.Nil(err)
}
