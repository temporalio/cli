//go:build !windows

package tests

import (
	"os"
	"syscall"
)

// sendInterrupt sends an interrupt signal to the given process for graceful shutdown.
func sendInterrupt(process *os.Process) error {
	return process.Signal(syscall.SIGINT)
}
