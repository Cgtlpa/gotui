//go:build windows
// +build windows

package gotui

import (
	"os"
	"strconv"
)

// terminalWidth returns the terminal width in columns on Windows.
// It checks the COLUMNS environment variable first, then falls back to 0.
func terminalWidth() int {
	if w := os.Getenv("COLUMNS"); w != "" {
		if n, err := strconv.Atoi(w); err == nil && n > 0 {
			return n
		}
	}
	return 0
}
