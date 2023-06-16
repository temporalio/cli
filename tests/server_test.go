package tests

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerInterruptRC(t *testing.T) {
	exePath := "./test-temporal"

	// build and run the binary. Don't use "go run" as it modifies the exit code
	build := exec.Command("go", "build", "-o", exePath, "../cmd/temporal")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	err := build.Run()
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(exePath)
		assert.NoError(t, err)
	}()

	cmd := newCmd(exePath, "server", "start-dev", "--headless")
	err = cmd.Start()
	assert.NoError(t, err)

	// ensure the dev server process is killed, even if SIGTERM fails
	defer func() {
		_ = cmd.Process.Kill()
	}()

	// Wait for the server to start before sending signal
	time.Sleep(time.Second * 2)

	assert.NoError(t, sendInterrupt(cmd.Process))

	err = cmd.Wait()
	assert.NoError(t, err)

	assert.Equal(t, 0, cmd.ProcessState.ExitCode())
}
