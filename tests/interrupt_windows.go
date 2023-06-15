//go:build windows

package tests

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

// sendInterrupt calls the break event on the given process for graceful shutdown.
func sendInterrupt(process *os.Process) error {
	dll, err := windows.LoadDLL("kernel32.dll")
	if err != nil {
		return err
	}
	defer dll.Release()
	f, err := dll.FindProc("AttachConsole")
	if err != nil {
		return err
	}
	r1, _, err := f.Call(uintptr(process.Pid))
	if r1 == 0 && err != syscall.ERROR_ACCESS_DENIED {
		return err
	}

	f, err = dll.FindProc("SetConsoleCtrlHandler")
	if err != nil {
		return err
	}
	r1, _, err = f.Call(0, 1)
	if r1 == 0 {
		return err
	}
	f, err = dll.FindProc("GenerateConsoleCtrlEvent")
	if err != nil {
		return err
	}
	r1, _, err = f.Call(windows.CTRL_BREAK_EVENT, uintptr(process.Pid))
	if r1 == 0 {
		return err
	}

	// Free the console after sending the interrupt
	f, err = dll.FindProc("FreeConsole")
	if err != nil {
		return err
	}
	r1, _, err = f.Call()
	if r1 == 0 {
		return err
	}

	return nil
}
