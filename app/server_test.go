package app_test

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/temporalio/cli/app"
	"github.com/urfave/cli/v2"
)

func (s *cliAppSuite) TestUIPortAlreadyInUse() {
	var (
		expectedServerPort = 7235
	)

	temporalCLI := app.BuildApp()
	// Don't call os.Exit
	temporalCLI.ExitErrHandler = func(_ *cli.Context, err error) {}

	unlockPorts := s.lockPorts(
		7233,
		7234,
		8233,
		8234,
	)
	defer unlockPorts()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	args, opts := newServerAndClientOpts(7233)
	opts.HostPort = fmt.Sprintf("localhost:%d", expectedServerPort)

	go func() {
		if err := temporalCLI.RunContext(ctx, args); err != nil {
			fmt.Println("Server closed with error:", err)
		}
	}()

	assertServerHealth(ctx, s.T(), opts)
}

func (s *cliAppSuite) lockPorts(ports ...uint) func() {
	var closeFunctions []func() error

	for _, port := range ports {
		address := fmt.Sprintf("127.0.0.1:%d", port)
		address6 := fmt.Sprintf("[::1]:%d", port)

		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			if addr, err = net.ResolveTCPAddr("tcp6", address6); err != nil {
				s.NoError(err, "failed to resolve tcp6 address")
			}

			s.NoError(err, "failed to resolve tcp address")
		}

		l, err := net.ListenTCP("tcp", addr)
		if err != nil {
			s.NoError(err, "failed to lock port")
		}

		closeFunctions = append(closeFunctions, l.Close)
	}

	return func() {
		for _, closeFn := range closeFunctions {
			err := closeFn()
			s.NoError(err, "failed to close locked port")
		}
	}
}
