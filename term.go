package gotui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	ansiCursorHide      = "\x1b[?25l"
	ansiCursorShow      = "\x1b[?25h"
	ansiClearScreen     = "\x1b[2J"
	ansiCursorHome      = "\x1b[H"
	ansiReset           = "\x1b[0m"
	ansiBold            = "\x1b[1m"
	ansiReverse         = "\x1b[7m"
	ansiAltScreen       = "\x1b[?1049h"
	ansiExitAltScreen   = "\x1b[?1049l"
	ansiClearScrollback = "\x1b[3J"
)

// Terminal wraps stdin/stdout in raw mode and reads keypresses one at a time.
// Any leftover bytes from partial escape sequences are buffered internally so
// no keypress is ever skipped or duplicated.
type Terminal struct {
	in  *os.File
	out io.Writer
	old *syscall.Termios
	fd  uintptr
	buf []byte // leftover bytes from a partial escape sequence
}

// NewTerminal opens stdin/stdout and puts stdin into raw mode.
func NewTerminal() (*Terminal, error) {
	in := os.Stdin
	out := os.Stdout

	old, err := makeRaw(in.Fd())
	if err != nil {
		return nil, fmt.Errorf("failed to enter raw mode: %w", err)
	}

	return &Terminal{in: in, out: out, old: old, fd: in.Fd()}, nil
}

// Restore restores the original terminal state and shows the cursor again.
func (t *Terminal) Restore() error {
	if t.old == nil {
		return nil
	}
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL, t.fd, uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(t.old)), 0, 0, 0,
	)
	if errno != 0 {
		return errno
	}
	fmt.Fprint(t.out, ansiCursorShow+ansiExitAltScreen)
	return nil
}

// makeRaw puts the fd into raw mode: no echo, no line buffering, no signals.
func makeRaw(fd uintptr) (*syscall.Termios, error) {
	var old, new syscall.Termios

	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL, fd, uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&old)), 0, 0, 0,
	)
	if errno != 0 {
		return nil, errno
	}

	new = old
	new.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.ECHOE |
		syscall.ECHOK | syscall.ECHONL | syscall.ISIG | syscall.IEXTEN
	new.Iflag &^= syscall.IXON | syscall.IXOFF | syscall.IXANY |
		syscall.IGNCR | syscall.INLCR | syscall.ICRNL
	new.Oflag &^= syscall.OPOST
	new.Cc[syscall.VMIN] = 1
	new.Cc[syscall.VTIME] = 0

	_, _, errno = syscall.Syscall6(
		syscall.SYS_IOCTL, fd, uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(&new)), 0, 0, 0,
	)
	if errno != 0 {
		return nil, errno
	}

	return &old, nil
}

// Write sends bytes to the terminal output.
func (t *Terminal) Write(b []byte) (int, error) {
	return t.out.Write(b)
}

// WriteString sends a string to the terminal output.
func (t *Terminal) WriteString(s string) (int, error) {
	return t.out.Write([]byte(s))
}

// ReadKey reads exactly one keypress. It first drains any buffered bytes,
// then reads from stdin. If the first byte is ESC it waits briefly for the
// rest of an escape sequence (arrow keys, etc.).
func (t *Terminal) ReadKey() (Key, error) {
	// Drain buffered leftover bytes first.
	if len(t.buf) > 0 {
		key, rest := parseKey(t.buf)
		t.buf = rest
		return key, nil
	}

	first := make([]byte, 1)
	if _, err := t.in.Read(first); err != nil {
		return Key{}, err
	}

	b := first[0]
	if b != 0x1b {
		return decodeByte(b), nil
	}

	// ESC pressed: wait briefly for the rest of an escape sequence.
	t.in.SetReadDeadline(time.Now().Add(50 * time.Millisecond))
	rest := make([]byte, 16)
	n, _ := t.in.Read(rest)
	t.in.SetReadDeadline(time.Time{})

	if n == 0 {
		return Key{Type: KeyEsc}, nil
	}

	all := append([]byte{0x1b}, rest[:n]...)
	key, leftover := parseKey(all)
	t.buf = leftover
	return key, nil
}

// parseKey parses a complete byte slice that starts (optionally) with ESC and
// returns the decoded Key plus any leftover bytes that belong to a later key.
func parseKey(data []byte) (Key, []byte) {
	if len(data) == 0 {
		return Key{Type: KeyUnknown}, nil
	}

	if data[0] != 0x1b {
		return decodeByte(data[0]), data[1:]
	}

	if len(data) == 1 {
		return Key{Type: KeyEsc}, nil
	}

	// Strip the leading ESC and optional '[' (CSI).
	seq := string(data[1:])
	seq = strings.TrimPrefix(seq, "[")

	if len(seq) == 0 {
		return Key{Type: KeyEsc}, nil
	}

	switch seq[0] {
	case 'A':
		return Key{Type: KeyUp, Code: seq}, nil
	case 'B':
		return Key{Type: KeyDown, Code: seq}, nil
	case 'C':
		return Key{Type: KeyRight, Code: seq}, nil
	case 'D':
		return Key{Type: KeyLeft, Code: seq}, nil
	case 'H':
		return Key{Type: KeyHome, Code: seq}, nil
	case 'F':
		return Key{Type: KeyEnd, Code: seq}, nil
	case '~':
		if strings.HasPrefix(seq, "3") || strings.HasPrefix(seq, "4") {
			return Key{Type: KeyDelete, Code: seq}, nil
		}
		return Key{Type: KeyUnknown}, nil
	}

	return Key{Type: KeyUnknown}, nil
}

// decodeByte turns a single non-ESC byte into a Key.
func decodeByte(b byte) Key {
	switch b {
	case 0x03:
		return Key{Type: KeyCtrlC}
	case 0x04:
		return Key{Type: KeyCtrlD}
	case 0x0a, 0x0d:
		return Key{Type: KeyEnter}
	case 0x09:
		return Key{Type: KeyTab}
	case 0x7f, 0x08: // DEL and BS are both backspace
		return Key{Type: KeyBackspace}
	}

	if b < 0x20 {
		return Key{Type: KeyUnknown}
	}

	return Key{Type: KeyRune, Rune: rune(b)}
}

// runePrev returns the byte index of the rune immediately before idx.
func runePrev(s string, idx int) int {
	if idx <= 0 {
		return 0
	}
	if idx > len(s) {
		idx = len(s)
	}
	r := []rune(s[:idx])
	return len(string(r[:len(r)-1]))
}

// runeNext returns the byte index of the rune immediately after idx.
func runeNext(s string, idx int) int {
	if idx >= len(s) {
		return len(s)
	}
	r := []rune(s[idx:])
	return idx + len(string(r[:1]))
}

// delBefore removes the rune immediately before idx.
func delBefore(s string, idx int) string {
	p := runePrev(s, idx)
	return s[:p] + s[idx:]
}

// delAfter removes the rune at/after idx.
func delAfter(s string, idx int) string {
	n := runeNext(s, idx)
	return s[:idx] + s[n:]
}

// TerminalWidth returns the current terminal width in columns, or 0 if it
// cannot be determined. It uses the TIOCGWINSZ ioctl on stdin.
func TerminalWidth() int {
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
