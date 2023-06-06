package tests

import (
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServerInterruptRC(t *testing.T) {
	exeName := "./test-temporal"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}

	// build and run the binary. Don't use "go run" as it modifies the exit code
	build := exec.Command("go", "build", "-o", exeName, "../cmd/temporal")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	err := build.Run()
	assert.NoError(t, err)
	defer func() {
		err = os.Remove(exeName)
		assert.NoError(t, err)
	}()

	cmd := exec.Command(exeName, "server", "start-dev", "--headless")
	err = cmd.Start()
	assert.NoError(t, err)

	// ensure the dev server process is killed, even if SIGTERM fails
	defer func() {
		_ = cmd.Process.Signal(syscall.SIGKILL)
	}()

	// Wait for the app to start
	time.Sleep(time.Second * 2)

	err = cmd.Process.Signal(syscall.SIGTERM)
	assert.NoError(t, err)
	err = cmd.Wait()
	assert.NoError(t, err)

	assert.Equal(t, 0, cmd.ProcessState.ExitCode())
}
