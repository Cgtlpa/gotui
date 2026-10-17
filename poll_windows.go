//go:build windows

package gotui

import (
	"fmt"
	"time"

	"golang.org/x/sys/windows"
)

// waitReadable reports whether fd has data to read, waiting up to timeout.
// A negative timeout blocks until data arrives. Console and pipe handles are
// waitable on Windows, so WaitForSingleObject provides the same behavior as
// poll(2) on Unix.
func waitReadable(fd uintptr, timeout time.Duration) (bool, error) {
	var deadline time.Time
	if timeout >= 0 {
		deadline = time.Now().Add(timeout)
	}

	h := windows.Handle(fd)
	for {
		ms := uint32(windows.INFINITE)
		if timeout >= 0 {
			remaining := time.Until(deadline)
			if remaining <= 0 {
				return false, nil
			}
			ms = uint32(remaining / time.Millisecond)
			if ms == 0 {
				ms = 1
			}
		}

		ev, err := windows.WaitForSingleObject(h, ms)
		if err != nil {
			return false, err
		}
		switch ev {
		case windows.WAIT_OBJECT_0:
			var n uint32
			if err := windows.GetNumberOfConsoleInputEvents(h, &n); err != nil {
				// Not a console handle (pipe, file): readiness is enough.
				return true, nil
			}
			if n > 0 {
				return true, nil
			}
			// A console wake-up with no input events is spurious.
		case uint32(windows.WAIT_TIMEOUT):
			return false, nil
		default:
			return false, fmt.Errorf("waiting for input failed (status 0x%x)", ev)
		}
	}
}
