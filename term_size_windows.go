//go:build windows
// +build windows

package gotui

import (
	"os"
	"strconv"

	"golang.org/x/sys/windows"
)

// terminalWidth returns the terminal width in columns on Windows.
// It uses GetConsoleScreenBufferInfo via the Windows API, falling back to
// the COLUMNS environment variable, then 0.
func terminalWidth() int {
	// Try Windows API first.
	if h, err := windows.GetStdHandle(windows.STD_OUTPUT_HANDLE); err == nil && h != 0 {
		var csbi windows.ConsoleScreenBufferInfo
		if err := windows.GetConsoleScreenBufferInfo(h, &csbi); err == nil {
			return int(csbi.Size.X)
		}
	}

	// Fall back to COLUMNS env var.
	if w := os.Getenv("COLUMNS"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n > 0 {
			return n
		}
	}
	return 0
}
