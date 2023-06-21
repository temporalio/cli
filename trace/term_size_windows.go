//go:build windows

package trace

import (
	"syscall"
	"unsafe"
)

type (
	SHORT int16
	WORD  uint16

	COORD struct {
		X SHORT
		Y SHORT
	}

	SMALL_RECT struct {
		Left   SHORT
		Top    SHORT
		Right  SHORT
		Bottom SHORT
	}

	CONSOLE_SCREEN_BUFFER_INFO struct {
		Size              COORD
		CursorPosition    COORD
		Attributes        WORD
		Window            SMALL_RECT
		MaximumWindowSize COORD
	}
)

var (
	kernel32                       = syscall.NewLazyDLL("kernel32.dll")
	procGetConsoleScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

func getTerminalSize() (width int, height int) {
	var csbi CONSOLE_SCREEN_BUFFER_INFO

	_, _, err := procGetConsoleScreenBufferInfo.Call(uintptr(syscall.Stdout), uintptr(unsafe.Pointer(&csbi)))
	if err != syscall.Errno(0) {
		return 80, 25 // assume default terminal size
	}

	width = int(csbi.Size.X)
	height = int(csbi.Size.Y)

	return
}
