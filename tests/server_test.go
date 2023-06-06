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

	// build and run the binary, not "go run" as it modifies the exit code
	build := exec.Command("go", "build", "-o", exeName, "../cmd/temporal")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	err := build.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(exeName)

	cmd := exec.Command(exeName, "server", "start-dev", "--headless")
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}

	// ensure the dev server process is killed, even if SIGTERM fails
	defer cmd.Process.Signal(syscall.SIGKILL)

	// Wait for the app to start
	time.Sleep(time.Second * 2)

	cmd.Process.Signal(syscall.SIGTERM)
	_ = cmd.Wait()

	assert.Equal(t, 0, cmd.ProcessState.ExitCode())
}
