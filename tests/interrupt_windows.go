//go:build windows

package tests

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

// sendInterrupt calls the break event on the given process for graceful shutdown.
func sendInterrupt(process *os.Process) error {
	dll, err := syscall.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	f, err := d.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r, _, err := p.Call(uintptr(syscall.CTRL_BREAK_EVENT), uintptr(process.Pid))
	if r == 0 {
		return err
	}
	return nil
}
