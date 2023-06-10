//go:build !windows

package tests

import (
	"os"
	"syscall"
)

// sendTerminate sends a terminate signal to the given process for graceful shutdown.
func sendTerminate(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}
