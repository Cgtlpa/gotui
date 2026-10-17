//go:build unix

package gotui

import (
	"time"

	"golang.org/x/sys/unix"
)

// waitReadable reports whether fd has data to read, waiting up to timeout.
// A negative timeout blocks until data arrives. Polling is what lets ReadKey
// tell a lone ESC apart from the start of an escape sequence without relying
// on read deadlines, which terminal files do not support.
func waitReadable(fd uintptr, timeout time.Duration) (bool, error) {
	ms := -1
	if timeout >= 0 {
		ms = int(timeout / time.Millisecond)
	}

	fds := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
	for {
		n, err := unix.Poll(fds, ms)
		if err == unix.EINTR {
			continue
		}
		if err != nil {
			return false, err
		}
		return n > 0, nil
	}
}
