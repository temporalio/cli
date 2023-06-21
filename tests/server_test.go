package tests

import (
	"time"
)

func (s *e2eSuite) TestServerInterruptRC() {
	s.T().Parallel()

	cmd := newCmd(s.exePath, "server", "start-dev", "--headless")
	err := cmd.Start()
	s.NoError(err)

	// ensure the dev server process is killed in case of SIGTERM failure
	defer func() {
		_ = cmd.Process.Kill()
	}()

	// Wait for the server to start before sending signal
	time.Sleep(time.Second * 2)

	s.NoError(sendInterrupt(cmd.Process))

	err = cmd.Wait()
	s.NoError(err)

	s.Equal(0, cmd.ProcessState.ExitCode())
}
