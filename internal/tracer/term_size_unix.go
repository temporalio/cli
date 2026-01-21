//go:build !windows

package tracer

import (
	"syscall"
	"unsafe"
)

// getTerminalSize returns the current number of columns and rows in the active console window.
// The return value of this function is in the order of cols, rows.
// Copied from https://github.com/nathan-fiscaletti/consolesize-go/blob/master/consolesize_unix.go
func getTerminalSize() (width int, height int) {
	var size struct {
		rows    uint16
		cols    uint16
		xpixels uint16
		ypixels uint16
	}
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdout), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)))

	width = int(size.cols)
	height = int(size.rows)

	return
}
