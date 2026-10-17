//go:build unix
// +build unix

package gotui

import (
	"os"
	"syscall"
	"unsafe"
)

// terminalWidth returns the terminal width in columns on Unix-like systems.
func terminalWidth() int {
	var dims [4]uint16
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL, os.Stdin.Fd(), uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(&dims)), 0, 0, 0,
	)
	if errno != 0 {
		return 0
	}
	return int(dims[1]) // columns
}
